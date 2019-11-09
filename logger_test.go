// +build unit

package sqldblogger

import (
	"bytes"
	"context"
	"database/sql/driver"
	"encoding/json"
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
		Level(99):   "invalid level: 99",
	}

	for l, s := range tt {
		assert.Equal(t, l.String(), s)
	}
}

func TestNullLogger_Log(t *testing.T) {
	lg := &bufferTestLogger{}
	lg.Log(context.TODO(), LevelInfo, "msg", nil)
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
	assert.Equal(t, []interface{}{}, v)
}

func TestWithNamedArgs(t *testing.T) {
	args := []driver.NamedValue{
		{
			Name:    "n1",
			Ordinal: 0,
			Value:   "arg1",
		},
	}

	cfg := &config{}
	setDefaultConfig(cfg)
	l := &logger{cfg: cfg}
	k, v := l.withNamedArgs(args)()
	assert.Equal(t, cfg.sqlArgsFieldname, k)
	assert.Equal(t, []interface{}{"arg1"}, v)
}

func TestLogInternalWithMinimumLevel(t *testing.T) {
	cfg := &config{}
	setDefaultConfig(cfg)
	l := &logger{cfg: cfg, logger: &bufferTestLogger{}}
	l.log(context.TODO(), LevelDebug, "msg", time.Now(), nil)
}

func TestLogInternal(t *testing.T) {
	cfg := &config{}
	setDefaultConfig(cfg)
	l := &logger{cfg: cfg, logger: &bufferTestLogger{}}
	l.log(context.TODO(), LevelInfo, "msg", time.Now(), nil)
}

func TestLogInternalWithData(t *testing.T) {
	cfg := &config{}
	setDefaultConfig(cfg)
	l := &logger{cfg: cfg, logger: &bufferTestLogger{}}
	l.log(context.TODO(), LevelInfo, "msg", time.Now(), nil, l.withQuery("query"))
}

func TestLogInternalErrorLevel(t *testing.T) {
	cfg := &config{}
	setDefaultConfig(cfg)
	bl := &bufferTestLogger{}
	l := &logger{cfg: cfg, logger: bl}
	l.log(context.TODO(), LevelError, "msg", time.Now(), fmt.Errorf("dummy"), l.withQuery("query"))

	var content bufLog
	err := json.Unmarshal(bl.Bytes(), &content)
	assert.NoError(t, err)
	assert.Contains(t, content.Data, cfg.errorFieldname)
	assert.Contains(t, content.Data, cfg.sqlQueryFieldname)
	assert.Contains(t, content.Data, cfg.timestampFieldname)
	assert.Contains(t, content.Data, cfg.durationFieldname)
	bl.Reset()
}

func TestLogTrimStringArgs(t *testing.T) {
	cfg := &config{}
	setDefaultConfig(cfg)

	longArgVal := "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."
	bl := &bufferTestLogger{}
	l := &logger{cfg: cfg, logger: bl}
	l.log(
		context.TODO(),
		LevelInfo,
		"msg",
		time.Now(),
		nil,
		l.withQuery("query"),
		l.withArgs([]driver.Value{
			longArgVal,
			[]byte(longArgVal),
			[]byte("short"),
		}),
	)

	var content bufLog
	err := json.Unmarshal(bl.Bytes(), &content)
	assert.NoError(t, err)
	assert.Contains(t, content.Data, cfg.sqlQueryFieldname)
	assert.Contains(t, content.Data, cfg.timestampFieldname)
	assert.Contains(t, content.Data, cfg.durationFieldname)
	assert.Contains(t, content.Data, cfg.sqlArgsFieldname)
	trimmedArg, ok := content.Data[cfg.sqlArgsFieldname].([]interface{})
	assert.True(t, ok)
	assert.Equal(t,
		fmt.Sprintf("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do  (truncated %d bytes)", len(longArgVal)-maxArgValueLen),
		trimmedArg[0],
	)
	assert.Equal(t,
		fmt.Sprintf("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do  (truncated %d bytes)", len(longArgVal)-maxArgValueLen),
		trimmedArg[1],
	)
	assert.Equal(t,
		"short",
		trimmedArg[2],
	)
	bl.Reset()
}

type bufferTestLogger struct {
	bytes.Buffer
}

type bufLog struct {
	Level string                 `json:"level"`
	Data  map[string]interface{} `json:"data"`
}

func (bl *bufferTestLogger) Log(_ context.Context, level Level, msg string, data map[string]interface{}) {
	bl.Reset()
	_ = json.NewEncoder(bl).Encode(bufLog{level.String(), data})
}
