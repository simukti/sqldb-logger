package sqldblogger

type options struct {
	errorFieldname     string
	durationFieldname  string
	timestampFieldname string
	sqlQueryFieldname  string
	sqlArgsFieldname   string
	minimumLogLevel    Level
}

type Option func(*options)

func setDefaultOptions(opt *options) {
	opt.errorFieldname = "error"
	opt.durationFieldname = "duration"
	opt.timestampFieldname = "timestamp"
	opt.sqlQueryFieldname = "query"
	opt.sqlArgsFieldname = "args"
	opt.minimumLogLevel = LevelInfo
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
