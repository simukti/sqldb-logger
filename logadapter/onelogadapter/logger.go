package onelogadapter

import (
	"context"

	"github.com/francoispqt/onelog"

	sqldblogger "github.com/ntwrk1/sqldb-logger"
)

type onelogAdapter struct {
	logger *onelog.Logger
}

// New set onelog logger as backend as an example on how it process log from sqldblogger.Log().
func New(logger *onelog.Logger) sqldblogger.Logger {
	return &onelogAdapter{logger: logger}
}

// Log implement sqldblogger.Logger and log it as is.
// To use context.Context values, please copy this file and adjust to your needs.
func (oa *onelogAdapter) Log(_ context.Context, level sqldblogger.Level, msg string, data map[string]interface{}) {
	var chain onelog.ChainEntry

	switch level {
	case sqldblogger.LevelError:
		chain = oa.logger.ErrorWith(msg)
	case sqldblogger.LevelInfo:
		chain = oa.logger.InfoWith(msg)
	case sqldblogger.LevelDebug:
		chain = oa.logger.DebugWith(msg)
	default:
		// trace will use onelog debug
		chain = oa.logger.DebugWith(msg)
	}

	for k, v := range data {
		chain.Any(k, v)
	}

	chain.Write()
}
