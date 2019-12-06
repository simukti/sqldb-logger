// +build unit

package sqldblogger

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfigs(t *testing.T) {
	cfg := &options{}
	setDefaultOptions(cfg)
	assert.Equal(t, "error", cfg.errorFieldname)
	assert.Equal(t, "duration", cfg.durationFieldname)
	assert.Equal(t, "time", cfg.timeFieldname)
	assert.Equal(t, "query", cfg.sqlQueryFieldname)
	assert.Equal(t, "args", cfg.sqlArgsFieldname)
	assert.Equal(t, false, cfg.sqlQueryAsMsg)
	assert.Equal(t, true, cfg.logArgs)
	assert.Equal(t, false, cfg.logDriverErrSkip)
	assert.Equal(t, LevelDebug, cfg.minimumLogLevel)
	assert.Equal(t, DurationMillisecond, cfg.durationUnit)
	assert.Equal(t, "conn_id", cfg.connIDFieldname)
	assert.Equal(t, "stmt_id", cfg.stmtIDFieldname)
	assert.Equal(t, "tx_id", cfg.txIDFieldname)
}

func TestWithErrorFieldname(t *testing.T) {
	cfg := &options{}
	WithErrorFieldname("errorfield")(cfg)
	assert.Equal(t, "errorfield", cfg.errorFieldname)
}

func TestWithDurationFieldname(t *testing.T) {
	cfg := &options{}
	WithDurationFieldname("durfield")(cfg)
	assert.Equal(t, "durfield", cfg.durationFieldname)
}

func TestWithMinimumLevel(t *testing.T) {
	cfg := &options{}
	WithMinimumLevel(LevelTrace)(cfg)
	assert.Equal(t, LevelTrace, cfg.minimumLogLevel)

	WithMinimumLevel(Level(99))(cfg)
	assert.NotEqual(t, Level(99), cfg.minimumLogLevel)
}

func TestWithTimestampFieldname(t *testing.T) {
	cfg := &options{}
	WithTimeFieldname("ts")(cfg)
	assert.Equal(t, "ts", cfg.timeFieldname)
}

func TestWithSQLQueryFieldname(t *testing.T) {
	cfg := &options{}
	WithSQLQueryFieldname("sqlq")(cfg)
	assert.Equal(t, "sqlq", cfg.sqlQueryFieldname)
}

func TestWithSQLArgsFieldname(t *testing.T) {
	cfg := &options{}
	WithSQLArgsFieldname("sqlargs")(cfg)
	assert.Equal(t, "sqlargs", cfg.sqlArgsFieldname)
}

func TestWithLogArguments(t *testing.T) {
	cfg := &options{}
	WithLogArguments(false)(cfg)
	assert.Equal(t, false, cfg.logArgs)
}

func TestWithLogDriverErrSkip(t *testing.T) {
	cfg := &options{}
	WithLogDriverErrorSkip(true)(cfg)
	assert.Equal(t, true, cfg.logDriverErrSkip)
}

func TestWithDurationUnit(t *testing.T) {
	cfg := &options{}
	WithDurationUnit(DurationMicrosecond)(cfg)
	assert.Equal(t, DurationMicrosecond, cfg.durationUnit)
}

func TestWithDurationUnitFormat(t *testing.T) {
	dur := time.Second * 1

	tt := []struct {
		dur DurationUnit
		val float64
	}{
		{dur: DurationNanosecond, val: 1000000000},
		{dur: DurationMicrosecond, val: 1000000},
		{dur: DurationMillisecond, val: 1000},
		{dur: DurationUnit(99), val: 1000000000},
	}

	for _, tc := range tt {
		v := tc.dur.format(dur)
		assert.Equal(t, tc.val, v)
	}
}

func TestWithTimeFormat(t *testing.T) {
	t.Run("Valid format", func(t *testing.T) {
		cfg := &options{}
		WithTimeFormat(TimeFormatRFC3339)(cfg)
		assert.Equal(t, TimeFormatRFC3339, cfg.timeFormat)
	})

	t.Run("Invalid format", func(t *testing.T) {
		cfg := &options{}
		WithTimeFormat(TimeFormat(99))(cfg)
		assert.Equal(t, TimeFormatUnix, cfg.timeFormat)
	})
}

func TestWithTimeFormatResult(t *testing.T) {
	now := time.Now()
	tt := []struct {
		tf  TimeFormat
		val interface{}
	}{
		{tf: TimeFormatUnix, val: now.Unix()},
		{tf: TimeFormatUnixNano, val: now.UnixNano()},
		{tf: TimeFormatRFC3339, val: now.Format(time.RFC3339)},
		{tf: TimeFormatRFC3339Nano, val: now.Format(time.RFC3339Nano)},
		{tf: TimeFormat(99), val: now.Unix()},
	}

	for _, tc := range tt {
		v := tc.tf.format(now)
		assert.Equal(t, tc.val, v)
	}
}

func TestWithSQLQueryAsMessage(t *testing.T) {
	cfg := &options{}
	setDefaultOptions(cfg)
	WithSQLQueryAsMessage(true)(cfg)
	assert.Equal(t, true, cfg.sqlQueryAsMsg)
}

func TestWithConnectionIDFieldname(t *testing.T) {
	cfg := &options{}
	WithConnectionIDFieldname("connid")(cfg)
	assert.Equal(t, "connid", cfg.connIDFieldname)
}

func TestWithStatementIDFieldname(t *testing.T) {
	cfg := &options{}
	WithStatementIDFieldname("stmtid")(cfg)
	assert.Equal(t, "stmtid", cfg.stmtIDFieldname)
}

func TestWithTransactionIDFieldname(t *testing.T) {
	cfg := &options{}
	WithTransactionIDFieldname("trxid")(cfg)
	assert.Equal(t, "trxid", cfg.txIDFieldname)
}
