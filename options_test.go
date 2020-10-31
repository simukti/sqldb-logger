package sqldblogger

import (
	"context"
	"database/sql/driver"
	"encoding/json"
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
	setDefaultOptions(cfg)
	WithErrorFieldname("errorfield")(cfg)
	assert.Equal(t, "errorfield", cfg.errorFieldname)
}

func TestWithDurationFieldname(t *testing.T) {
	cfg := &options{}
	setDefaultOptions(cfg)
	WithDurationFieldname("durfield")(cfg)
	assert.Equal(t, "durfield", cfg.durationFieldname)
}

func TestWithMinimumLevel(t *testing.T) {
	cfg := &options{}
	setDefaultOptions(cfg)
	WithMinimumLevel(LevelTrace)(cfg)
	assert.Equal(t, LevelTrace, cfg.minimumLogLevel)

	WithMinimumLevel(Level(99))(cfg)
	assert.NotEqual(t, Level(99), cfg.minimumLogLevel)
}

func TestWithTimestampFieldname(t *testing.T) {
	cfg := &options{}
	setDefaultOptions(cfg)
	WithTimeFieldname("ts")(cfg)
	assert.Equal(t, "ts", cfg.timeFieldname)
}

func TestWithSQLQueryFieldname(t *testing.T) {
	cfg := &options{}
	setDefaultOptions(cfg)
	WithSQLQueryFieldname("sqlq")(cfg)
	assert.Equal(t, "sqlq", cfg.sqlQueryFieldname)
}

func TestWithSQLArgsFieldname(t *testing.T) {
	cfg := &options{}
	setDefaultOptions(cfg)
	WithSQLArgsFieldname("sqlargs")(cfg)
	assert.Equal(t, "sqlargs", cfg.sqlArgsFieldname)
}

func TestWithLogArguments(t *testing.T) {
	cfg := &options{}
	setDefaultOptions(cfg)
	WithLogArguments(false)(cfg)
	assert.Equal(t, false, cfg.logArgs)
}

func TestWithLogDriverErrSkip(t *testing.T) {
	cfg := &options{}
	setDefaultOptions(cfg)
	WithLogDriverErrorSkip(true)(cfg)
	assert.Equal(t, true, cfg.logDriverErrSkip)
}

func TestWithDurationUnit(t *testing.T) {
	cfg := &options{}
	setDefaultOptions(cfg)
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
		setDefaultOptions(cfg)
		WithTimeFormat(TimeFormatRFC3339)(cfg)
		assert.Equal(t, TimeFormatRFC3339, cfg.timeFormat)
	})

	t.Run("Invalid format", func(t *testing.T) {
		cfg := &options{}
		setDefaultOptions(cfg)
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
	setDefaultOptions(cfg)
	WithConnectionIDFieldname("connid")(cfg)
	assert.Equal(t, "connid", cfg.connIDFieldname)
}

func TestWithStatementIDFieldname(t *testing.T) {
	cfg := &options{}
	setDefaultOptions(cfg)
	WithStatementIDFieldname("stmtid")(cfg)
	assert.Equal(t, "stmtid", cfg.stmtIDFieldname)
}

func TestWithTransactionIDFieldname(t *testing.T) {
	cfg := &options{}
	setDefaultOptions(cfg)
	WithTransactionIDFieldname("trxid")(cfg)
	assert.Equal(t, "trxid", cfg.txIDFieldname)
}

func TestWithWrapResult(t *testing.T) {
	cfg := &options{}
	setDefaultOptions(cfg)
	WithWrapResult(false)(cfg)
	assert.Equal(t, false, cfg.wrapResult)
}

func TestWithUIDGenerator(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		cfg := &options{}
		setDefaultOptions(cfg)
		WithUIDGenerator(&NullUID{})(cfg)

		_, ok := interface{}(cfg.uidGenerator).(*NullUID)
		assert.True(t, ok)
	})

	t.Run("Empty UID should not exist in log output", func(t *testing.T) {
		cfg := &options{}
		setDefaultOptions(cfg)
		WithUIDGenerator(&NullUID{})(cfg)

		bl := &bufferTestLogger{}
		l := &logger{opt: cfg, logger: bl}

		l.log(
			context.TODO(),
			LevelInfo,
			"msg",
			time.Now(),
			nil,
			testLogger.withUID(cfg.stmtIDFieldname, l.opt.uidGenerator.UniqueID()),
			testLogger.withQuery("query"),
			testLogger.withArgs([]driver.Value{}),
		)

		var content bufLog
		err := json.Unmarshal(bl.Bytes(), &content)
		assert.NoError(t, err)
		assert.NotContains(t, content.Data, cfg.stmtIDFieldname)
		bl.Reset()
	})
}

func TestWithIncludeStartTime(t *testing.T) {
	t.Run("Default not include", func(t *testing.T) {
		cfg := &options{}
		setDefaultOptions(cfg)

		assert.False(t, cfg.includeStartTime)
	})

	t.Run("Set start time flag true", func(t *testing.T) {
		cfg := &options{}
		setDefaultOptions(cfg)
		WithIncludeStartTime(true)(cfg)
		WithStartTimeFieldname("start_time")(cfg)
		WithTimeFormat(TimeFormatUnix)(cfg)

		assert.True(t, cfg.includeStartTime)
		assert.Equal(t, "start_time", cfg.startTimeFieldname)

		bl := &bufferTestLogger{}
		l := &logger{opt: cfg, logger: bl}
		start := time.Now()
		l.log(
			context.TODO(),
			LevelInfo,
			"msg",
			start,
			nil,
			testLogger.withUID(cfg.stmtIDFieldname, l.opt.uidGenerator.UniqueID()),
			testLogger.withQuery("query"),
			testLogger.withArgs([]driver.Value{}),
		)

		var content bufLog
		err := json.Unmarshal(bl.Bytes(), &content)
		assert.NoError(t, err)
		assert.Contains(t, content.Data, cfg.startTimeFieldname)
		assert.Equal(t, float64(start.Unix()), content.Data["start_time"])
		bl.Reset()
	})
}

func TestWithPreparerLevel(t *testing.T) {
	t.Run("Default value", func(t *testing.T) {
		cfg := &options{}
		setDefaultOptions(cfg)

		assert.Equal(t, cfg.preparerLevel, LevelInfo)
	})

	t.Run("Custom value", func(t *testing.T) {
		cfg := &options{}
		setDefaultOptions(cfg)
		WithPreparerLevel(LevelDebug)(cfg)

		assert.Equal(t, cfg.preparerLevel, LevelDebug)
	})
}

func TestWithQueryerLevel(t *testing.T) {
	t.Run("Default value", func(t *testing.T) {
		cfg := &options{}
		setDefaultOptions(cfg)

		assert.Equal(t, cfg.queryerLevel, LevelInfo)
	})

	t.Run("Custom value", func(t *testing.T) {
		cfg := &options{}
		setDefaultOptions(cfg)
		WithQueryerLevel(LevelDebug)(cfg)

		assert.Equal(t, cfg.queryerLevel, LevelDebug)
	})
}

func TestWithExecerLevel(t *testing.T) {
	t.Run("Default value", func(t *testing.T) {
		cfg := &options{}
		setDefaultOptions(cfg)

		assert.Equal(t, cfg.execerLevel, LevelInfo)
	})

	t.Run("Custom value", func(t *testing.T) {
		cfg := &options{}
		setDefaultOptions(cfg)
		WithExecerLevel(LevelDebug)(cfg)

		assert.Equal(t, cfg.execerLevel, LevelDebug)
	})
}

var uidBtest = newDefaultUIDDGenerator()

func BenchmarkUniqueID(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			uidBtest.UniqueID()
		}
	})
}
