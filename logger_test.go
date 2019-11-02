package sqldblogger

import (
	"context"
	"database/sql/driver"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLevel_String(t *testing.T) {
	tt := map[Level]string{
		LevelError:  "error",
		LevelInfo:   "info",
		LevelNotice: "notice",
		LevelDebug:  "debug",
	}

	for l, s := range tt {
		assert.Equal(t, l.String(), s)
	}
}

func TestNullLogger_Log(t *testing.T) {
	lg := &NullLogger{}
	lg.Log(context.Background(), LevelInfo, "msg", nil)
	assert.Implements(t, (*Logger)(nil), lg)
}

func TestWithQuery(t *testing.T) {
	cfg := &config{}
	setDefaultConfig(cfg)
	l := &logger{cfg: cfg}
	k, v := l.withQuery("query")()
	assert.Equal(t, cfg.sqlQueryFieldname, k)
	assert.Equal(t, "query", fmt.Sprint(v))
}

func TestWithArgs(t *testing.T) {
	cfg := &config{}
	setDefaultConfig(cfg)
	l := &logger{cfg: cfg}
	k, v := l.withArgs([]driver.Value{})()
	assert.Equal(t, cfg.sqlArgsFieldname, k)
	assert.Equal(t, []driver.Value{}, v)
}

func TestWithNamedArgs(t *testing.T) {
	cfg := &config{}
	setDefaultConfig(cfg)
	l := &logger{cfg: cfg}
	k, v := l.withNamedArgs([]driver.NamedValue{})()
	assert.Equal(t, cfg.sqlArgsFieldname, k)
	assert.Equal(t, []driver.NamedValue{}, v)
}

func TestLogInternalWithMinimumLevel(t *testing.T) {
	cfg := &config{}
	setDefaultConfig(cfg)
	l := &logger{cfg: cfg, logger: &NullLogger{}}
	l.log(context.Background(), LevelDebug, "msg", time.Now(), nil)
}

func TestLogInternal(t *testing.T) {
	cfg := &config{}
	setDefaultConfig(cfg)
	l := &logger{cfg: cfg, logger: &NullLogger{}}
	l.log(context.Background(), LevelInfo, "msg", time.Now(), nil)
}

func TestLogInternalWithData(t *testing.T) {
	cfg := &config{}
	setDefaultConfig(cfg)
	l := &logger{cfg: cfg, logger: &NullLogger{}}
	l.log(context.Background(), LevelInfo, "msg", time.Now(), nil, l.withQuery("query"))
}
