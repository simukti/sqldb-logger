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

// Option Logger option func.
type Option func(*options)

// WithErrorFieldname to customize error fieldname on log output.
//
// Default: "error"
func WithErrorFieldname(name string) Option {
	return func(opt *options) {
		opt.errorFieldname = name
	}
}

// WithDurationFieldname to customize duration fieldname on log output.
//
// Default: "duration"
func WithDurationFieldname(name string) Option {
	return func(opt *options) {
		opt.durationFieldname = name
	}
}

// WithTimeFieldname to customize log timestamp fieldname on log output.
//
// Default: "time"
func WithTimeFieldname(name string) Option {
	return func(opt *options) {
		opt.timeFieldname = name
	}
}

// WithSQLQueryFieldname to customize SQL query fieldname on log output.
//
// Default: "query"
func WithSQLQueryFieldname(name string) Option {
	return func(opt *options) {
		opt.sqlQueryFieldname = name
	}
}

// WithSQLArgsFieldname to customize SQL query arguments fieldname on log output.
//
// Default: "args"
func WithSQLArgsFieldname(name string) Option {
	return func(opt *options) {
		opt.sqlArgsFieldname = name
	}
}

// WithSQLArgsFieldname set minimum level to be logged.
//
// Default: LevelDebug
//
// Options: LevelDebug < LevelInfo < LevelError
func WithMinimumLevel(lvl Level) Option {
	return func(opt *options) {
		if lvl < LevelError || lvl > LevelDebug {
			return
		}

		opt.minimumLogLevel = lvl
	}
}

// WithLogArguments set flag to log SQL query argument or not.
//
// Default: true
//
// For some system it is not recommended to log SQL argument because it may contain sensitive payload.
func WithLogArguments(flag bool) Option {
	return func(opt *options) {
		opt.logArgs = flag
	}
}

// WithLogDriverErrorSkip set flag for driver.ErrSkip.
//
// Default: true
//
// If driver not implement optional interfaces, driver will return driver.ErrSkip and sql.DB will handle that.
// driver.ErrSkip could be false alarm in log analyzer because it was not actual error from app.
func WithLogDriverErrorSkip(flag bool) Option {
	return func(opt *options) {
		opt.logDriverErrSkip = flag
	}
}

// WithDurationUnit to customize log duration unit.
//
// Default: DurationMillisecond
//
// Options:
// - DurationMillisecond
// - DurationMicrosecond
// - DurationNanosecond
func WithDurationUnit(unit DurationUnit) Option {
	return func(opt *options) {
		opt.durationUnit = unit
	}
}

// WithTimeFormat to customize log time format.
//
// Default: TimeFormatUnix
//
// Options:
// - TimeFormatUnix
// - TimeFormatUnixNano
// - TimeFormatRFC3339
// - TimeFormatRFC3339Nano
func WithTimeFormat(format TimeFormat) Option {
	return func(opt *options) {
		if format < TimeFormatUnix || format > TimeFormatRFC3339Nano {
			return
		}

		opt.timeFormat = format
	}
}
