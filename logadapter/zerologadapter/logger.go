package zerologadapter

import (
	"context"

	"github.com/rs/zerolog"
	sqldblogger "github.com/simukti/sqldb-logger"
)

type zerologAdapter struct {
	logger zerolog.Logger
}

func New(logger zerolog.Logger) sqldblogger.Logger {
	return &zerologAdapter{logger: logger}
}

// Log implements sqldblogger.Logger
func (zl *zerologAdapter) Log(ctx context.Context, level sqldblogger.Level, msg string, data map[string]interface{}) {
	var lvl zerolog.Level

	switch level {
	case sqldblogger.LevelError:
		lvl = zerolog.ErrorLevel
	case sqldblogger.LevelInfo:
		lvl = zerolog.InfoLevel
	case sqldblogger.LevelDebug:
		lvl = zerolog.DebugLevel
	default:
		lvl = zerolog.DebugLevel
	}

	zl.logger.WithLevel(lvl).Fields(data).Msg(msg)
}
