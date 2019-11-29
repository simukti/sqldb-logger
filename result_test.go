// +build unit

package sqldblogger

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestResult_LastInsertId(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		resMock := &resultMock{}
		resMock.On("LastInsertId").Return(0, errors.New("dummy"))
		r := &result{Result: resMock, logger: testLogger, connID: uniqueID(), query: "SELECT 1"}
		id, err := r.LastInsertId()
		assert.Equal(t, int64(0), id)
		assert.Error(t, err)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, LevelError.String(), output.Level)
		assert.Equal(t, "dummy", output.Data[testOpts.errorFieldname])
		assert.NotEmpty(t, output.Data[connID])
		assert.NotEmpty(t, output.Data[testOpts.sqlQueryFieldname])
		bufLogger.Reset()
	})

	t.Run("Success", func(t *testing.T) {
		resMock := &resultMock{}
		resMock.On("LastInsertId").Return(1, nil)
		r := &result{Result: resMock, logger: testLogger, connID: uniqueID(), query: "SELECT 1"}
		id, err := r.LastInsertId()
		assert.Equal(t, int64(1), id)
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.Error(t, err)
		bufLogger.Reset()
	})
}

func TestResult_RowsAffected(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		resMock := &resultMock{}
		resMock.On("RowsAffected").Return(0, errors.New("dummy"))
		r := &result{Result: resMock, logger: testLogger, connID: uniqueID(), query: "SELECT 1"}
		id, err := r.RowsAffected()
		assert.Equal(t, int64(0), id)
		assert.Error(t, err)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, LevelError.String(), output.Level)
		assert.Equal(t, "dummy", output.Data[testOpts.errorFieldname])
		assert.NotEmpty(t, output.Data[connID])
		assert.NotEmpty(t, output.Data[testOpts.sqlQueryFieldname])
		bufLogger.Reset()
	})

	t.Run("Success", func(t *testing.T) {
		resMock := &resultMock{}
		resMock.On("RowsAffected").Return(1, nil)
		r := &result{Result: resMock, logger: testLogger, connID: uniqueID()}
		id, err := r.RowsAffected()
		assert.Equal(t, int64(1), id)
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.Error(t, err)
		bufLogger.Reset()
	})
}

type resultMock struct {
	mock.Mock
}

func (m *resultMock) LastInsertId() (int64, error) {
	arg := m.Called()

	return int64(arg.Int(0)), arg.Error(1)
}

func (m *resultMock) RowsAffected() (int64, error) {
	arg := m.Called()

	return int64(arg.Int(0)), arg.Error(1)
}
