// influxclean config package provides access to configuration files
//
// Author: Tesifonte Belda
// License: The MIT License (MIT)

package config

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/pelletier/go-toml"
)

type InfluxCleanConfig struct {
	Name      string
	Influxdb1 []Influxdb1Info
}

type Influxdb1Info struct {
	Url                  string
	Env_user             string
	Env_password         string
	User                 string
	Password             string
	Insecure_skip_verify bool
	Oldseries            []OldSeriesInfo
}

type OldSeriesInfo struct {
	Name           string
	Databases      []string
	Rp             string
	Measurement    string
	Field          string
	Filter         string
	Tags           []string
	Drop_from_all  bool
	Sleep_duration string
	History_window []string
	Current_window []string
}

var ErrorString_ParseFailed = "Configuration parse failed"

func NewInfluxCleanConfig() *InfluxCleanConfig {
	var c = &InfluxCleanConfig{}
	return c
}

// ReadFile reads a config reader, expands env variables and parse config
func (c *InfluxCleanConfig) ReadFile(f io.Reader) error {
	var err error

	if err = toml.NewDecoder(f).Decode(c); err != nil {
		return err
	}
	c.defaultOldSeriesConfig()
	return c.parseConfig()
}

// defaultOldSeriesConfig sets default values if not provided
func (c *InfluxCleanConfig) defaultOldSeriesConfig() {
	for i := range c.Influxdb1 {
		for j := range c.Influxdb1[i].Oldseries {
			var job = &c.Influxdb1[i].Oldseries[j]
			job.Sleep_duration = defaultDuration(job.Sleep_duration)
			job.History_window = defaultWindowDuration(job.History_window)
			job.Current_window = defaultWindowDuration(job.Current_window)
		}
	}
}

// parseConfig parses an InfluxCleanConfig's contents
func (c *InfluxCleanConfig) parseConfig() error {
	var err error

	for i, inf := range c.Influxdb1 {
		if len(inf.Env_user) > 0 {
			c.Influxdb1[i].User = os.Getenv(inf.Env_user)
		}
		if len(inf.Env_password) > 0 {
			c.Influxdb1[i].Password = os.Getenv(inf.Env_password)
		}
		if err = parseOldSeriesConfig(inf, c); err != nil {
			return err
		}
	}
	return err
}

// parseOldSeriesConfig parses an OldSeries job config
func parseOldSeriesConfig(inf Influxdb1Info, c *InfluxCleanConfig) error {
	var err error
	for _, job := range inf.Oldseries {
		if len(job.Tags) == 0 || len(job.Tags) > 2 {
			return fmt.Errorf("%s. Only one or two tags clean jobs are possible",
				ErrorString_ParseFailed,
			)
		}
		if _, err = time.ParseDuration(job.Sleep_duration); err != nil {
			return fmt.Errorf("%s. Sleep_duration field could not be parsed: %v",
				ErrorString_ParseFailed,
				err,
			)
		}
		if err = parseWindow(job.History_window, "History"); err != nil {
			return err
		}
		if err = parseWindow(job.Current_window, "Current"); err != nil {
			return err
		}
	}
	return err
}

// parseWindow parses a relative time window config entry
func parseWindow(w []string, desc string) error {
	var t, tf time.Duration
	var err error
	if len(w) > 2 {
		return fmt.Errorf("%s. %s and current window should include two durations",
			ErrorString_ParseFailed,
			desc,
		)
	}
	for k := 0; k < len(w); k++ {
		if _, err = time.ParseDuration(w[k]); err != nil {
			return fmt.Errorf("%s. %s time window could not be parsed: %v",
				ErrorString_ParseFailed,
				desc,
				err,
			)
		}
		if k == 1 && t.Seconds() > tf.Seconds() {
			return fmt.Errorf("%s. %s relative times are not from older to newer",
				ErrorString_ParseFailed,
				desc,
			)
		}
		tf = t
	}
	return err
}

func defaultDuration(s string) string {
	if len(s) == 0 {
		s = "0s"
	}
	return s
}

func defaultWindowDuration(w []string) []string {
	var zw = []string{"0s", "0s"}
	for k := range w {
		w[k] = defaultDuration(w[k])
	}
	if len(w) == 0 {
		w = zw
	}
	return w
}
