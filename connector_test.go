package sqldblogger

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConnector_Connect(t *testing.T) {

	t.Run("Connect Error", func(t *testing.T) {
		ml := newMockLogger()

		mockDriver := &driverMock{}
		mockDriver.On("Open", mock.Anything).Return(&driverConnMock{}, driver.ErrBadConn)

		con := &connector{dsn: "test", driver: mockDriver, logger: ml.testLogger}
		_, err := con.Connect(context.TODO())
		assert.Error(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "Connect", output.Message)
		assert.Equal(t, LevelError.String(), output.Level)
		assert.NotEmpty(t, output.Data[ml.testOpts.connIDFieldname])
	})

	t.Run("Connect Success", func(t *testing.T) {
		ml := newMockLogger()

		mockDriver := &driverMock{}
		mockDriver.On("Open", mock.Anything).Return(&driverConnMock{}, nil)

		con := &connector{dsn: "test", driver: mockDriver, logger: ml.testLogger}
		_, err := con.Connect(context.TODO())
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "Connect", output.Message)
		assert.Equal(t, LevelDebug.String(), output.Level)
		assert.NotEmpty(t, output.Data[ml.testOpts.connIDFieldname])
	})
}

func TestConnector_Driver(t *testing.T) {
	ml := newMockLogger()

	mockDriver := &driverMock{}
	con := &connector{dsn: "test", driver: mockDriver, logger: ml.testLogger}
	drv := con.Driver()

	assert.Equal(t, mockDriver, drv)
}
