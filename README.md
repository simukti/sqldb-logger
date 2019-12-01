# SQLDB-Logger

[![Build Status](https://travis-ci.com/simukti/sqldb-logger.svg)](https://travis-ci.com/simukti/sqldb-logger) [![Coverage Status](https://coveralls.io/repos/github/simukti/sqldb-logger/badge.svg)](https://coveralls.io/github/simukti/sqldb-logger) [![Go Report Card](https://goreportcard.com/badge/github.com/simukti/sqldb-logger)](https://goreportcard.com/report/github.com/simukti/sqldb-logger) [![GolangCI Status](https://golangci.com/badges/github.com/simukti/sqldb-logger.svg)](https://golangci.com/r/github.com/simukti/sqldb-logger) [![License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat)](https://raw.githubusercontent.com/simukti/sqldb-logger/master/LICENSE.txt)

A logger for Go SQL database driver without modify existing `*sql.DB` stdlib usage.

![shameless console output sample](./logadapter/zerologadapter/console.jpg?raw=true "go sql database logger output") 
_Colored console writer output above only for sample/development_

## FEATURES

- Leveled and [configurable](./options.go) logging.
- Keep `*sql.DB` as is (_from existing sql.DB usage perspective_).
- Bring your own logger backend using simple log interface.
- Trackable log output:
    - Every call has its own unique ID.
    - Prepared statement and execution will have same ID.
    - On execution/result error, it will include query, arguments, params, and related IDs. 

## INSTALL

```bash
go get -u -v github.com/simukti/sqldb-logger
```

_Version pinning using dependency manager such as [Mod](https://github.com/golang/go/wiki/Modules) or [Dep](https://github.com/golang/dep) is highly recommended._

## USAGE

### TL;DR VERSION

Replace `sql.Open()` with `sqldblogger.OpenDriver()`, both will return `*sql.DB`.

### DETAILED VERSION

Assuming we have existing `sql.Open` using commonly-used [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql) driver, 
and wants to log its `*sql.DB` interaction using [rs/zerolog](https://github.com/rs/zerolog).

```go
// import _ "github.com/go-sql-driver/mysql"
dsn := "username:passwd@tcp(mysqlserver:3306)/dbname?parseTime=true"
db, err := sql.Open("mysql", dsn) // db is *sql.DB
```

Change it with:

```go
// import "github.com/go-sql-driver/mysql"
loggerAdapter := zerologadapter.New(zerolog.New(os.Stdout)) // zerolog.New(zerolog.NewConsoleWriter()) // <-- for colored console
dsn := "username:passwd@tcp(mysqlserver:3306)/dbname?parseTime=true"
db, err := sqldblogger.OpenDriver(dsn, &mysql.MySQLDriver{}, loggerAdapter) // db is still *sql.DB
``` 

Without giving 4th argument to `OpenDriver`, it will use [default options](./options.go#L19-L29).

That's it. Use `db` object as usual.

### DETAILED + FULL OPTIONS VERSION

For full control of log output (field name, time format, etc...), pass variadic `sqldblogger.Option` as 4th argument as below:

```go
db, err := sqldblogger.OpenDriver(
    dsn, 
    &mysql.MySQLDriver{}, 
    loggerAdapter,
    // options
    sqldblogger.WithErrorFieldname("sql_error"), // default: error
    sqldblogger.WithDurationFieldname("query_duration"), // default: duration
    sqldblogger.WithTimeFieldname("log_time"), // default: time
    sqldblogger.WithSQLQueryFieldname("sql_query"), // default: query
    sqldblogger.WithSQLArgsFieldname("sql_args"), // default: args
    sqldblogger.WithMinimumLevel(sqldblogger.LevelInfo), // default: LevelDebug
    sqldblogger.WithLogArguments(false), // default: true
    sqldblogger.WithDurationUnit(sqldblogger.DurationNanosecond), // default: millisecond
    sqldblogger.WithTimeFormat(sqldblogger.TimeFormatRFC3339), // default: unix timestamp
    sqldblogger.WithLogDriverErrorSkip(true), // default: false
    sqldblogger.WithSQLQueryAsMessage(true), // default: false
)
```

### SQL DRIVER INTEGRATION

It is compatible with following public empty struct driver: 

#### MySQL ([go-sql-driver/mysql](https://github.com/go-sql-driver/mysql))

```go
db, err := sqldblogger.OpenDriver(dsn, &mysql.MySQLDriver{}, loggerAdapter /*, ...options */)
```

#### PostgreSQL ([lib/pq](https://github.com/lib/pq))

```go
db, err := sqldblogger.OpenDriver(dsn, &pq.Driver{}, loggerAdapter /*, ...options */) 
```

#### SQLite3 ([mattn/go-sqlite3](https://github.com/mattn/go-sqlite3))

```go
db, err := sqldblogger.OpenDriver(dsn, &sqlite3.SQLiteDriver{}, loggerAdapter /*, ...options */)
```

_Following drivers **maybe** compatible:_ 

#### SQL Server ([denisenkom/go-mssqldb](https://github.com/denisenkom/go-mssqldb))

```go
db, err := sqldblogger.OpenDriver(dsn, &mssql.Driver{}, loggerAdapter /*, ...options */)
```

#### Oracle ([mattn/go-oci8](https://github.com/mattn/go-oci8))

```go
db, err := sqldblogger.OpenDriver(dsn, oci8.OCI8Driver, loggerAdapter /*, ...options */)
```

## LOGGER ADAPTER

This library does not include a logger backend, but provide adapters that uses well known JSON structured logger:

- [Zerolog adapter](logadapter/zerologadapter): Using [rs/zerolog](https://github.com/rs/zerolog) as its logger.
- [Onelog adapter](logadapter/onelogadapter): Using [francoispqt/onelog](https://github.com/francoispqt/onelog) as its logger.
- [Zap adapter](logadapter/zapadapter): Using [uber-go/zap](https://github.com/uber-go/zap) as its logger.
- [Logrus adapter](logadapter/logrusadapter): Using [sirupsen/logrus](https://github.com/sirupsen/logrus) as its logger.

For another/custom logger backend, `Logger` interface is just as simple as following:

```go
type Logger interface {
	Log(ctx context.Context, level Level, msg string, data map[string]interface{})
}
``` 

## MOTIVATION

I want to:

- Stick to use `*sql.DB`.
- Have configurable output field.
- Leverage structured logging.
- Re-use [pgx log interface](https://github.com/jackc/pgx/blob/f3a3ee1a0e5c8fc8991928bcd06fdbcd1ee9d05c/logger.go#L46-L49).

I haven't found SQL logger with that features, so why not created myself? 

## REFERENCES

- [Stdlib sql.DB](https://github.com/golang/go/blob/master/src/database/sql/sql.go)
- [SQL driver interfaces](https://github.com/golang/go/blob/master/src/database/sql/driver/driver.go)
- [SQL driver implementation](https://github.com/golang/go/wiki/SQLDrivers)

## CONTRIBUTE

If you found bug, typo, wrong test, idea, help with existing issue, or anything constructive.
 
Don't hesitate to create an issue or pull request.

## CREDITS

- [pgx](https://github.com/jackc/pgx) for awesome PostgreSQL driver.

## LICENSE

[MIT](./LICENSE.txt)