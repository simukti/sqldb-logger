## SQLDB-LOGGER ZAP ADAPTER

![stdout sample](./zap.jpg?raw=true "stdout output")

```go
zapCfg := zap.NewProductionConfig()
zapCfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel) // whatever minimum level
zapCfg.DisableCaller = true
logger, _ := zapCfg.Build()
// populate log pre-fields here before set to OpenDriver
db, err := sqldblogger.OpenDriver(
    dsn,
    &mysql.MySQLDriver{},
    zapadapter.New(logger),
    // optional config...
)
```