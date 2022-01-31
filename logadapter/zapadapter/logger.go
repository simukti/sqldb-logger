package zapadapter

import (
	"context"

	"go.uber.org/zap"

	sqldblogger "github.com/ntwrk1/sqldb-logger"
)

type zapAdapter struct {
	logger *zap.Logger
}

// New set zap logger as backend as an example on how it process log from sqldblogger.Log().
func New(logger *zap.Logger) sqldblogger.Logger {
	return &zapAdapter{logger: logger}
}

// Log implement sqldblogger.Logger and log it as is.
// To use context.Context values, please copy this file and adjust to your needs.
func (zp *zapAdapter) Log(_ context.Context, level sqldblogger.Level, msg string, data map[string]interface{}) {
	fields := make([]zap.Field, len(data))
	i := 0

	for k, v := range data {
		fields[i] = zap.Any(k, v)
		i++
	}

	switch level {
	case sqldblogger.LevelError:
		zp.logger.Error(msg, fields...)
	case sqldblogger.LevelInfo:
		zp.logger.Info(msg, fields...)
	case sqldblogger.LevelImportantInfo:
		zp.logger.Info(msg, fields...)
	case sqldblogger.LevelDebug:
		zp.logger.Debug(msg, fields...)
	default:
		// trace will use zap debug
		zp.logger.Debug(msg, fields...)
	}
}
