// +build unit

package sqldblogger

import (
	"database/sql/driver"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTransaction_Commit(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		txMock := &transactionMock{}
		txMock.On("Commit").Return(driver.ErrBadConn)

		conn := &transaction{tx: txMock, logger: testLogger}
		err := conn.Commit()
		assert.Error(t, err)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "Commit", output.Message)
		assert.Equal(t, LevelError.String(), output.Level)
	})

	t.Run("Success", func(t *testing.T) {
		txMock := &transactionMock{}
		txMock.On("Commit").Return(nil)

		conn := &transaction{tx: txMock, logger: testLogger}
		err := conn.Commit()
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "Commit", output.Message)
		assert.Equal(t, LevelDebug.String(), output.Level)
	})
}

func TestTransaction_Rollback(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		txMock := &transactionMock{}
		txMock.On("Rollback").Return(driver.ErrBadConn)

		conn := &transaction{tx: txMock, logger: testLogger}
		err := conn.Rollback()
		assert.Error(t, err)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "Rollback", output.Message)
		assert.Equal(t, LevelError.String(), output.Level)
	})

	t.Run("Success", func(t *testing.T) {
		txMock := &transactionMock{}
		txMock.On("Rollback").Return(nil)

		conn := &transaction{tx: txMock, logger: testLogger}
		err := conn.Rollback()
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "Rollback", output.Message)
		assert.Equal(t, LevelDebug.String(), output.Level)
	})
}

type transactionMock struct {
	mock.Mock
}

func (m *transactionMock) Commit() error {
	return m.Called().Error(0)
}

func (m *transactionMock) Rollback() error {
	return m.Called().Error(0)
}
