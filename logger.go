package sqldblogger

import (
	"context"
	"database/sql/driver"
	"time"
)

type Level uint8

const (
	LevelError Level = iota
	LevelInfo
	LevelNotice
	LevelDebug
)

func (l Level) String() string {
	var stringMap = map[Level]string{
		LevelError:  "error",
		LevelInfo:   "info",
		LevelNotice: "notice",
		LevelDebug:  "debug",
	}

	var s string
	if v, ok := stringMap[l]; ok {
		s = v
	}

	return s
}

// Logger interface copied from https://github.com/jackc/pgx/blob/master/logger.go
// Copyright (c) 2013 Jack Christensen
// https://github.com/jackc/pgx/blob/master/LICENSE
type Logger interface {
	Log(ctx context.Context, level Level, msg string, data map[string]interface{})
}

type NullLogger struct{}

func (nl *NullLogger) Log(ctx context.Context, level Level, msg string, data map[string]interface{}) {}

// logger internal logger wrapper
type logger struct {
	logger Logger
	cfg    *config
}

// dataFunc for extra data to be added to log
type dataFunc func() (string, interface{})

func (l *logger) withQuery(query string) dataFunc {
	return func() (string, interface{}) {
		return l.cfg.sqlQueryFieldname, query
	}
}

func (l *logger) withArgs(args []driver.Value) dataFunc {
	return func() (string, interface{}) {
		return l.cfg.sqlArgsFieldname, args
	}
}

func (l *logger) withNamedArgs(args []driver.NamedValue) dataFunc {
	return func() (string, interface{}) {
		return l.cfg.sqlArgsFieldname, args
	}
}

func (l *logger) log(ctx context.Context, lvl Level, msg string, start time.Time, err error, datas ...dataFunc) {
	if !(lvl <= l.cfg.minimumLogLevel) {
		return
	}

	data := map[string]interface{}{
		l.cfg.timestampFieldname: time.Now().Unix(),
		l.cfg.durationFieldname:  time.Since(start),
		l.cfg.errorFieldname:     err,
	}

	for _, d := range datas {
		k, v := d()
		data[k] = v
	}

	l.logger.Log(ctx, lvl, msg, data)
}
