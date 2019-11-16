# sqldb-logger - Go SQL Database Driver Logger

A wrapper for Go [database SQL driver](https://github.com/golang/go/blob/master/src/database/sql/driver/driver.go) which will log driver method call from `sql.DB` call.

## WHY?

My preferred way to interact with SQL database in Go are using `*sql.DB` methods directly as much as I can.

I want to:

- have simple [log interface](https://github.com/jackc/pgx/blob/f3a3ee1a0e5c8fc8991928bcd06fdbcd1ee9d05c/logger.go#L46-L49) but for all my commonly use SQL driver.
- have configurable log field name.
- have configurable duration unit.
- have configurable minimum logging level.
- do one thing only, log my SQL database interaction.
- leverage structured logging with any provider.

Because I haven't found SQL logger with that features, so why not created myself?

## HOW IT WORKS?

Having database connection initiation only once when the app start, it should be minimal change.

Let's say we have existing `sql.Open` using commonly used driver.

### MYSQL DRIVER

```go
// import _ "github.com/go-sql-driver/mysql"

dsn := "username:passwd@tcp(mysqlserver:3306)/dbname?parseTime=true"
db, err := sql.Open("mysql", dsn) // db is *sql.DB
```

Change it with:

```go
// import "github.com/go-sql-driver/mysql"
// "github.com/rs/zerolog"
// sqldblogger "github.com/simukti/sqldb-logger"
// "github.com/simukti/sqldb-logger/logadapter/zerologadapter"

// first define your logger (here we use zerolog for example)
// logger := zerolog.New() // for development with colorful console output
logger := zerolog.New() // for JSON output to stdout
db, err := sqldblogger.Open(dsn, &mysql.MySQLDriver{}, zerologadapter.New(logger))
```

## CAN I USE IT WITH *** QUERY BUILDER?

As long as your query builder runner accept `*sql.DB`, you can use it as is.

## PROVIDED LOGGER ADAPTER

There are 3 adapters within this repo:

- [Zerolog adapter](logadapter/zerologadapter): Using [zerolog](https://github.com/rs/zerolog) as its logger.
- [Onelog adapter](logadapter/onelogadapter): Using [onelog](https://github.com/francoispqt/onelog) as its logger.
- [Zap adapter](logadapter/zapadapter): Using [zap](https://github.com/uber-go/zap) as its logger.

Implements another logger must follow these simple interface:

```go
type Logger interface {
	Log(ctx context.Context, level Level, msg string, data map[string]interface{})
}
```

## CUSTOMIZE LOG OUTPUT

All of log fieldname are configurable, including duration format, whether to log query arguments or not, and minimum level to be logged. [See here for all of its options](./options.go)

## CREDITS

- [pgx](https://github.com/jackc/pgx) for awesome PostgreSQL driver.