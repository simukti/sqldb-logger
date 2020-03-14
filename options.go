package sqldblogger

import (
	cryptoRand "crypto/rand"
	"encoding/binary"
	"fmt"
	"math/rand"
	"time"
)

type options struct {
	errorFieldname     string
	durationFieldname  string
	timeFieldname      string
	startTimeFieldname string
	sqlQueryFieldname  string
	sqlArgsFieldname   string
	stmtIDFieldname    string
	connIDFieldname    string
	txIDFieldname      string
	sqlQueryAsMsg      bool
	logArgs            bool
	logDriverErrSkip   bool
	wrapResult         bool
	minimumLogLevel    Level
	durationUnit       DurationUnit
	timeFormat         TimeFormat
	uidGenerator       UIDGenerator
	includeStartTime   bool
	preparerLevel      Level
	queryerLevel       Level
	execerLevel        Level
}

// setDefaultOptions called first time before Log() called (see: OpenDriver()).
// To change option value, use With* functions below.
func setDefaultOptions(opt *options) {
	opt.errorFieldname = "error"
	opt.durationFieldname = "duration"
	opt.timeFieldname = "time"
	opt.startTimeFieldname = "start"
	opt.sqlQueryFieldname = "query"
	opt.sqlArgsFieldname = "args"
	opt.stmtIDFieldname = "stmt_id"
	opt.connIDFieldname = "conn_id"
	opt.txIDFieldname = "tx_id"
	opt.sqlQueryAsMsg = false
	opt.minimumLogLevel = LevelDebug
	opt.logArgs = true
	opt.logDriverErrSkip = false
	opt.wrapResult = true
	opt.durationUnit = DurationMillisecond
	opt.timeFormat = TimeFormatUnix
	opt.uidGenerator = newDefaultUIDDGenerator()
	opt.includeStartTime = false
	opt.preparerLevel = LevelInfo
	opt.queryerLevel = LevelInfo
	opt.execerLevel = LevelInfo
}

// DurationUnit is total time spent on an actual driver function call calculated by time.Since(start).
type DurationUnit uint8

const (
	// DurationNanosecond will format time.Since() result to nanosecond unit (1/1_000_000_000 second).
	DurationNanosecond DurationUnit = iota
	// DurationMicrosecond will format time.Since() result to microsecond unit (1/1_000_000 second).
	DurationMicrosecond
	// DurationMillisecond will format time.Since() result to millisecond unit (1/1_000 second).
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

// TimeFormat is time.Now() format when Log() deliver the log message.
type TimeFormat uint8

const (
	// TimeFormatUnix will format log time to unix timestamp.
	TimeFormatUnix TimeFormat = iota
	// TimeFormatUnixNano will format log time to unix timestamp with nano seconds.
	TimeFormatUnixNano
	// TimeFormatRFC3339 will format log time to time.RFC3339 format.
	TimeFormatRFC3339
	// TimeFormatRFC3339Nano will format log time to time.RFC3339Nano format.
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

// UIDGenerator is an interface to generate unique ID for context call (connection, statement, transaction).
// The point of having unique id per context call is to easily track and analyze logs.
//
// Note: no possible way to track id when statement Execer(Context),Queryer(Context) called from under db.Tx.
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

	// using math/rand.Read because it's slightly faster than crypto/rand.Read
	// unique id always scoped under connectionID so there is no need to super-secure-random using crypto/rand.
	//
	// nolint // disable gosec check as it does not need crypto/rand
	if _, err := rand.Read(random[:]); err != nil {
		panic(fmt.Sprintf("sqldblogger: random read error from math/rand: '%s'", err.Error()))
	}

	for i := 0; i < defaultUIDLen; i++ {
		uid[i] = defaultUIDCharlist[random[i]&62]
	}

	return string(uid[:])
}

// NullUID is used to disable unique id when set to WithUIDGenerator().
type NullUID struct{}

// UniqueID return empty string and unique id will not logged.
func (u *NullUID) UniqueID() string { return "" }

// Option is optional variadic type in OpenDriver().
type Option func(*options)

// WithUIDGenerator set custom unique id generator for context call (connection, statement, transaction).
//
// To disable unique id in log output, use &NullUID{}.
//
// Default: newDefaultUIDDGenerator() called from setDefaultOptions().
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

// WithMinimumLevel set minimum level to be logged. Logger will always log level >= minimum level.
//
// Options: LevelTrace < LevelDebug < LevelInfo < LevelError
//
// Default: LevelDebug
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
// When set to false, any SQL and result/rows argument on Queryer(Context) and Execer(Context) will not logged.
//
// When set to true, argument type string and []byte will subject to trim on parseArgs() log output.
//
// Default: true
func WithLogArguments(flag bool) Option {
	return func(opt *options) {
		opt.logArgs = flag
	}
}

// WithLogDriverErrorSkip set flag for driver.ErrSkip.
//
// If driver not implement optional interfaces, driver will return driver.ErrSkip and sql.DB will handle that.
// driver.ErrSkip could be false alarm in log analyzer because it was not actual error from app.
//
// When set to false, logger will log any driver.ErrSkip.
//
// Default: true
func WithLogDriverErrorSkip(flag bool) Option {
	return func(opt *options) {
		opt.logDriverErrSkip = flag
	}
}

// WithDurationUnit to customize log duration unit.
//
// Options: DurationMillisecond | DurationMicrosecond | DurationNanosecond
//
// Default: DurationMillisecond
func WithDurationUnit(unit DurationUnit) Option {
	return func(opt *options) {
		opt.durationUnit = unit
	}
}

// WithTimeFormat to customize log time format.
//
// Options: TimeFormatUnix | TimeFormatUnixNano | TimeFormatRFC3339 | TimeFormatRFC3339Nano
//
// Default: TimeFormatUnix
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

// WithWrapResult set flag to wrap Queryer(Context) and Execer(Context) driver.Rows/driver.Result response.
//
// When set to false, result returned from db (driver.Rows/driver.Result object),
// will returned as is without wrapped inside &rows{} and &result{}.
//
// Default: true
func WithWrapResult(flag bool) Option {
	return func(opt *options) {
		opt.wrapResult = flag
	}
}

// WithIncludeStartTime flag to include actual start time before actual driver execution.
//
// Can be useful if we want to combine Log implementation with tracing from context
// and set start time span manually.
//
// Default: false
func WithIncludeStartTime(flag bool) Option {
	return func(opt *options) {
		opt.includeStartTime = flag
	}
}

// WithStartTimeFieldname to customize start time fieldname on log output.
//
// If WithIncludeStartTime true, start time fieldname will use this value.
//
// Default: "start"
func WithStartTimeFieldname(name string) Option {
	return func(opt *options) {
		opt.startTimeFieldname = name
	}
}

// WithPreparerLevel set default level of Prepare(Context) method calls.
//
// Default: LevelInfo
func WithPreparerLevel(lvl Level) Option {
	return func(opt *options) {
		opt.preparerLevel = lvl
	}
}

// WithQueryerLevel set default level of Query(Context) method calls.
//
// Default: LevelInfo
func WithQueryerLevel(lvl Level) Option {
	return func(opt *options) {
		opt.queryerLevel = lvl
	}
}

// WithExecerLevel set default level of Exec(Context) method calls.
//
// Default: LevelInfo
func WithExecerLevel(lvl Level) Option {
	return func(opt *options) {
		opt.execerLevel = lvl
	}
}
