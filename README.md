# xhplconsole-api
REST API for xhplconsole DB logging

## About

This API is a part of [deployxhpl](https://github.com/JustAddRobots/deployxhpl), a
CI/CD proof-of-concept for running XHPL on baremetal computes. It logs XHPL test
output to a local SQL database (MariaDB) via serialised JSON. See 
[runxhpl](https://github.com/JustAddRobots/runxhpl) for corresponding the API client.

This is very much a work-in-progess. For XHPL tests, automatically modifying or 
deleting a record is not the best idea (at the institutional level, at least). So 
currently, this API is a barebones implemenation including GET and POST requests. 
Time permitting, the whole CRUD (and use cases) will be implemented.

This project is **not supported**.

## Features

* GET/POST XHPL test runs to SQL DB

## Installing

```
❯ go get github.com/JustAddRobots/xhplconsole-api
❯ cd github.com/JustAddRobots/xhplconsole-api
❯ go build
❯ ./xhplconsole-api
```
The default endpoint is *locahost:3456/v1/machines*


This POC has some specific environment requirements:

It requires access to a local DB (schema below). It also uses a local INI config file
to resolve the DB password.

```
MariaDB [xhplconsole]> SHOW COLUMNS FROM xhpltest;
+---------------------------+-------------+------+-----+---------------------+----------------+
| Field                     | Type        | Null | Key | Default             | Extra          |
+---------------------------+-------------+------+-----+---------------------+----------------+
| id                        | int(6)      | NO   | PRI | NULL                | auto_increment |
| uuid                      | char(36)    | YES  |     | NULL                |                |
| serial_num                | varchar(64) | YES  |     | NULL                |                |
| log_id                    | varchar(40) | YES  |     | NULL                |                |
| cpu_vendor                | varchar(32) | YES  |     | NULL                |                |
| cpu_family_model_stepping | text        | YES  |     | NULL                |                |
| cpu_core_count            | int(4)      | YES  |     | NULL                |                |
| cpu_flags                 | text        | YES  |     | NULL                |                |
| lscpu                     | longtext    | YES  |     | NULL                |                |
| cpuinfo                   | longtext    | YES  |     | NULL                |                |
| dmidecode                 | longtext    | YES  |     | NULL                |                |
| meminfo                   | longtext    | YES  |     | NULL                |                |
| test_name                 | varchar(32) | YES  |     | NULL                |                |
| test_cmd                  | text        | YES  |     | NULL                |                |
| time_start                | datetime    | YES  |     | NULL                |                |
| time_end                  | datetime    | YES  |     | NULL                |                |
| test_params               | longtext    | YES  |     | NULL                |                |
| test_metric               | varchar(32) | YES  |     | NULL                |                |
| test_status               | varchar(16) | YES  |     | NULL                |                |
| test_log                  | mediumtext  | YES  |     | NULL                |                |
| log_time                  | timestamp   | NO   |     | current_timestamp() |                |
+---------------------------+-------------+------+-----+---------------------+----------------+

```

## Usage

```
❯ cd github.com/JustAddRobots/xhplconsole-api
❯ ./xhplconsole-api
```

Example from Python interpreter, logs generated from *runxhpl*.

```
❯ python3
>>> import requests
>>> import json
>>> with open("/tmp/xhpllogs.json", "r") as f:
...     s = f.read()
>>> m = json.loads(s)
>>> r = requests.post("http://hosaka.local:3456/v1/machines", json=m)
```
```
>>> r = requests.get("http://hosaka.local:3456/v1/machines/89")
>>> r.status_code
200
>>> m = r.json()
>>> m["test_cmd"]
'mpirun --allow-run-as-root -mca btl_vader_single_copy_mechanism none -np 4 xhpl-x86_64'
```

## Todo

* Add UPDATE/DELETE and use cases
* Add collections for DIMMs and CPUs
* Add collections for arch/CPU/SIMD
* Add collections for test params

## License

Licensed under GNU GPL v3. See **LICENSE.md**.
