// influxclean jobs package is responsible for launching queries and drops
//
// Author: Tesifonte Belda
// License: The MIT License (MIT)

package jobs

import (
	"github.com/tesibelda/influxclean/config"
	"github.com/tesibelda/influxclean/datastore/influxdb1"
	"github.com/tesibelda/influxclean/log"
)

var l *log.Logger

// RunJobs runs cleanup jobs defined in the provided configuration
func RunJobs(cfg *config.InfluxCleanConfig, lo *log.Logger, dryrun bool) error {
	var err error

	l = lo
	if cfg.Influxdb1 != nil {
		err = runInfluxdb1Jobs(cfg, dryrun)
	}
	// here more job and db types in the future (influxdb2, emptydbs,...)
	if err == nil {
		l.Info("Jobs completed")
	}
	return err
}

// runInfluxdb1Jobs runs all jobs of influxdb1 database type
func runInfluxdb1Jobs(cfg *config.InfluxCleanConfig, dryrun bool) error {
	var drywarn string
	var err, worsterr error
	var ic = &influxdb1.Influxdb1Client{Log: l}
	for _, inf := range cfg.Influxdb1 {
		switch dryrun {
		case true:
			drywarn = "with dry run enabled"
		default:
			drywarn = "with dry run DISABLED"
		}
		l.Infof("Connecting to influxdb1 at %s %s", inf.Url, drywarn)
		if err = ic.Open(inf.Url, inf.User, inf.Password, dryrun); err != nil {
			l.Errorf("Could not connect to influxdb1 %s: %v", inf.Url, err)
			worsterr = err
			continue
		}
		if err = runInfluxdb1OldSeries(ic, inf, dryrun); err != nil {
			worsterr = err
		}
		// here more job types in the future (emptydbs,...)
		ic.Close()
	}
	return worsterr
}
