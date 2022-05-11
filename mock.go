package sqldblogger

import (
	"bytes"
	"context"
	"encoding/json"
)

type mockLogger struct {
	testOpts   *options
	bufLogger  *bufferTestLogger
	testLogger *logger
}

func newMockLogger(o ...*options) *mockLogger {
	opt := &options{}
	if len(o) != 0 {
		opt = o[0]
	} else {
		setDefaultOptions(opt)
		opt.minimumLogLevel = LevelTrace
	}
	bufLogger := newBufferTestLogger()
	return &mockLogger{
		testOpts:  opt,
		bufLogger: bufLogger,
		testLogger: &logger{
			logger: bufLogger,
			opt:    opt,
		},
	}
}

type bufLog struct {
	Level   string                 `json:"level"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

func newBufferTestLogger() *bufferTestLogger {
	bufLogger := &bufferTestLogger{}
	json.NewEncoder(bufLogger).Encode(bufLog{})
	return bufLogger
}

type bufferTestLogger struct {
	bytes.Buffer
}

func (bl *bufferTestLogger) Log(_ context.Context, level Level, msg string, data map[string]interface{}) {
	bl.Reset()
	_ = json.NewEncoder(bl).Encode(bufLog{level.String(), msg, data})
}
