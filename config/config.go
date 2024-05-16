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
	URL                string
	EnvUser            string
	EnvPassword        string
	User               string
	Password           string
	InsecureSkipVerify bool
	TLSServerName      string `toml:"tls_server_name"`
	Oldseries          []OldSeriesInfo
}

type OldSeriesInfo struct {
	Name          string
	Databases     []string
	Rp            string
	Measurement   string
	Field         string
	Filter        string
	Tags          []string
	DropFromAll   bool
	SleepDuration string
	HistoryWindow []string
	CurrentWindow []string
}

var ErrorStringParseFailed = "Configuration parse failed"

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
			job.SleepDuration = defaultDuration(job.SleepDuration)
			job.HistoryWindow = defaultWindowDuration(job.HistoryWindow)
			job.CurrentWindow = defaultWindowDuration(job.CurrentWindow)
		}
	}
}

// parseConfig parses an InfluxCleanConfig's contents
func (c *InfluxCleanConfig) parseConfig() error {
	var err error

	for i, inf := range c.Influxdb1 {
		if len(inf.EnvUser) > 0 {
			c.Influxdb1[i].User = os.Getenv(inf.EnvUser)
		}
		if len(inf.EnvPassword) > 0 {
			c.Influxdb1[i].Password = os.Getenv(inf.EnvPassword)
		}
		if err = parseOldSeriesConfig(inf); err != nil {
			return err
		}
	}
	return err
}

// parseOldSeriesConfig parses an OldSeries job config
func parseOldSeriesConfig(inf Influxdb1Info) error {
	var err error
	for _, job := range inf.Oldseries {
		if len(job.Tags) == 0 || len(job.Tags) > 2 {
			return fmt.Errorf("%s. Only one or two tags clean jobs are possible",
				ErrorStringParseFailed,
			)
		}
		if _, err = time.ParseDuration(job.SleepDuration); err != nil {
			return fmt.Errorf("%s. SleepDuration field could not be parsed: %w",
				ErrorStringParseFailed, err)
		}
		if err = parseWindow(job.HistoryWindow, "History"); err != nil {
			return err
		}
		if err = parseWindow(job.CurrentWindow, "Current"); err != nil {
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
			ErrorStringParseFailed, desc)
	}
	for k := 0; k < len(w); k++ {
		if _, err = time.ParseDuration(w[k]); err != nil {
			return fmt.Errorf("%s. %s time window could not be parsed: %w",
				ErrorStringParseFailed, desc, err)
		}
		if k == 1 && t.Seconds() > tf.Seconds() {
			return fmt.Errorf("%s. %s relative times are not from older to newer",
				ErrorStringParseFailed, desc)
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
