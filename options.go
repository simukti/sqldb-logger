package sqldblogger

import (
	cryptoRand "crypto/rand"
	"encoding/binary"
	"fmt"
	"math/rand"
	"time"
)

type options struct {
	errorFieldname    string
	durationFieldname string
	timeFieldname     string
	sqlQueryFieldname string
	sqlArgsFieldname  string
	stmtIDFieldname   string
	connIDFieldname   string
	txIDFieldname     string
	sqlQueryAsMsg     bool
	logArgs           bool
	logDriverErrSkip  bool
	minimumLogLevel   Level
	durationUnit      DurationUnit
	timeFormat        TimeFormat
	uidGenerator      UIDGenerator
}

// setDefaultOptions called first time before Log() called (see: OpenDriver()).
// To change option value, use With* functions below.
func setDefaultOptions(opt *options) {
	opt.errorFieldname = "error"
	opt.durationFieldname = "duration"
	opt.timeFieldname = "time"
	opt.sqlQueryFieldname = "query"
	opt.sqlArgsFieldname = "args"
	opt.stmtIDFieldname = "stmt_id"
	opt.connIDFieldname = "conn_id"
	opt.txIDFieldname = "tx_id"
	opt.sqlQueryAsMsg = false
	opt.minimumLogLevel = LevelDebug
	opt.logArgs = true
	opt.logDriverErrSkip = false
	opt.durationUnit = DurationMillisecond
	opt.timeFormat = TimeFormatUnix
	opt.uidGenerator = newDefaultUIDDGenerator()
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

// UIDGenerator interface to generate unique ID for context call (connection, statement, transaction).
type UIDGenerator interface {
	UniqueID() string
}

const (
	defaultUIDLen      = 16
	defaultUIDCharlist = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_"
)

// newDefaultUIDDGenerator default unique id generator using crypto/rand as math/rand seed.
func newDefaultUIDDGenerator() UIDGenerator {
	var s [16]byte
	if _, err := cryptoRand.Read(s[:]); err != nil {
		panic(fmt.Sprintf("sqldblogger: could not get random bytes from cryto/rand: '%s'", err.Error()))
	}

	// seed math/rand with 16 random bytes from crypto/rand to make sure rand.Seed is not 1.
	// rand.Seed will be used by rand.Read inside UniqueID().
	rand.Seed(int64(binary.LittleEndian.Uint64(s[:])))

	return &defaultUID{}
}

type defaultUID struct{}

// UniqueID Generate default 16 byte unique id using math/rand.
func (u *defaultUID) UniqueID() string {
	var random, uid [defaultUIDLen]byte
	// nolint: gosec
	if _, err := rand.Read(random[:]); err != nil {
		panic(fmt.Sprintf("sqldblogger: random read error from math/rand: '%s'", err.Error()))
	}

	for i := 0; i < defaultUIDLen; i++ {
		uid[i] = defaultUIDCharlist[random[i]&62]
	}

	return string(uid[:])
}

// Option Logger option func.
type Option func(*options)

// WithUIDGenerator set custom unique id generator for context call (connection, statement, transaction).
// To disable unique id, set UIDGenerator with UniqueID() return empty string.
func WithUIDGenerator(gen UIDGenerator) Option {
	return func(opt *options) {
		opt.uidGenerator = gen
	}
}

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
// Options: LevelTrace < LevelDebug < LevelInfo < LevelError
func WithMinimumLevel(lvl Level) Option {
	return func(opt *options) {
		if lvl > LevelError || lvl < LevelTrace {
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

// WithSQLQueryAsMessage set SQL query as message in log output (only for function call with SQL query).
//
// Default: false
func WithSQLQueryAsMessage(flag bool) Option {
	return func(opt *options) {
		opt.sqlQueryAsMsg = flag
	}
}

// WithConnectionIDFieldname to customize connection ID fieldname on log output.
//
// Default: "conn_id"
func WithConnectionIDFieldname(name string) Option {
	return func(opt *options) {
		opt.connIDFieldname = name
	}
}

// WithStatementIDFieldname to customize prepared statement ID fieldname on log output.
//
// Default: "stmt_id"
func WithStatementIDFieldname(name string) Option {
	return func(opt *options) {
		opt.stmtIDFieldname = name
	}
}

// WithTransactionIDFieldname to customize database transaction ID fieldname on log output.
//
// Default: "tx_id"
func WithTransactionIDFieldname(name string) Option {
	return func(opt *options) {
		opt.txIDFieldname = name
	}
}
