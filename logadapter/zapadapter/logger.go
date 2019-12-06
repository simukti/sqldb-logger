package zapadapter

import (
	"context"

	sqldblogger "github.com/simukti/sqldb-logger"
	"go.uber.org/zap"
)

type zapAdapter struct {
	logger *zap.Logger
}

func New(logger *zap.Logger) sqldblogger.Logger {
	return &zapAdapter{logger: logger}
}

// Log implements sqldblogger.Logger
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
	case sqldblogger.LevelDebug:
		zp.logger.Debug(msg, fields...)
	default:
		// trace will use zap debug
		zp.logger.Debug(msg, fields...)
	}
}
