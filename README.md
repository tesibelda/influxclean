# influxclean

influxclean is an [InfluxDB](https://github.com/influxdata/influxdb) cleanup utility that allows you to run the following job types:

* oldseries: drop series with no specific data received for the specified time. Useful when you want to drop specific not active series before the retention policy applies.

More job types may be added in the future.

**Warning**
Removal of data from the database cannot be undone. It is strongly suggested to take a backup before using this tool. Be cautius and double-check the results using dry run mode.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/tesibelda/influxclean/raw/master/LICENSE)

# Compatibility

Currently access to the database is only done using Influxdbv1 API. 

# Configuration

* Edit influxclean.conf file as needed. Example:

```toml
[[influxdb1]]
  url = "http://localhost:8086"
  env_user = "INFLUX_USER"
  env_password = "INFLUX_PWD"
  user = ""
  password = ""
  ## Use TLS but skip chain & host verification (default false)
  insecure_skip_verify = false
  # drop series from all measurements for Windows servers
  # with no win_system data in telegraf db for three days (72h)
  [[influxdb1.oldseries]]
    name = "Windows servers"
    # if databases is empty all databases are checked
    databases = ["telegraf"]
    # retention policy, measurement and field to use in queries
    rp = "autogen"
    measurement = "win_system"
    field = "Processor_Queue_Length"
	# additional filtering clause to use in queries (tag='value')
	filter = ""
    # tags to detect old series (no more than two)
    tags = ["host"]
    # if drop_from_all is true series are dropped from all
    # measurements, otherwise (default) only from measurement
    drop_from_all = true
    # sleep duration time before jumping to the next database
    sleep_duration = "0s"
    # time windows (relative to now) to query for data in db
    # "0m", "0m" performs a search without time restriction
    history_window = ["0m", "0m"]
    current_window = ["72h", "1m"]
```

Environment variables specified with env_user and env_password take preference over user and password config entries.

If databases list is empty (\[]) the job will be launched against all databases.

For oldseries job type, time windows are relative to the current time and are specified as duration (possible units: s, m, h). history_window is used to search historic series and current_window is used to search series with data currently received (from now-72h to now-1m in the example). Both queries take a list of tags values ("host" in the example), and the difference between them gives the series to drop. A filter can be added to work on more specific series using an expression like in [where clause](https://docs.influxdata.com/influxdb/v1.8/query_language/explore-schema/#show-tag-values) (usually tag='value').

More than one influxdb1 config entry can be specified to launch cleanup jobs to different influxdb servers. Also more than one job can be configured for each influxdb1 entry.

* Run influxclean in dry run mode first to check results first and then run it with dry run mode disabled to actually clean your database(s).

# Running in your environment

* Edit influxclean.conf file as needed (see above)

* Run influxclean with --config argument using that file and dry run mode enabled (default).
```
/path/to/influxclean --config /path/to/influxclean.conf
```
Debug mode is enabled by default to let you see the action that would be taken with dry run mode disabled.

* Check the output and if you see the expected results, you may launch the cleanup from your database(s). Anyway remember the warning above.
```
/path/to/influxclean --dryrun=false --config /path/to/influxclean.conf
```

You can disable debug logging by adding the flag --debug=false to the command.

# Example output

```plain
time="2023/03/17 15:44:26" level=info msg="Connecting to influxdb1 at http://localhost:8086 with dry run DISABLED"
time="2023/03/17 15:44:26" level=debug msg="Connected to http://localhost:8086 version 1.8.10"
time="2023/03/17 15:44:26" level=info msg="oldseries job Windows servers..."
time="2023/03/17 15:44:26" level=info msg="Working on database telegraf"
time="2023/03/17 15:44:26" level=debug msg="querying: SHOW TAG VALUES FROM win_system WITH KEY=host"
time="2023/03/17 15:44:26" level=debug msg="querying: SELECT host FROM (SELECT first(Processor_Queue_Length), host::tag AS host FROM win_system WHERE (time > now() - 72h AND time < now() - 1m) GROUP BY host)"
time="2023/03/17 15:44:26" level=info msg="About to drop series from telegraf db for tag host with 2 values"
time="2023/03/17 15:44:26" level=debug msg="dropping: DROP SERIES WHERE host='myawsserver01' OR host='myserver02'"
time="2023/03/17 15:44:27" level=info msg="Jobs completed"
```

# Build Instructions

Download the repo

    $ git clone git@github.com:tesibelda/influxclean.git

build the "influxclean" binary

    $ go build -o bin/influxclean cmd/main.go
    
 (if you're using windows, you'll want to give it an .exe extension)
 
    $ go build -o bin\influxclean.exe cmd/main.go

# Author

Tesifonte Belda (https://github.com/tesibelda)

# License

See [LICENSE](https://github.com/tesibelda/influxclean/blob/master/LICENSE)
