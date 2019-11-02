package sqldblogger

type config struct {
	errorFieldname     string
	durationFieldname  string
	timestampFieldname string
	sqlQueryFieldname  string
	sqlArgsFieldname   string
	minimumLogLevel    Level
}

type Option func(*config)

func setDefaultConfig(cfg *config) {
	cfg.errorFieldname = "error"
	cfg.durationFieldname = "duration"
	cfg.timestampFieldname = "timestamp"
	cfg.sqlQueryFieldname = "query"
	cfg.sqlArgsFieldname = "args"
	cfg.minimumLogLevel = LevelInfo
}

func WithErrorFieldname(name string) Option {
	return func(cfg *config) {
		cfg.errorFieldname = name
	}
}

func WithDurationFieldname(name string) Option {
	return func(cfg *config) {
		cfg.durationFieldname = name
	}
}

func WithTimestampFieldname(name string) Option {
	return func(cfg *config) {
		cfg.timestampFieldname = name
	}
}

func WithSQLQueryFieldname(name string) Option {
	return func(cfg *config) {
		cfg.sqlQueryFieldname = name
	}
}

func WithSQLArgsFieldname(name string) Option {
	return func(cfg *config) {
		cfg.sqlArgsFieldname = name
	}
}

func WithMinimumLevel(lvl Level) Option {
	return func(cfg *config) {
		if lvl < LevelError || lvl > LevelDebug {
			return
		}

		cfg.minimumLogLevel = lvl
	}
}
