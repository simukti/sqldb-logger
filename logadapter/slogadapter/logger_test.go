package slogadapter

import (
	"context"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	sqldblogger "github.com/simukti/sqldb-logger"
)

// A TestHandler is an slog.Handler that simply records the latest record,
// which is used to verify the expected values provided by the sqldblogger.Logger.
type TestHandler struct {
	latestRecord slog.Record
}

func NewTestHandler() *TestHandler {
	return &TestHandler{}
}

// Enabled implements slog.Handler.
func (h *TestHandler) Enabled(_ context.Context, level slog.Level) bool {
	// All levels are always enabled.
	return true
}

// Handle implements slog.Handler.
func (h *TestHandler) Handle(_ context.Context, r slog.Record) error {
	// Simply store the latest record. We'll use it to verify expected
	// values.
	h.latestRecord = r
	return nil
}

// WithAttrs implements slog.Handler.
func (h *TestHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// Not needed by the adapter.
	return h
}

// WithGroup implements slog.Handler.
func (h *TestHandler) WithGroup(name string) slog.Handler {
	// Not needed by the adapter.
	return h
}

// Handler implements slog.Handler.
func (h *TestHandler) Handler() slog.Handler {
	// Not needed by the adapter.
	return h
}

func TestSlogAdapter_Log(t *testing.T) {
	testHandler := NewTestHandler()
	logger := New(slog.New(testHandler))

	levelMap := map[sqldblogger.Level]slog.Level{
		sqldblogger.LevelError: slog.LevelError,
		sqldblogger.LevelInfo:  slog.LevelInfo,
		sqldblogger.LevelDebug: slog.LevelDebug,
		sqldblogger.LevelTrace: slog.LevelDebug,
		sqldblogger.Level(99):  slog.LevelDebug, // unknown
	}

	now := time.Now()
	const queryStr = "SELECT at.* FROM a_table AS at WHERE a.id = ? LIMIT 1"

	for sqldbLevel, slogLevel := range levelMap {

		data := map[string]interface{}{
			"time":     now.Unix(),
			"duration": time.Since(now).Nanoseconds(),
			"query":    queryStr,
			"args":     []interface{}{1},
		}

		if sqldbLevel == sqldblogger.LevelError {
			data["error"] = fmt.Errorf("some error").Error()
		}

		// Log the message with associated data
		logger.Log(context.TODO(), sqldbLevel, "query msg", data)

		// Check expected values by inspecting the latest record
		// stored in the test handler.
		record := testHandler.latestRecord

		assert.Equal(t, "query msg", record.Message)
		assert.Equal(t, slogLevel, record.Level)
		assert.Equal(t, len(data), record.NumAttrs())

		record.Attrs(func(a slog.Attr) bool {
			switch a.Key {
			case "time":
				assert.Equal(t, now.Unix(), a.Value.Int64())
			case "query":
				assert.Equal(t, queryStr, a.Value.String())
			case "duration":
				assert.True(t, a.Value.Int64() > 0)
			case "args":
				assert.Equal(t, []interface{}{1}, a.Value.Any())
			case "error":
				assert.Equal(t, sqldblogger.LevelError, sqldbLevel)
				assert.Equal(t, "some error", a.Value.String())
			}

			return true
		})
	}
}
