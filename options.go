package sqldblogger

import "time"

type options struct {
	errorFieldname    string
	durationFieldname string
	timeFieldname     string
	sqlQueryFieldname string
	sqlArgsFieldname  string
	logArgs           bool
	logDriverErrSkip  bool
	minimumLogLevel   Level
	durationUnit      DurationUnit
	timeFormat        TimeFormat
}

// setDefaultOptions called first time before Log() called (see: OpenDriver()).
// To change option value, use With* functions below.
func setDefaultOptions(opt *options) {
	opt.errorFieldname = "error"
	opt.durationFieldname = "duration"
	opt.timeFieldname = "time"
	opt.sqlQueryFieldname = "query"
	opt.sqlArgsFieldname = "args"
	opt.minimumLogLevel = LevelDebug
	opt.logArgs = true
	opt.logDriverErrSkip = false
	opt.durationUnit = DurationMillisecond
	opt.timeFormat = TimeFormatUnix
}

type DurationUnit uint8

const (
	DurationNanosecond DurationUnit = iota
	DurationMicrosecond
	DurationMillisecond
)

func (du DurationUnit) format(duration time.Duration) float64 {
	nanosecond := float64(duration.Nanoseconds())

	switch du {
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

type TimeFormat uint8

const (
	TimeFormatUnix TimeFormat = iota
	TimeFormatUnixNano
	TimeFormatRFC3339
	TimeFormatRFC3339Nano
)

func (tf TimeFormat) format(logTime time.Time) interface{} {
	switch tf {
	case TimeFormatUnix:
		return logTime.Unix()
	case TimeFormatUnixNano:
		return logTime.UnixNano()
	case TimeFormatRFC3339:
		return logTime.Format(time.RFC3339)
	case TimeFormatRFC3339Nano:
		return logTime.Format(time.RFC3339Nano)
	default:
		return logTime.Unix()
	}
}

type Option func(*options)

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

func WithTimeFieldname(name string) Option {
	return func(opt *options) {
		opt.timeFieldname = name
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

func WithLogDriverErrorSkip(flag bool) Option {
	return func(opt *options) {
		opt.logDriverErrSkip = flag
	}
}

func WithDurationUnit(unit DurationUnit) Option {
	return func(opt *options) {
		opt.durationUnit = unit
	}
}

func WithTimeFormat(format TimeFormat) Option {
	return func(opt *options) {
		if format < TimeFormatUnix || format > TimeFormatRFC3339Nano {
			return
		}

		opt.timeFormat = format
	}
}
