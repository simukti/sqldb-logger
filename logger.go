package sqldblogger

import (
	"context"
	"database/sql/driver"
	"fmt"
	"strconv"
	"time"
)

type Level uint8

const (
	LevelError Level = iota
	LevelInfo
	LevelDebug
)

func (l Level) String() string {
	switch l {
	case LevelError:
		return "error" // nolint: goconst
	case LevelInfo:
		return "info" // nolint: goconst
	case LevelDebug:
		return "debug" // nolint: goconst
	default:
		return fmt.Sprintf("invalid level: %d", l)
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

func (l *logger) withQuery(query string) dataFunc {
	return func() (string, interface{}) {
		return l.opt.sqlQueryFieldname, query
	}
}

func (l *logger) withArgs(args []driver.Value) dataFunc {
	return func() (string, interface{}) {
		return l.opt.sqlArgsFieldname, parseArgs(args)
	}
}

func (l *logger) withNamedArgs(args []driver.NamedValue) dataFunc {
	return func() (string, interface{}) {
		argsVal := make([]driver.Value, len(args))

		for k, v := range args {
			argsVal[k] = v.Value
		}

		return l.opt.sqlArgsFieldname, parseArgs(argsVal)
	}
}

func (l *logger) log(ctx context.Context, lvl Level, msg string, start time.Time, err error, datas ...dataFunc) {
	if !(lvl <= l.opt.minimumLogLevel) {
		return
	}

	data := map[string]interface{}{
		l.opt.timestampFieldname: time.Now().Unix(),
		l.opt.durationFieldname:  time.Since(start),
	}

	if lvl == LevelError {
		data[l.opt.errorFieldname] = err.Error()
	}

	for _, d := range datas {
		k, v := d()

		if !l.opt.logArgs && k == l.opt.sqlArgsFieldname {
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
				a = string(v[:maxArgValueLen]) + " (truncated " + strconv.Itoa(len(v)-maxArgValueLen) + " bytes)"
			}
		case string:
			if len(v) > maxArgValueLen {
				a = v[:maxArgValueLen] + " (truncated " + strconv.Itoa(len(v)-maxArgValueLen) + " bytes)"
			}
		}

		args[k] = a
	}

	return args
}
