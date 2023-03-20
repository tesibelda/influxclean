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
	Url          string
	Env_user     string
	Env_password string
	User         string
	Password     string
	Oldseries    []OldSeriesInfo
}

type OldSeriesInfo struct {
	Name           string
	Databases      []string
	Rp             string
	Measurement    string
	Field          string
	Tags           []string
	Drop_from_all  bool
	Sleep_period   string
	History_window []string
	Current_window []string
}

// ReadFile reads a config reader, expands env variables and parse config
func (c *InfluxCleanConfig) ReadFile(f io.Reader) error {
	var err error

	if err = toml.NewDecoder(f).Decode(c); err != nil {
		return err
	}
	return parseConfig(c)
}

// parseConfig parses an InfluxCleanConfig's contents
func parseConfig(c *InfluxCleanConfig) error {
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
			return fmt.Errorf("Only one or two tags clean jobs are possible")
		}
		if _, err = time.ParseDuration(job.Sleep_period); err != nil {
			return err
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
		return fmt.Errorf("%s and current window should include two durations", desc)
	}
	for k := 0; k < len(w); k++ {
		if _, err = time.ParseDuration(w[k]); err != nil {
			return err
		}
		if k == 1 && t.Seconds() > tf.Seconds() {
			return fmt.Errorf("%s relative times are not from older to newer", desc)
		}
		tf = t
	}
	return err
}
