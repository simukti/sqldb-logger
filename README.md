# sqldb-logger - Go SQL Database Logger

[![Build Status](https://travis-ci.com/simukti/sqldb-logger.svg)](https://travis-ci.com/simukti/sqldb-logger) [![Coverage Status](https://coveralls.io/repos/github/simukti/sqldb-logger/badge.svg)](https://coveralls.io/github/simukti/sqldb-logger) [![Go Report Card](https://goreportcard.com/badge/github.com/simukti/sqldb-logger)](https://goreportcard.com/report/github.com/simukti/sqldb-logger) [![GolangCI Status](https://golangci.com/badges/github.com/simukti/sqldb-logger.svg)](https://golangci.com/r/github.com/simukti/sqldb-logger) [![License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat)](https://raw.githubusercontent.com/simukti/sqldb-logger/master/LICENSE.txt)

A simple wrapper for Go SQL database driver to log its interaction from `sql.DB`.

![shameless console output sample](./logadapter/zerologadapter/console.jpg?raw=true "go sql database logger output")

## INSTALL

```bash
go get -u -v github.com/simukti/sqldb-logger
```

## USAGE

Assuming we have existing `sql.Open` using commonly-used go-sql-driver/mysql driver and want to log its interaction with `sql.DB` using Zerolog.

```go
// import _ "github.com/go-sql-driver/mysql"
dsn := "username:passwd@tcp(mysqlserver:3306)/dbname?parseTime=true"
db, err := sql.Open("mysql", dsn) // db is *sql.DB
```

Change it with:

```go
// import "github.com/go-sql-driver/mysql"
zlogger := zerolog.New(os.Stdout) // zerolog.New(zerolog.NewConsoleWriter()) // <-- for colored console
dsn := "username:passwd@tcp(mysqlserver:3306)/dbname?parseTime=true"
db, err := sqldblogger.OpenDriver(dsn, &mysql.MySQLDriver{}, zlogger) // db is *sql.DB
``` 

Without giving 4th argument to `OpenDriver`, it will automatically set [default options](./options.go#L19-L29).

## OPTIONS

For full control of log output (field name, time format, etc...), pass variadic `sqldblogger.Option` as 4th argument as below:

```go
db, err := sqldblogger.OpenDriver(
    dsn, 
    &mysql.MySQLDriver{}, 
    zlogger,
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
)
```

That's it. Use `db` object as usual.

## COMPATIBLITY

It should be compatible with following empty public struct driver: 

- [go-sql-driver/mysql](https://github.com/go-sql-driver/mysql/blob/15462c1d60d42ecca11d6ef9fec0b0afd5833459/driver.go#L84)
- [lib/pq](https://github.com/lib/pq/blob/f91d3411e481ed313eeab65ebfe9076466c39d01/conn.go#L52)
- [mattn/go-sqlite3](https://github.com/mattn/go-sqlite3/blob/590d44c02bca83987d23f6eab75e6d0ddf95f644/sqlite3.go#L230)
- [denisenkom/go-mssqldb](https://github.com/denisenkom/go-mssqldb/blob/cfbb681360f0a7de54ae77703318f0e60d422e00/mssql.go#L33)

## PROVIDED LOGGER ADAPTER

There are 3 adapters within this repo:

- [Zerolog adapter](logadapter/zerologadapter): Using [rs/zerolog](https://github.com/rs/zerolog) as its logger.

![zerolog output sample](./logadapter/zerologadapter/zerolog.jpg?raw=true "go sql database logger output")

- [Onelog adapter](logadapter/onelogadapter): Using [francoispqt/onelog](https://github.com/francoispqt/onelog) as its logger.

![onelog output sample](./logadapter/onelogadapter/onelog.jpg?raw=true "go sql database logger output")

- [Zap adapter](logadapter/zapadapter): Using [uber-go/zap](https://github.com/uber-go/zap) as its logger.

![zap output sample](./logadapter/zapadapter/zap.jpg?raw=true "go sql database logger output")

Implements another logger must follow these simple interface:

```go
type Logger interface {
	Log(ctx context.Context, level Level, msg string, data map[string]interface{})
}
``` 

## MOTIVATION

I want to:

- Stick to use `sql.DB`.
- Level-logging SQL database interaction.
- Re-use [pgx log interface](https://github.com/jackc/pgx/blob/f3a3ee1a0e5c8fc8991928bcd06fdbcd1ee9d05c/logger.go#L46-L49) for commonly use SQL driver.
- Have configurable output field.
- Leverage structured logging with any provider.

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