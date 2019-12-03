```go
logger := logrus.New()
logger.Level = logrus.DebugLevel // miminum level
logger.Formatter = &logrus.JSONFormatter{} // logrus automatically add time field
// other logrus variable setup
// populate log pre-fields here before set to OpenDriver
db := sqldblogger.OpenDriver(
    dsn,
    &mysql.MySQLDriver{},
    logrusadapter.New(logger),
    // optional config...
)
```