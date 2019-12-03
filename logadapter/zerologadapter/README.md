## SQLDB-LOGGER ZEROLOG ADAPTER

![stdout sample](./zerolog.jpg?raw=true "stdout output")

```go
logger := zerolog.New(os.Stdout)
// populate log pre-fields here before set to
db := sqldblogger.OpenDriver(
    dsn,
    &mysql.MySQLDriver{},
    zerologadapter.New(logger),
    // optional config...
)
```