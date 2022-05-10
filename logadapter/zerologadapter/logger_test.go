package zerologadapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	sqldblogger "github.com/simukti/sqldb-logger"
)

type logContent struct {
	Level    string        `json:"level"`
	Time     int64         `json:"time"`
	Duration float64       `json:"duration"`
	Query    string        `json:"query"`
	Args     []interface{} `json:"args"`
	Error    string        `json:"error"`
}

func TestZerologAdapter_Log(t *testing.T) {
	now := time.Now()
	wr := &bytes.Buffer{}
	lg := New(zerolog.New(wr))
	lvls := map[sqldblogger.Level]string{
		sqldblogger.LevelError: "error",
		sqldblogger.LevelInfo:  "info",
		sqldblogger.LevelDebug: "debug",
		sqldblogger.LevelTrace: "trace",
		sqldblogger.Level(99):  "debug", // unknown
	}

	for lvl, lvlStr := range lvls {
		data := map[string]interface{}{
			"time":     now.Unix(),
			"duration": time.Since(now).Nanoseconds(),
			"query":    "SELECT at.* FROM a_table AS at WHERE a.id = ? LIMIT 1",
			"args":     []interface{}{1},
		}

		if lvl == sqldblogger.LevelError {
			data["error"] = fmt.Errorf("dummy error").Error()
		}

		lg.Log(context.TODO(), lvl, "query", data)

		var content logContent

		err := json.Unmarshal(wr.Bytes(), &content)
		assert.NoError(t, err)
		assert.Equal(t, now.Unix(), content.Time)
		assert.True(t, content.Duration > 0)
		assert.Equal(t, lvlStr, content.Level)
		assert.Equal(t, "SELECT at.* FROM a_table AS at WHERE a.id = ? LIMIT 1", content.Query)
		if lvl == sqldblogger.LevelError {
			assert.Equal(t, "dummy error", content.Error)
		}

		wr.Reset()
	}
}
