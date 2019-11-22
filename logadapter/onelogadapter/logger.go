package onelogadapter

import (
	"context"

	"github.com/francoispqt/onelog"
	sqldblogger "github.com/simukti/sqldb-logger"
)

type onelogAdapter struct {
	logger *onelog.Logger
}

func New(logger *onelog.Logger) sqldblogger.Logger {
	return &onelogAdapter{logger: logger}
}

func (oa *onelogAdapter) Log(ctx context.Context, level sqldblogger.Level, msg string, data map[string]interface{}) {
	var chain onelog.ChainEntry

	switch level {
	case sqldblogger.LevelError:
		chain = oa.logger.ErrorWith(msg)
	case sqldblogger.LevelInfo:
		chain = oa.logger.InfoWith(msg)
	case sqldblogger.LevelDebug:
		chain = oa.logger.DebugWith(msg)
	default:
		chain = oa.logger.DebugWith(msg)
	}

	for k, v := range data {
		chain.Any(k, v)
	}

	chain.Write()
}
