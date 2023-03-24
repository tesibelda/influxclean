// influxclean jobs package is responsible for launching queries and drops
//
// Author: Tesifonte Belda
// License: The MIT License (MIT)

package jobs

import (
	"time"

	"github.com/tesibelda/influxclean/config"
	"github.com/tesibelda/influxclean/datastore/influxdb1"
	"github.com/tesibelda/influxclean/internal/sliceplus"
)

// runInfluxdb1OldSeries runs all oldseries jobs for influxdb1 databases
func runInfluxdb1OldSeries(
	ic *influxdb1.Influxdb1Client,
	inf config.Influxdb1Info,
	dryrun bool,
) error {
	var err, lasterr error
	for _, job := range inf.Oldseries {
		if len(job.Databases) == 0 {
			job.Databases, err = ic.QueryShowDatabases()
			if err != nil {
				l.Errorf("Error listing databases while runing oldseries job %s: %v",
					job.Name,
					err,
				)
			}
		}
		l.Infof("oldseries job %s...", job.Name)
		switch len(job.Tags) {
		case 1:
			err = runInfl1OldSeries1Dim(ic, job)
		case 2:
			err = runInfl1OldSeries2Dims(ic, job)
		}
		if err != nil {
			l.Errorf("Error runing oldseries job %s: %v", job.Name, err)
			lasterr = err
		}
	}
	return lasterr
}

// runInfl1OldSeries1Dim runs all oldseries job of influxdb1 type and one tag dimension
func runInfl1OldSeries1Dim(ic *influxdb1.Influxdb1Client, oc config.OldSeriesInfo) error {
	var (
		hdata, cdata, remdata []string
		tag, m                string
		hwb, hwe, cwb, cwe    string
		sl                    time.Duration
		err, lasterr          error
	)

	tag = oc.Tags[0]
	sl, _ = time.ParseDuration(oc.Sleep_duration)
	hwb = oc.History_window[0]
	hwe = oc.History_window[1]
	cwb = oc.Current_window[0]
	cwe = oc.Current_window[1]
	for i, db := range oc.Databases {
		if i > 0 {
			time.Sleep(sl)
		}
		l.Infof("Working on database %s", db)
		m = oc.Measurement
		hdata, err = ic.Query1Dim(db, oc.Rp, m, oc.Field, tag, oc.Filter, hwb, hwe)
		if err != nil {
			lasterr = err
			continue
		}
		if len(hdata) == 0 {
			l.Infof("No historic series found for oldseries job %s in %s db", oc.Name, db)
			continue
		}
		cdata, err = ic.Query1Dim(db, oc.Rp, m, oc.Field, tag, oc.Filter, cwb, cwe)
		if err != nil {
			lasterr = err
			continue
		}
		remdata = sliceplus.Difference(hdata, cdata)
		switch len(remdata) {
		case 0:
			l.Infof("No series where found to drop from %s db", db)
		default:
			var about = "About to drop series from"
			switch oc.Drop_from_all {
			case true:
				l.Infof("%s %s db for tag %s with %d values",
					about,
					db,
					tag,
					len(remdata),
				)
			default:
				l.Infof("%s measurement %s in %s db for tag %s with %d values",
					about,
					m,
					db,
					tag,
					len(remdata),
				)
			}
		}
		for _, ch := range sliceplus.ChunkSlice(remdata, 60) {
			if oc.Drop_from_all {
				m = ""
			}
			if err = ic.DropSeries1Dim(db, m, tag, ch); err != nil {
				lasterr = err
				continue
			}
		}
	}
	return lasterr
}

// runInfl1OldSeries2Dims runs all oldseries job of influxdb1 type and two tag dimensions
func runInfl1OldSeries2Dims(ic *influxdb1.Influxdb1Client, oc config.OldSeriesInfo) error {
	var (
		hdata, cdata, remdata []string
		vals1, vals2          []string
		tag1, tag2, m         string
		hwb, hwe, cwb, cwe    string
		sl                    time.Duration
		err, lasterr          error
	)

	tag1 = oc.Tags[0]
	tag2 = oc.Tags[1]
	sl, _ = time.ParseDuration(oc.Sleep_duration)
	hwb = oc.History_window[0]
	hwe = oc.History_window[1]
	cwb = oc.Current_window[0]
	cwe = oc.Current_window[1]
	for i, db := range oc.Databases {
		if i > 0 {
			time.Sleep(sl)
		}
		l.Infof("Working on database %s", db)
		m = oc.Measurement
		hdata, err = ic.Query2Dims(db, oc.Rp, m, oc.Field, tag1, tag2, oc.Filter, hwb, hwe)
		if err != nil {
			lasterr = err
			continue
		}
		if len(hdata) == 0 {
			l.Infof("No historic series found for oldseries job %s in %s db", oc.Name, db)
			continue
		}
		cdata, err = ic.Query2Dims(db, oc.Rp, m, oc.Field, tag1, tag2, oc.Filter, cwb, cwe)
		if err != nil {
			lasterr = err
			continue
		}
		remdata = sliceplus.Difference(hdata, cdata)
		switch len(remdata) {
		case 0:
			l.Infof("No series where found to drop from %s db", db)
		default:
			var about = "About to drop series from"
			switch oc.Drop_from_all {
			case true:
				l.Infof("%s %s db for tags %s and %s with %d values",
					about,
					db,
					tag1,
					tag2,
					len(remdata),
				)
			default:
				l.Infof("%s measurement %s in %s db for tags %s and %s with %d values",
					about,
					m,
					db,
					tag1,
					tag2,
					len(remdata),
				)
			}
		}

		for _, ch := range sliceplus.ChunkSlice(remdata, 40) {
			if oc.Drop_from_all {
				m = ""
			}
			vals1, vals2 = sliceplus.Split2Dims(ch, influxdb1.Separator)
			if err = ic.DropSeries2Dims(db, m, tag1, vals1, tag2, vals2); err != nil {
				lasterr = err
				continue
			}
		}
	}
	return lasterr
}
