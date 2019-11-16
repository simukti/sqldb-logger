package sqldblogger

import "time"

type options struct {
	errorFieldname     string
	durationFieldname  string
	timestampFieldname string
	sqlQueryFieldname  string
	sqlArgsFieldname   string
	logArgs            bool
	minimumLogLevel    Level
	durationUnit       DurationUnit
}

type DurationUnit uint8

const (
	DurationNanosecond DurationUnit = iota
	DurationMicrosecond
	DurationMillisecond
)

func (unit DurationUnit) format(duration time.Duration) float64 {
	nanosecond := float64(duration.Nanoseconds())

	switch unit {
	case DurationNanosecond:
		return nanosecond
	case DurationMicrosecond:
		return nanosecond / float64(time.Microsecond)
	case DurationMillisecond:
		return nanosecond / float64(time.Millisecond)
	default:
		return nanosecond
	}
}

type Option func(*options)

func setDefaultOptions(opt *options) {
	opt.errorFieldname = "error"
	opt.durationFieldname = "duration"
	opt.timestampFieldname = "timestamp"
	opt.sqlQueryFieldname = "query"
	opt.sqlArgsFieldname = "args"
	opt.minimumLogLevel = LevelInfo
	opt.logArgs = true
	opt.durationUnit = DurationMillisecond
}

func WithErrorFieldname(name string) Option {
	return func(opt *options) {
		opt.errorFieldname = name
	}
}

func WithDurationFieldname(name string) Option {
	return func(opt *options) {
		opt.durationFieldname = name
	}
}

func WithTimestampFieldname(name string) Option {
	return func(opt *options) {
		opt.timestampFieldname = name
	}
}

func WithSQLQueryFieldname(name string) Option {
	return func(opt *options) {
		opt.sqlQueryFieldname = name
	}
}

func WithSQLArgsFieldname(name string) Option {
	return func(opt *options) {
		opt.sqlArgsFieldname = name
	}
}

func WithMinimumLevel(lvl Level) Option {
	return func(opt *options) {
		if lvl < LevelError || lvl > LevelDebug {
			return
		}

		opt.minimumLogLevel = lvl
	}
}

func WithLogArguments(flag bool) Option {
	return func(opt *options) {
		opt.logArgs = flag
	}
}

func WithDurationUnit(unit DurationUnit) Option {
	return func(opt *options) {
		opt.durationUnit = unit
	}
}
