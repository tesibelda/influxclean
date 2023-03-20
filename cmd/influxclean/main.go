// influxclean main package provides cli and starts DB cleaning jobs
//
// Author: Tesifonte Belda
// License: The MIT License (MIT)

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/tesibelda/influxclean/config"
	"github.com/tesibelda/influxclean/jobs"
	"github.com/tesibelda/influxclean/log"
)

var Version string = ""

func main() {
	var (
		cfgfile string
		f       *os.File
		ecode   int
		err     error
		dryrun  bool
		debug   bool
	)

	// parameters
	flag.BoolVar(&debug, "debug", true, "display queries and results")
	flag.BoolVar(&dryrun, "dryrun", true, "dry run does not drop any series")
	flag.StringVar(&cfgfile, "config", "influxclean.toml", "config file")
	var showVersion = flag.Bool("version", false, "show version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Println("influxclean", Version)
		os.Exit(0)
	}

	// config file
	if f, err = os.Open(cfgfile); err != nil {
		fmt.Println("Error opening configuration file:", err)
		os.Exit(1)
	}
	var cfg = &config.InfluxCleanConfig{}
	if err = cfg.ReadFile(f); err != nil {
		fmt.Fprintf(os.Stderr, "Could not load config in file %s: %s", cfgfile, err.Error())
		os.Exit(1)
	}
	f.Close()

	// run cleanup jobs
	var l = log.NewLogger(debug)
	if err = jobs.RunJobs(cfg, l, dryrun); err != nil {
		ecode = 2
	}
	os.Exit(ecode)
}
