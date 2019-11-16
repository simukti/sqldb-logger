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
	assert.Equal(t, "timestamp", cfg.timestampFieldname)
	assert.Equal(t, "query", cfg.sqlQueryFieldname)
	assert.Equal(t, "args", cfg.sqlArgsFieldname)
	assert.Equal(t, true, cfg.logArgs)
	assert.Equal(t, LevelInfo, cfg.minimumLogLevel)
	assert.Equal(t, DurationMillisecond, cfg.durationUnit)
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
	WithMinimumLevel(LevelDebug)(cfg)
	assert.Equal(t, LevelDebug, cfg.minimumLogLevel)

	WithMinimumLevel(Level(99))(cfg)
	assert.NotEqual(t, Level(99), cfg.minimumLogLevel)
}

func TestWithTimestampFieldname(t *testing.T) {
	cfg := &options{}
	WithTimestampFieldname("ts")(cfg)
	assert.Equal(t, "ts", cfg.timestampFieldname)
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
