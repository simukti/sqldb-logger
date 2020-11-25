package sqldblogger

import (
	"context"
	"database/sql/driver"
	"fmt"
	"strconv"
	"time"
)

// Level is a log level which filterable by minimum level option.
type Level uint8

const (
	// LevelTrace is the lowest level and the most detailed.
	// Use this if you want to know interaction flow from prepare, statement, execution to result/rows.
	LevelTrace Level = iota
	// LevelDebug is used by non Queryer(Context) and Execer(Context) call like Ping() and Connect().
	LevelDebug
	// LevelInfo is used by Queryer, Execer, Preparer, and Stmt.
	LevelInfo
	// LevelError is used on actual driver error or when driver not implement some optional sql/driver interface.
	LevelError
)

// String implement Stringer to convert type Level to string.
// nolint // disable goconst check
func (l Level) String() string {
	switch l {
	case LevelTrace:
		return "trace"
	case LevelDebug:
		return "debug"
	case LevelInfo:
		return "info"
	case LevelError:
		return "error"
	default:
		return fmt.Sprintf("(invalid level): %d", l)
	}
}

// Logger interface copied from:
// https://github.com/jackc/pgx/blob/f3a3ee1a0e5c8fc8991928bcd06fdbcd1ee9d05c/logger.go#L46-L49
type Logger interface {
	Log(ctx context.Context, level Level, msg string, data map[string]interface{})
}

// logger internal logger wrapper
type logger struct {
	logger Logger
	opt    *options
}

// dataFunc for extra data to be added to log
type dataFunc func() (string, interface{})

// withUID used to set unique id per call scope.
func (l *logger) withUID(k, v string) dataFunc {
	return func() (string, interface{}) {
		if v == "" {
			return k, nil
		}

		return k, v
	}
}

func (l *logger) withQuery(query string) dataFunc {
	return func() (string, interface{}) {
		return l.opt.sqlQueryFieldname, query
	}
}

func (l *logger) withArgs(args []driver.Value) dataFunc {
	return func() (string, interface{}) {
		if !l.opt.logArgs {
			return l.opt.sqlArgsFieldname, nil
		}

		return l.withKeyArgs(l.opt.sqlArgsFieldname, args)()
	}
}

func (l *logger) withKeyArgs(key string, args []driver.Value) dataFunc {
	return func() (string, interface{}) {
		if len(args) == 0 {
			return key, nil
		}

		return key, parseArgs(args)
	}
}

func (l *logger) log(ctx context.Context, lvl Level, msg string, start time.Time, err error, datas ...dataFunc) {
	if lvl < l.opt.minimumLogLevel {
		return
	}

	if !l.opt.logDriverErrSkip && err == driver.ErrSkip {
		return
	}

	data := map[string]interface{}{
		l.opt.timeFieldname:     l.opt.timeFormat.format(time.Now()),
		l.opt.durationFieldname: l.opt.durationUnit.format(time.Since(start)),
	}

	if l.opt.includeStartTime {
		data[l.opt.startTimeFieldname] = l.opt.timeFormat.format(start)
	}

	if lvl == LevelError && err != nil {
		data[l.opt.errorFieldname] = err.Error()
	}

	for _, d := range datas {
		k, v := d()

		if k == l.opt.sqlArgsFieldname && !l.opt.logArgs {
			continue
		}

		// don't log nil value
		if v == nil {
			continue
		}

		if k == l.opt.sqlQueryFieldname && l.opt.sqlQueryAsMsg {
			msg = v.(string)
			continue
		}

		data[k] = v
	}

	l.logger.Log(ctx, lvl, msg, data)
}

// maxArgValueLen []byte and string more than this length will be truncated.
const maxArgValueLen int = 64

// parseArgs will trim argument value if it is []byte or string more than maxArgValueLen.
// Copied from https://github.com/jackc/pgx/blob/f3a3ee1a0e5c8fc8991928bcd06fdbcd1ee9d05c/logger.go#L79
// and modified accordingly.
func parseArgs(argsVal []driver.Value) []interface{} {
	args := make([]interface{}, len(argsVal))

	for k, a := range argsVal {
		switch v := a.(type) {
		case []byte:
			if len(v) < maxArgValueLen {
				a = string(v)
			} else {
				a = string(v[:maxArgValueLen]) + " (" + strconv.Itoa(len(v)-maxArgValueLen) + " bytes truncated)"
			}
		case string:
			if len(v) > maxArgValueLen {
				a = v[:maxArgValueLen] + " (" + strconv.Itoa(len(v)-maxArgValueLen) + " bytes truncated)"
			}
		}

		args[k] = a
	}

	return args
}

// namedValuesToValues is type conversion ONLY for logging arguments.
func namedValuesToValues(args []driver.NamedValue) []driver.Value {
	argsVal := make([]driver.Value, len(args))

	for k, v := range args {
		argsVal[k] = v.Value
	}

	return argsVal
}
