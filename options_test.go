package sqldblogger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfigs(t *testing.T) {
	cfg := &config{}
	setDefaultConfig(cfg)
	assert.Equal(t, "error", cfg.errorFieldname)
	assert.Equal(t, "duration", cfg.durationFieldname)
	assert.Equal(t, "timestamp", cfg.timestampFieldname)
	assert.Equal(t, "query", cfg.sqlQueryFieldname)
	assert.Equal(t, "args", cfg.sqlArgsFieldname)
	assert.Equal(t, LevelInfo, cfg.minimumLogLevel)
}

func TestWithErrorFieldname(t *testing.T) {
	cfg := &config{}
	WithErrorFieldname("errorfield")(cfg)
	assert.Equal(t, "errorfield", cfg.errorFieldname)
}

func TestWithDurationFieldname(t *testing.T) {
	cfg := &config{}
	WithDurationFieldname("durfield")(cfg)
	assert.Equal(t, "durfield", cfg.durationFieldname)
}

func TestWithMinimumLevel(t *testing.T) {
	cfg := &config{}
	WithMinimumLevel(LevelNotice)(cfg)
	assert.Equal(t, LevelNotice, cfg.minimumLogLevel)

	WithMinimumLevel(Level(99))(cfg)
	assert.NotEqual(t, Level(99), cfg.minimumLogLevel)
}

func TestWithTimestampFieldname(t *testing.T) {
	cfg := &config{}
	WithTimestampFieldname("ts")(cfg)
	assert.Equal(t, "ts", cfg.timestampFieldname)
}

func TestWithSQLQueryFieldname(t *testing.T) {
	cfg := &config{}
	WithSQLQueryFieldname("sqlq")(cfg)
	assert.Equal(t, "sqlq", cfg.sqlQueryFieldname)
}

func TestWithSQLArgsFieldname(t *testing.T) {
	cfg := &config{}
	WithSQLArgsFieldname("sqlargs")(cfg)
	assert.Equal(t, "sqlargs", cfg.sqlArgsFieldname)
}
