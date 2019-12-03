## SQLDB-LOGGER ONELOG ADAPTER

![stdout sample](./onelog.jpg?raw=true "stdout output")

```go
logger := onelog.New(os.Stdout, onelog.ALL)
// populate log pre-fields here before set to OpenDriver
db := sqldblogger.OpenDriver(
    dsn,
    &mysql.MySQLDriver{},
    onelogadapter.New(logger),
    // optional config...
)
```