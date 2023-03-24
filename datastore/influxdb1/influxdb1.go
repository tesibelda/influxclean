// influxclean influxdb1 package provides access to InfluxDB v1.x
//
// Author: Tesifonte Belda
// License: The MIT License (MIT)

package influxdb1

import (
	"fmt"
	"time"

	"github.com/influxdata/influxdb1-client/models"
	client "github.com/influxdata/influxdb1-client/v2"

	"github.com/tesibelda/influxclean/log"
)

type Influxdb1Client struct {
	con    client.Client
	Log    *log.Logger
	url    string
	user   string
	dryrun bool
}

var Separator = "#"

// Open opens a connection to the provided influxdb1
func (ic *Influxdb1Client) Open(url, user, password string, skip bool, dry bool) error {
	var err error
	var conf = client.HTTPConfig{
		Addr:               url,
		Username:           user,
		Password:           password,
		InsecureSkipVerify: skip,
	}

	ic.url = url
	ic.user = user
	ic.dryrun = dry
	if ic.con, err = client.NewHTTPClient(conf); err != nil {
		return err
	}

	timeout, _ := time.ParseDuration("3s")
	_, ver, err := ic.con.Ping(timeout)
	if err != nil {
		return err
	}
	ic.Log.Debugf("Connected to %s version %s", url, ver)
	return err
}

// Close closes the opened connection
func (ic *Influxdb1Client) Close() {
	ic.con.Close()
	ic.con = nil
}

// QueryShowDatabases returns the list of database names
func (ic *Influxdb1Client) QueryShowDatabases() ([]string, error) {
	var bogus models.Row
	var q client.Query
	var response *client.Response
	var query string
	var err error

	query = "SHOW DATABASES"
	q = client.NewQuery(query, "", "")
	if response, err = ic.con.Query(q); err != nil {
		return nil, err
	}
	if response.Error() != nil {
		return nil, fmt.Errorf("Query show databases failed: %s", response.Error())
	}
	if len(response.Results[0].Series) > 0 {
		bogus = response.Results[0].Series[0]
	}
	return rowShowSlice(bogus), err
}

// QueryShowTagValues returns all posible values for a tag in the index
func (ic *Influxdb1Client) QueryShowTagValues(db, rp, m, d1, f string) ([]string, error) {
	var bogus models.Row
	var q client.Query
	var response *client.Response
	var query string
	var err error

	// use Sprintf as client.NewQueryWithParameters does not work with all versions
	query = fmt.Sprintf("SHOW TAG VALUES FROM %s WITH KEY=%s", m, d1)
	if len(f) > 0 {
		query = fmt.Sprintf("%s WHERE %s", query, f)
	}
	q = client.NewQuery(query, db, "")
	q.RetentionPolicy = rp

	ic.Log.Debugf("querying: %s", q.Command)
	if response, err = ic.con.Query(q); err != nil {
		return nil, err
	}
	if response.Error() != nil {
		return nil, fmt.Errorf("Query show tag values failed: %s", response.Error())
	}
	if len(response.Results[0].Series) > 0 {
		bogus = response.Results[0].Series[0]
	}
	return rowShowSlice(bogus), err
}

// Query1Dim return the list of values for a tag with data in the given time window
func (ic *Influxdb1Client) Query1Dim(db, rp, m, p, d1, f, rb, re string) ([]string, error) {
	var bogus models.Row
	var q client.Query
	var response *client.Response
	var query string
	var err error

	wb, _ := time.ParseDuration(rb)
	we, _ := time.ParseDuration(re)
	if wb == we {
		cd, _ := time.ParseDuration("0m")
		if wb == cd {
			return ic.QueryShowTagValues(db, rp, m, d1, f)
		}
	}

	// use Sprintf as client.NewQueryWithParameters does not work with all versions
	query = fmt.Sprintf("SELECT %s FROM (SELECT first(%s), %s::tag AS %s FROM %s WHERE (time > now() - %s AND time < now() - %s)", d1, p, d1, d1, m, rb, re)
	if len(f) > 0 {
		query = fmt.Sprintf("%s AND %s", query, f)
	}
	query = fmt.Sprintf("%s GROUP BY %s)", query, d1)
	q = client.NewQuery(query, db, "")
	q.RetentionPolicy = rp

	ic.Log.Debugf("querying: %s", q.Command)
	if response, err = ic.con.Query(q); err != nil {
		return nil, err
	}
	if response.Error() != nil {
		return nil, fmt.Errorf("Query with dimension %s failed: %s", d1, response.Error())
	}
	if len(response.Results[0].Series) > 0 {
		bogus = response.Results[0].Series[0]
	}
	return rowSelectSlice(bogus), err
}

// Query2Dims return the list of values for the combination of two tags with data
// in the given time window
func (ic *Influxdb1Client) Query2Dims(db, rp, m, p, d1, d2, f, rb, re string) ([]string, error) {
	var (
		bogus        models.Row
		q            client.Query
		response     *client.Response
		query, where string
		err          error
	)

	query = fmt.Sprintf("SELECT %s, %s FROM (SELECT first(%s), %s::tag AS %s, %s::tag AS %s FROM %s", d1, d2, p, d1, d1, d2, d2, m)

	where = fmt.Sprintf("(time > now() - %s AND time < now() - %s)", rb, re)
	wb, _ := time.ParseDuration(rb)
	we, _ := time.ParseDuration(re)
	if wb == we {
		cd, _ := time.ParseDuration("0m")
		if wb == cd {
			where = ""
		}
	}
	switch len(f) {
	case 0:
		if len(where) > 0 {
			where = fmt.Sprintf("WHERE %s", where)
		}
	default:
		where = fmt.Sprintf("WHERE %s AND %s", where, f)
	}
	query = fmt.Sprintf("%s %s GROUP BY %s, %s)", query, where, d1, d2)

	q = client.NewQuery(query, db, "")
	q.RetentionPolicy = rp

	ic.Log.Debugf("querying: %s", q.Command)
	if response, err = ic.con.Query(q); err != nil {
		return nil, err
	}
	if response.Error() != nil {
		return nil, fmt.Errorf("Query with dimensions %s and %s failed: %s", d1, d2, response.Error())
	}
	if len(response.Results[0].Series) > 0 {
		bogus = response.Results[0].Series[0]
	}
	return rowSelectSlice(bogus), err
}

func (ic *Influxdb1Client) DropSeries1Dim(db, m, dim string, vals []string) error {
	var q client.Query
	var response *client.Response
	var query string
	var err error

	switch len(m) {
	case 0:
		query = "DROP SERIES WHERE"
	default:
		query = fmt.Sprintf("DROP SERIES FROM %s WHERE", m)
	}
	for i, val := range vals {
		if i > 0 {
			query = fmt.Sprintf("%s OR", query)
		}
		query = fmt.Sprintf("%s %s='%s'", query, dim, val)
	}
	q = client.NewQuery(query, db, "")

	ic.Log.Debugf("dropping: %s", q.Command)
	switch ic.dryrun {
	case false:
		response, err = ic.con.Query(q)
		if err == nil && response.Error() != nil {
			return fmt.Errorf("Dropping series failed: %s", response.Error())
		}
	case true:
		ic.Log.Debug("dryrun mode on, drops skipped")
	}
	return err
}

func (ic *Influxdb1Client) DropSeries2Dims(db, m, d1 string,
	vals1 []string,
	d2 string,
	vals2 []string,
) error {
	var q client.Query
	var response *client.Response
	var query string
	var err error

	if len(vals1) != len(vals2) {
		return fmt.Errorf("Received different size lists for the two tag values")
	}
	switch len(m) {
	case 0:
		query = "DROP SERIES WHERE"
	default:
		query = fmt.Sprintf("DROP SERIES FROM %s WHERE", m)
	}
	for i, val1 := range vals1 {
		if i > 0 {
			query = fmt.Sprintf("%s OR", query)
		}
		query = fmt.Sprintf("%s (%s='%s' AND %s='%s')", query, d1, val1, d2, vals2[i])
	}
	q = client.NewQuery(query, db, "")

	ic.Log.Debugf("dropping: %s", q.Command)
	switch ic.dryrun {
	case false:
		response, err = ic.con.Query(q)
		if err == nil && response.Error() != nil {
			return fmt.Errorf("Dropping series failed: %s", response.Error())
		}
	case true:
		ic.Log.Debug("dryrun mode on, drop skipped")
	}
	return err
}

func rowShowSlice(row models.Row) []string {
	var data []string
	var record, col string
	for _, point := range row.Values {
		for j, column := range row.Columns {
			col = string(column)
			if col == "value" || col == "name" {
				record = point[j].(string)
			}
		}
		data = append(data, record)
		record = ""
	}

	return data
}

func rowSelectSlice(row models.Row) []string {
	var data []string
	var record string
	for _, point := range row.Values {
		for j, column := range row.Columns {
			if string(column) != "time" && point[j] != nil {
				var actual = point[j].(string)
				switch len(record) {
				case 0:
					record = actual
				default:
					record = record + Separator + actual
				}
			}
		}
		data = append(data, record)
		record = ""
	}

	return data
}
