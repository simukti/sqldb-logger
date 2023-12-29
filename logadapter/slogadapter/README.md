## SQLDB-LOGGER log/slog ADAPTER

sqldb-logger log adapter for go's standard lib's [log/slog](https://pkg.go.dev/log/slog)

```go
logger, _ := slog.Default()
// populate log pre-fields here before set to OpenDriver
db := sqldblogger.OpenDriver(
    dsn,
    &mysql.MySQLDriver{},
    slogadapter.New(logger),
    // optional config...
)
```
