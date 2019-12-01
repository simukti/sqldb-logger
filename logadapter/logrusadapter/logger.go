package logrusadapter

import (
	"context"

	sqldblogger "github.com/simukti/sqldb-logger"
	"github.com/sirupsen/logrus"
)

type logrusAdapter struct {
	logger *logrus.Logger
}

func New(logger *logrus.Logger) sqldblogger.Logger {
	return &logrusAdapter{logger: logger}
}

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
	default:
		entry.Debug(msg)
	}
}
