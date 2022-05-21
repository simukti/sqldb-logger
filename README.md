# SQLDB-Logger

[![Coverage Status](https://coveralls.io/repos/github/simukti/sqldb-logger/badge.svg)](https://coveralls.io/github/simukti/sqldb-logger) [![Go Report Card](https://goreportcard.com/badge/github.com/simukti/sqldb-logger)](https://goreportcard.com/report/github.com/simukti/sqldb-logger) [![Sonar Violations (long format)](https://img.shields.io/sonar/violations/simukti_sqldb-logger?server=https%3A%2F%2Fsonarcloud.io)](https://sonarcloud.io/dashboard?id=simukti_sqldb-logger) [![Sonar Tech Debt](https://img.shields.io/sonar/tech_debt/simukti_sqldb-logger?server=https%3A%2F%2Fsonarcloud.io)](https://sonarcloud.io/dashboard?id=simukti_sqldb-logger) [![Sonar Quality Gate](https://img.shields.io/sonar/quality_gate/simukti_sqldb-logger?server=https%3A%2F%2Fsonarcloud.io)](https://sonarcloud.io/dashboard?id=simukti_sqldb-logger) [![Documentation](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/simukti/sqldb-logger) [![License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat)](https://raw.githubusercontent.com/simukti/sqldb-logger/master/LICENSE.txt)

A logger for Go SQL database driver without modify existing `*sql.DB` stdlib usage.

![shameless console output sample](./logadapter/zerologadapter/console.jpg?raw=true "go sql database logger output") 
_Colored console writer output above only for sample/development_

## FEATURES

- Leveled, detailed and [configurable](./options.go) logging.
- Keep using (or re-use existing) `*sql.DB` as is.
- Bring your own logger backend via simple log interface.
- Trackable log output:
    - Every call has its own unique ID.
    - Prepared statement and execution will have same ID.
    - On execution/result error, it will include the query, arguments, params, and related IDs. 

## INSTALL

```bash
go get -u -v github.com/simukti/sqldb-logger
```

_Version pinning using dependency manager such as [Mod](https://github.com/golang/go/wiki/Modules) or [Dep](https://github.com/golang/dep) is highly recommended._

## USAGE

As a start, `Logger` is just a simple interface:

```go
type Logger interface {
	Log(ctx context.Context, level Level, msg string, data map[string]interface{})
}
``` 

There are 4 included basic implementation that uses well-known JSON structured logger for quickstart:

- [Zerolog adapter](logadapter/zerologadapter): Using [rs/zerolog](https://github.com/rs/zerolog) as its logger.
- [Onelog adapter](logadapter/onelogadapter): Using [francoispqt/onelog](https://github.com/francoispqt/onelog) as its logger.
- [Zap adapter](logadapter/zapadapter): Using [uber-go/zap](https://github.com/uber-go/zap) as its logger.
- [Logrus adapter](logadapter/logrusadapter): Using [sirupsen/logrus](https://github.com/sirupsen/logrus) as its logger.

_Note: [those adapters](./logadapter) does not use given `context`, you need to modify it and adjust with your needs._ 
_(example: add http request id/whatever value from context to query log when you call `QueryerContext` and`ExecerContext` methods)_

Then for that logger to works, you need to integrate with a compatible driver which will be used by `*sql.DB`.

### INTEGRATE WITH EXISTING SQL DB DRIVER
 
Re-use from existing `*sql.DB` driver, this is the simplest way:

For example, from:

```go
dsn := "username:passwd@tcp(mysqlserver:3306)/dbname?parseTime=true"
db, err := sql.Open("mysql", dsn) // db is *sql.DB
db.Ping() // to check connectivity and DSN correctness
```

To:

```go
// import sqldblogger "github.com/simukti/sqldb-logger"
// import "github.com/simukti/sqldb-logger/logadapter/zerologadapter"
dsn := "username:passwd@tcp(mysqlserver:3306)/dbname?parseTime=true"
db, err := sql.Open("mysql", dsn) // db is *sql.DB
// handle err
loggerAdapter := zerologadapter.New(zerolog.New(os.Stdout))
db = sqldblogger.OpenDriver(dsn, db.Driver(), loggerAdapter/*, using_default_options*/) // db is STILL *sql.DB
db.Ping() // to check connectivity and DSN correctness
```

That's it, all `*sql.DB` interaction now logged.

### INTEGRATE WITH SQL DRIVER STRUCT

It is also possible to integrate with following public empty struct driver directly: 

#### MySQL ([go-sql-driver/mysql](https://github.com/go-sql-driver/mysql))

```go
db := sqldblogger.OpenDriver(dsn, &mysql.MySQLDriver{}, loggerAdapter /*, ...options */)
```

#### PostgreSQL ([lib/pq](https://github.com/lib/pq))

```go
db := sqldblogger.OpenDriver(dsn, &pq.Driver{}, loggerAdapter /*, ...options */) 
```

#### SQLite3 ([mattn/go-sqlite3](https://github.com/mattn/go-sqlite3))

```go
db := sqldblogger.OpenDriver(dsn, &sqlite3.SQLiteDriver{}, loggerAdapter /*, ...options */)
```

_Following struct drivers **maybe** compatible:_ 

#### SQL Server ([denisenkom/go-mssqldb](https://github.com/denisenkom/go-mssqldb))

```go
db := sqldblogger.OpenDriver(dsn, &mssql.Driver{}, loggerAdapter /*, ...options */)
```

#### Oracle ([mattn/go-oci8](https://github.com/mattn/go-oci8))

```go
db := sqldblogger.OpenDriver(dsn, oci8.OCI8Driver, loggerAdapter /*, ...options */)
```

## LOGGER OPTIONS

When using `sqldblogger.OpenDriver(dsn, driver, logger, opt...)` without 4th variadic argument, it will use [default options](./options.go#L37-L59).

Here is sample of `OpenDriver()` using all available options and use non-default value:

```go
db = sqldblogger.OpenDriver(
    dsn, 
    db.Driver(), 
    loggerAdapter,
    // AVAILABLE OPTIONS
    sqldblogger.WithErrorFieldname("sql_error"),                    // default: error
    sqldblogger.WithDurationFieldname("query_duration"),            // default: duration
    sqldblogger.WithTimeFieldname("log_time"),                      // default: time
    sqldblogger.WithSQLQueryFieldname("sql_query"),                 // default: query
    sqldblogger.WithSQLArgsFieldname("sql_args"),                   // default: args
    sqldblogger.WithMinimumLevel(sqldblogger.LevelTrace),           // default: LevelDebug
    sqldblogger.WithLogArguments(false),                            // default: true
    sqldblogger.WithDurationUnit(sqldblogger.DurationNanosecond),   // default: DurationMillisecond
    sqldblogger.WithTimeFormat(sqldblogger.TimeFormatRFC3339),      // default: TimeFormatUnix
    sqldblogger.WithLogDriverErrorSkip(true),                       // default: false
    sqldblogger.WithSQLQueryAsMessage(true),                        // default: false
    sqldblogger.WithUIDGenerator(sqldblogger.UIDGenerator),         // default: *defaultUID
    sqldblogger.WithConnectionIDFieldname("con_id"),                // default: conn_id
    sqldblogger.WithStatementIDFieldname("stm_id"),                 // default: stmt_id
    sqldblogger.WithTransactionIDFieldname("trx_id"),               // default: tx_id
    sqldblogger.WithWrapResult(false),                              // default: true
    sqldblogger.WithIncludeStartTime(true),                         // default: false
    sqldblogger.WithStartTimeFieldname("start_time"),               // default: start
    sqldblogger.WithPreparerLevel(sqldblogger.LevelDebug),          // default: LevelInfo
    sqldblogger.WithQueryerLevel(sqldblogger.LevelDebug),           // default: LevelInfo
    sqldblogger.WithExecerLevel(sqldblogger.LevelDebug),            // default: LevelInfo
)
```

[Click here](https://pkg.go.dev/github.com/simukti/sqldb-logger#Option) for options documentation.

## MOTIVATION

I want to:

- Keep using `*sql.DB`.
- Have configurable output field.
- Leverage structured logging.
- Fetch and log `context.Context` value if needed. 
- Re-use [pgx log interface](https://github.com/jackc/pgx/blob/f3a3ee1a0e5c8fc8991928bcd06fdbcd1ee9d05c/logger.go#L46-L49).

I haven't found Go `*sql.DB` logger with that features, so why not created myself? 

## REFERENCES

- [Stdlib sql.DB](https://github.com/golang/go/blob/master/src/database/sql/sql.go)
- [SQL driver interfaces](https://github.com/golang/go/blob/master/src/database/sql/driver/driver.go)
- [SQL driver implementation](https://github.com/golang/go/wiki/SQLDrivers)

## CONTRIBUTE

If you found a bug, typo, wrong test, idea, help with existing issue, or anything constructive.
 
Don't hesitate to create an issue or pull request.

## CREDITS

- [pgx](https://github.com/jackc/pgx) for awesome PostgreSQL driver.

## LICENSE

[MIT](./LICENSE.txt)