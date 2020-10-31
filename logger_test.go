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
		LevelError: "error",
		LevelInfo:  "info",
		LevelDebug: "debug",
		LevelTrace: "trace",
		Level(99):  "(invalid level): 99",
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
	cfg := &options{}
	setDefaultOptions(cfg)
	l := &logger{opt: cfg}
	k, v := l.withQuery("query")()
	assert.Equal(t, cfg.sqlQueryFieldname, k)
	assert.Equal(t, "query", fmt.Sprint(v))
}

func TestWithArgs(t *testing.T) {
	cfg := &options{}
	setDefaultOptions(cfg)
	l := &logger{opt: cfg}

	t.Run("Non Empty Args", func(t *testing.T) {
		k, v := l.withArgs([]driver.Value{1})()
		assert.Equal(t, cfg.sqlArgsFieldname, k)
		assert.Equal(t, []interface{}{1}, v)
	})

	t.Run("Non Empty Named Args", func(t *testing.T) {
		k, v := l.withArgs(namedValuesToValues([]driver.NamedValue{
			{"test", 1, 9},
		}))()
		assert.Equal(t, cfg.sqlArgsFieldname, k)
		assert.Equal(t, []interface{}{9}, v)
	})

	t.Run("Empty Args", func(t *testing.T) {
		k, v := l.withArgs([]driver.Value{})()
		assert.Equal(t, cfg.sqlArgsFieldname, k)
		assert.Equal(t, nil, v)
	})
}

func TestLogInternalWithMinimumLevel(t *testing.T) {
	cfg := &options{}
	setDefaultOptions(cfg)
	WithMinimumLevel(LevelError)(cfg)
	bl := &bufferTestLogger{}
	l := &logger{opt: cfg, logger: bl}
	l.log(context.TODO(), LevelDebug, "msg", time.Now(), nil)
	assert.Equal(t, 0, len(bl.Bytes()))
	bl.Reset()
}

func TestLogInternal(t *testing.T) {
	cfg := &options{}
	setDefaultOptions(cfg)
	bl := &bufferTestLogger{}
	l := &logger{opt: cfg, logger: bl}
	l.log(context.TODO(), LevelInfo, "msg", time.Now(), nil)

	var content bufLog
	err := json.Unmarshal(bl.Bytes(), &content)
	assert.NoError(t, err)
	assert.Contains(t, content.Data, cfg.timeFieldname)
	assert.Contains(t, content.Data, cfg.durationFieldname)
	assert.Equal(t, LevelInfo.String(), content.Level)
	bl.Reset()
}

func TestLogInternalWithData(t *testing.T) {
	cfg := &options{}
	setDefaultOptions(cfg)
	bl := &bufferTestLogger{}
	l := &logger{opt: cfg, logger: bl}
	l.log(context.TODO(), LevelInfo, "msg", time.Now(), nil, l.withQuery("query"))

	var content bufLog
	err := json.Unmarshal(bl.Bytes(), &content)
	assert.NoError(t, err)
	assert.Contains(t, content.Data, cfg.timeFieldname)
	assert.Contains(t, content.Data, cfg.durationFieldname)
	assert.Contains(t, content.Data, cfg.sqlQueryFieldname)
	assert.Equal(t, LevelInfo.String(), content.Level)
	assert.Equal(t, "msg", content.Message)
	assert.Equal(t, "query", content.Data[cfg.sqlQueryFieldname])
	bl.Reset()
}

func TestLogInternalErrorLevel(t *testing.T) {
	cfg := &options{}
	setDefaultOptions(cfg)
	bl := &bufferTestLogger{}
	l := &logger{opt: cfg, logger: bl}
	l.log(context.TODO(), LevelError, "msg", time.Now(), fmt.Errorf("dummy"), l.withQuery("query"))

	var content bufLog
	err := json.Unmarshal(bl.Bytes(), &content)
	assert.NoError(t, err)
	assert.Contains(t, content.Data, cfg.errorFieldname)
	assert.Contains(t, content.Data, cfg.sqlQueryFieldname)
	assert.Contains(t, content.Data, cfg.timeFieldname)
	assert.Contains(t, content.Data, cfg.durationFieldname)
	bl.Reset()
}

func TestLogTrimStringArgs(t *testing.T) {
	cfg := &options{}
	setDefaultOptions(cfg)

	longArgVal := "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua."
	bl := &bufferTestLogger{}
	l := &logger{opt: cfg, logger: bl}
	l.log(
		context.TODO(),
		LevelInfo,
		"msg",
		time.Now(),
		nil,
		l.withUID(cfg.stmtIDFieldname, ""),
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
	assert.Contains(t, content.Data, cfg.timeFieldname)
	assert.Contains(t, content.Data, cfg.durationFieldname)
	assert.Contains(t, content.Data, cfg.sqlArgsFieldname)
	assert.NotContains(t, content.Data, cfg.stmtIDFieldname)
	trimmedArg, ok := content.Data[cfg.sqlArgsFieldname].([]interface{})
	assert.True(t, ok)
	assert.Equal(t,
		fmt.Sprintf("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do  (%d bytes truncated)", len(longArgVal)-maxArgValueLen),
		trimmedArg[0],
	)
	assert.Equal(t,
		fmt.Sprintf("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do  (%d bytes truncated)", len(longArgVal)-maxArgValueLen),
		trimmedArg[1],
	)
	assert.Equal(t,
		"short",
		trimmedArg[2],
	)
	bl.Reset()
}

func TestWithLogArgumentsFalse(t *testing.T) {
	cfg := &options{}
	setDefaultOptions(cfg)
	WithLogArguments(false)(cfg)

	bl := &bufferTestLogger{}
	l := &logger{opt: cfg, logger: bl}
	l.log(
		context.TODO(),
		LevelInfo,
		"msg",
		time.Now(),
		nil,
		l.withQuery("query"),
		l.withArgs([]driver.Value{
			1,
			[]byte("kedua"),
			[]byte("lanjut"),
		}),
	)

	var content bufLog
	err := json.Unmarshal(bl.Bytes(), &content)
	assert.NoError(t, err)
	assert.Contains(t, content.Data, cfg.sqlQueryFieldname)
	assert.Contains(t, content.Data, cfg.timeFieldname)
	assert.Contains(t, content.Data, cfg.durationFieldname)
	// sql args should not logged
	assert.NotContains(t, content.Data, cfg.sqlArgsFieldname)
	bl.Reset()
}

func TestWithEmptyArgs(t *testing.T) {
	cfg := &options{}
	setDefaultOptions(cfg)

	bl := &bufferTestLogger{}
	l := &logger{opt: cfg, logger: bl}
	l.log(
		context.TODO(),
		LevelInfo,
		"msg",
		time.Now(),
		nil,
		l.withQuery("query"),
		l.withArgs([]driver.Value{}),
	)

	var content bufLog
	err := json.Unmarshal(bl.Bytes(), &content)
	assert.NoError(t, err)
	assert.Contains(t, content.Data, cfg.sqlQueryFieldname)
	assert.Contains(t, content.Data, cfg.timeFieldname)
	assert.Contains(t, content.Data, cfg.durationFieldname)
	// empty args will not logged
	assert.NotContains(t, content.Data, cfg.sqlArgsFieldname)
}

func TestWithErrorDriverSkip(t *testing.T) {
	cfg := &options{}
	setDefaultOptions(cfg)
	bl := &bufferTestLogger{}
	l := &logger{opt: cfg, logger: bl}

	t.Run("Skip", func(t *testing.T) {
		l.log(
			context.TODO(),
			LevelError,
			"msg",
			time.Now(),
			driver.ErrSkip,
		)

		assert.Empty(t, bl.Bytes())
	})

	t.Run("No Skip", func(t *testing.T) {
		WithLogDriverErrorSkip(true)(cfg)

		l.log(
			context.TODO(),
			LevelError,
			"msg",
			time.Now(),
			driver.ErrSkip,
		)

		var content bufLog
		err := json.Unmarshal(bl.Bytes(), &content)
		assert.NoError(t, err)
		assert.Contains(t, content.Data, cfg.timeFieldname)
		assert.Contains(t, content.Data, cfg.durationFieldname)
		assert.Contains(t, content.Data, cfg.errorFieldname)
	})
}

func TestWithSQLQueryAsMessage2(t *testing.T) {
	cfg := &options{}
	setDefaultOptions(cfg)
	bl := &bufferTestLogger{}
	l := &logger{opt: cfg, logger: bl}

	WithSQLQueryAsMessage(true)(cfg)

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
	assert.NotContains(t, content.Data, cfg.sqlQueryFieldname)
	assert.Equal(t, "query", content.Message)
	assert.Contains(t, content.Data, cfg.timeFieldname)
	assert.Contains(t, content.Data, cfg.durationFieldname)
	assert.Contains(t, content.Data, cfg.stmtIDFieldname)
	// empty args will not logged
	assert.NotContains(t, content.Data, cfg.sqlArgsFieldname)
}

type bufferTestLogger struct {
	bytes.Buffer
}

type bufLog struct {
	Level   string                 `json:"level"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

func (bl *bufferTestLogger) Log(_ context.Context, level Level, msg string, data map[string]interface{}) {
	bl.Reset()
	_ = json.NewEncoder(bl).Encode(bufLog{level.String(), msg, data})
}
