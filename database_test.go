// +build unit

package sqldblogger

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	sql.Register("mock", &driverMock{})
}

func TestOpenDriver(t *testing.T) {
	t.Run("Without Options", func(t *testing.T) {
		mockDriver := &driverMock{}
		mockDriver.On("Open", mock.Anything).Return(&driverConnMock{}, nil)

		db, err := OpenDriver("test", mockDriver, bufLogger)
		assert.NoError(t, err)
		_, ok := interface{}(db).(*sql.DB)
		assert.True(t, ok)
	})

	t.Run("With Options", func(t *testing.T) {
		mockDriver := &driverMock{}
		mockDriver.On("Open", mock.Anything).Return(&driverConnMock{}, driver.ErrBadConn)

		db, err := OpenDriver("test", mockDriver, bufLogger, WithErrorFieldname("errtest"), WithMinimumLevel(LevelDebug))
		assert.NoError(t, err)
		_, ok := interface{}(db).(*sql.DB)
		assert.True(t, ok)
		err = db.Ping()
		assert.Error(t, err)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "Connect", output.Message)
		assert.Equal(t, LevelError.String(), output.Level)
		assert.Contains(t, output.Data, "errtest")
	})
}

type driverMock struct {
	mock.Mock
}

func (m *driverMock) Open(name string) (driver.Conn, error) {
	arg := m.Called(name)

	return arg.Get(0).(driver.Conn), arg.Error(1)
}
