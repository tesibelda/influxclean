# influxclean

influxclean is an influxdb cleanup utility that includes the following maintenance jobs:

* oldseries: drop series without specific data received for the last specified time

More maintenance jobs may be added in the future.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/tesibelda/influxclean/raw/master/LICENSE)

# Compatibility

Currently only access through Influxdbv1 API is used. 

# Configuration

* Edit influxclean.conf file as needed. Example:

```toml
[[influxdb1]]
  url = "http://localhost:8086"
  env_user = "INFLUX_USER"
  env_password = "INFLUX_PWD"
  user = ""
  password = ""
  # drop series from all measurements for Windows servers
  # without win_system data in telegraf db for three days
  [[influxdb1.oldseries]]
    name = "Windows servers"
    # if databases is empty all databases are checked
    databases = ["telegraf"]
    # retention policy, measurement and field to use in queries
    rp = "autogen"
    measurement = "win_system"
    field = "Processor_Queue_Length"
    # tags to detect old series (no more than two)
    tags = ["host"]
    # if drop_from_all is true series are dropped from all
    # measurements, otherwise (default) only from measurement
    drop_from_all = true
    # sleep period before jumping to the next database
    sleep_period = "0s"
    # relative time windows to query for data in db (units: s, m, h)
    # "0m", "0m" performs a search without time restriction
    history_window = ["0m", "0m"]
    current_window = ["72h", "1m"]
```

Environment variables specified with env_user and env_password take preference over user and password config entries.

* Run influxclean in dryrun mode first to check results first and then run it with dryun mode disabled.

# Quick test in your environment

* Edit influxclean.conf file as needed (see above)

* Run influxclean with --config argument using that file and dryrun mode enabled (default).
```
/path/to/influxclean --config /path/to/influxclean.conf
```

* Check the output and if you see the expected results you may launch the cleanup from db.
```
/path/to/influxclean --dryrun=false --config /path/to/influxclean.conf
```

# Example output

```plain
time="2023/03/17 15:44:26" level=info msg="Connecting to influxdb1 at http://localhost:8086 with dry run DISABLED"
time="2023/03/17 15:44:26" level=debug msg="Connected to http://localhost:8086 version 1.8.10"
time="2023/03/17 15:44:26" level=info msg="Old series job Windows servers..."
time="2023/03/17 15:44:26" level=info msg="Working on database telegraf"
time="2023/03/17 15:44:26" level=debug msg="querying: SHOW TAG VALUES FROM win_system WITH KEY=host"
time="2023/03/17 15:44:26" level=debug msg="querying: SELECT host FROM (SELECT first(Processor_Queue_Length), host::tag AS host FROM win_system WHERE (time > now() - 72h AND time < now() - 1m) GROUP BY host)"
time="2023/03/17 15:44:26" level=info msg="About to drop series for tag host with 2 values"
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
