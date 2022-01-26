package logrusadapter

import (
	"context"

	"github.com/sirupsen/logrus"

	sqldblogger "github.com/ntwrk1/sqldb-logger"
)

type logrusAdapter struct {
	logger *logrus.Logger
}

// New set logrus logger as backend as an example on how it process log from sqldblogger.Log().
func New(logger *logrus.Logger) sqldblogger.Logger {
	return &logrusAdapter{logger: logger}
}

// Log implement sqldblogger.Logger and log it as is.
// To use context.Context values, please copy this file and adjust to your needs.
func (l *logrusAdapter) Log(ctx context.Context, level sqldblogger.Level, msg string, data map[string]interface{}) {
	// logrus will rename "time" field to "fields.time" and provide their own time value (RFC3339)
	// see: https://github.com/sirupsen/logrus#entries
	entry := l.logger.WithContext(ctx).WithFields(data)

	switch level {
	case sqldblogger.LevelError:
		entry.Error(msg)
	case sqldblogger.LevelInfo:
		entry.Info(msg)
	case sqldblogger.LevelDebug:
		entry.Debug(msg)
	case sqldblogger.LevelTrace:
		entry.Trace(msg)
	default:
		entry.Debug(msg)
	}
}
