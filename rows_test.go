package sqldblogger

import (
	"database/sql/driver"
	"encoding/json"
	"io"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRows_Columns(t *testing.T) {
	ml := newMockLogger()

	rowsMock := &rowsMock{}
	rowsMock.On("Columns").Return([]string{"a", "b"})
	rs := &rows{Rows: rowsMock, logger: ml.testLogger, connID: ml.testLogger.opt.uidGenerator.UniqueID()}

	cols := rs.Columns()
	assert.Implements(t, (*driver.Rows)(nil), rs)
	assert.Equal(t, cols, []string{"a", "b"})
}

func TestRows_Close(t *testing.T) {

	t.Run("Error", func(t *testing.T) {
		ml := newMockLogger()

		rowsMock := &rowsMock{}
		rowsMock.On("Close").Return(driver.ErrBadConn)
		rs := &rows{Rows: rowsMock, logger: ml.testLogger, connID: ml.testLogger.opt.uidGenerator.UniqueID(), stmtID: ml.testLogger.opt.uidGenerator.UniqueID(), query: "SELECT 1"}

		err := rs.Close()
		assert.Implements(t, (*driver.Rows)(nil), rs)
		assert.Error(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "RowsClose", output.Message)
		assert.Equal(t, LevelError.String(), output.Level)
		assert.NotEmpty(t, output.Data[ml.testOpts.connIDFieldname])
		assert.NotEmpty(t, output.Data[ml.testOpts.stmtIDFieldname])
		assert.NotEmpty(t, output.Data[ml.testOpts.sqlQueryFieldname])
		ml.bufLogger.Reset()
	})

	t.Run("Success", func(t *testing.T) {
		ml := newMockLogger()

		rowsMock := &rowsMock{}
		rowsMock.On("Close").Return(nil)
		rs := &rows{Rows: rowsMock, logger: ml.testLogger, connID: ml.testLogger.opt.uidGenerator.UniqueID()}

		err := rs.Close()
		assert.Implements(t, (*driver.Rows)(nil), rs)
		assert.NoError(t, err)
	})
}

func TestRows_Next(t *testing.T) {

	t.Run("Error io.EOF", func(t *testing.T) {
		ml := newMockLogger()

		rowsMock := &rowsMock{}
		rowsMock.On("Next", mock.Anything).Return(io.EOF)
		rs := &rows{Rows: rowsMock, logger: ml.testLogger, connID: ml.testLogger.opt.uidGenerator.UniqueID()}

		err := rs.Next([]driver.Value{1})
		assert.Implements(t, (*driver.Rows)(nil), rs)
		assert.Error(t, err)
		assert.Equal(t, io.EOF, err)
	})

	t.Run("Error Non-io.EOF With Dest Value", func(t *testing.T) {
		ml := newMockLogger()

		rowsMock := &rowsMock{}
		rowsMock.On("Next", mock.Anything).Return(driver.ErrBadConn)
		rs := &rows{Rows: rowsMock, logger: ml.testLogger, connID: ml.testLogger.opt.uidGenerator.UniqueID(), stmtID: ml.testLogger.opt.uidGenerator.UniqueID(), query: "SELECT 1"}

		err := rs.Next([]driver.Value{1})
		assert.Implements(t, (*driver.Rows)(nil), rs)
		assert.Error(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "RowsNext", output.Message)
		assert.Equal(t, LevelError.String(), output.Level)
		assert.NotEmpty(t, output.Data[ml.testOpts.connIDFieldname])
		assert.NotEmpty(t, output.Data[ml.testOpts.stmtIDFieldname])
		assert.NotEmpty(t, output.Data[ml.testOpts.sqlQueryFieldname])
		assert.NotEmpty(t, output.Data["rows_dest"])
		ml.bufLogger.Reset()
	})

	t.Run("Error Non-io.EOF Without Dest Value", func(t *testing.T) {
		ml := newMockLogger()
		ml.testOpts.logArgs = false

		rowsMock := &rowsMock{}
		rowsMock.On("Next", mock.Anything).Return(driver.ErrBadConn)

		rs := &rows{Rows: rowsMock, logger: ml.testLogger, connID: ml.testLogger.opt.uidGenerator.UniqueID(), stmtID: ml.testLogger.opt.uidGenerator.UniqueID(), query: "SELECT 1"}

		err := rs.Next([]driver.Value{1})
		assert.Implements(t, (*driver.Rows)(nil), rs)
		assert.Error(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "RowsNext", output.Message)
		assert.Equal(t, LevelError.String(), output.Level)
		assert.NotEmpty(t, output.Data[ml.testOpts.connIDFieldname])
		assert.NotEmpty(t, output.Data[ml.testOpts.stmtIDFieldname])
		assert.NotEmpty(t, output.Data[ml.testOpts.sqlQueryFieldname])
		assert.NotContains(t, output.Data, "rows_dest")
		ml.bufLogger.Reset()
	})

	t.Run("Success With Dest Value", func(t *testing.T) {
		ml := newMockLogger()
		ml.testOpts.minimumLogLevel = LevelTrace

		rowsMock := &rowsMock{}
		rowsMock.On("Next", mock.Anything).Return(nil)
		rs := &rows{Rows: rowsMock, logger: ml.testLogger, connID: ml.testLogger.opt.uidGenerator.UniqueID(), stmtID: ml.testLogger.opt.uidGenerator.UniqueID(), query: "SELECT 1"}

		err := rs.Next([]driver.Value{1})
		assert.Implements(t, (*driver.Rows)(nil), rs)
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "RowsNext", output.Message)
		assert.Equal(t, LevelTrace.String(), output.Level)
		assert.NotEmpty(t, output.Data[ml.testOpts.connIDFieldname])
		assert.NotEmpty(t, output.Data[ml.testOpts.stmtIDFieldname])
		assert.NotEmpty(t, output.Data[ml.testOpts.sqlQueryFieldname])
		assert.NotEmpty(t, output.Data["rows_dest"])
	})

	t.Run("Success Without Dest Value", func(t *testing.T) {
		ml := newMockLogger()
		ml.testOpts.minimumLogLevel = LevelTrace
		ml.testOpts.logArgs = false

		rowsMock := &rowsMock{}
		rowsMock.On("Next", mock.Anything).Return(nil)
		rs := &rows{Rows: rowsMock, logger: ml.testLogger, connID: ml.testLogger.opt.uidGenerator.UniqueID(), stmtID: ml.testLogger.opt.uidGenerator.UniqueID(), query: "SELECT 1"}

		err := rs.Next([]driver.Value{1})
		assert.Implements(t, (*driver.Rows)(nil), rs)
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "RowsNext", output.Message)
		assert.Equal(t, LevelTrace.String(), output.Level)
		assert.NotEmpty(t, output.Data[ml.testOpts.connIDFieldname])
		assert.NotEmpty(t, output.Data[ml.testOpts.stmtIDFieldname])
		assert.NotEmpty(t, output.Data[ml.testOpts.sqlQueryFieldname])
		assert.NotContains(t, output.Data, "rows_dest")
	})
}

func TestRows_HasNextResultSet(t *testing.T) {
	ml := newMockLogger()

	t.Run("Non driver.RowsNextResultSet", func(t *testing.T) {
		rowsMock := &rowsMock{}
		rs := &rows{Rows: rowsMock, logger: ml.testLogger, connID: ml.testLogger.opt.uidGenerator.UniqueID()}

		flag := rs.HasNextResultSet()
		assert.Equal(t, false, flag)
	})

	t.Run("driver.RowsNextResultSet", func(t *testing.T) {
		rowsMock := &rowsRowsNextResultSetMock{}
		rowsMock.On("HasNextResultSet").Return(true)
		rs := &rows{Rows: rowsMock, logger: ml.testLogger, connID: ml.testLogger.opt.uidGenerator.UniqueID()}

		flag := rs.HasNextResultSet()
		assert.Equal(t, true, flag)
	})
}

func TestRows_NextResultSet(t *testing.T) {

	t.Run("Non driver.RowsNextResultSet", func(t *testing.T) {
		ml := newMockLogger()

		rowsMock := &rowsMock{}
		rs := &rows{Rows: rowsMock, logger: ml.testLogger, connID: ml.testLogger.opt.uidGenerator.UniqueID()}

		err := rs.NextResultSet()
		assert.Error(t, err)
		assert.Equal(t, io.EOF, err)
	})

	t.Run("Error io.EOF", func(t *testing.T) {
		ml := newMockLogger()

		rowsMock := &rowsRowsNextResultSetMock{}
		rowsMock.On("NextResultSet").Return(io.EOF)
		rs := &rows{Rows: rowsMock, logger: ml.testLogger, connID: ml.testLogger.opt.uidGenerator.UniqueID()}

		err := rs.NextResultSet()
		assert.Error(t, err)
		assert.Equal(t, io.EOF, err)
	})

	t.Run("Not Error", func(t *testing.T) {
		ml := newMockLogger()
		ml.testOpts.minimumLogLevel = LevelTrace

		rowsMock := &rowsRowsNextResultSetMock{}
		rowsMock.On("NextResultSet").Return(nil)
		rs := &rows{Rows: rowsMock, logger: ml.testLogger, connID: ml.testLogger.opt.uidGenerator.UniqueID(), stmtID: ml.testLogger.opt.uidGenerator.UniqueID(), query: "SELECT 1"}

		err := rs.NextResultSet()
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "RowsNextResultSet", output.Message)
		assert.Equal(t, LevelTrace.String(), output.Level)
		assert.NotEmpty(t, output.Data[ml.testOpts.connIDFieldname])
		assert.NotEmpty(t, output.Data[ml.testOpts.stmtIDFieldname])
		assert.NotEmpty(t, output.Data[ml.testOpts.sqlQueryFieldname])
	})

	t.Run("Error Non io.EOF", func(t *testing.T) {
		ml := newMockLogger()

		rowsMock := &rowsRowsNextResultSetMock{}
		rowsMock.On("NextResultSet").Return(driver.ErrBadConn)
		rs := &rows{Rows: rowsMock, logger: ml.testLogger, connID: ml.testLogger.opt.uidGenerator.UniqueID(), stmtID: ml.testLogger.opt.uidGenerator.UniqueID(), query: "SELECT 1"}

		err := rs.NextResultSet()
		assert.Error(t, err)
		assert.NotEqual(t, io.EOF, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "RowsNextResultSet", output.Message)
		assert.Equal(t, LevelError.String(), output.Level)
		assert.NotEmpty(t, output.Data[ml.testOpts.connIDFieldname])
		assert.NotEmpty(t, output.Data[ml.testOpts.stmtIDFieldname])
		assert.NotEmpty(t, output.Data[ml.testOpts.sqlQueryFieldname])
	})
}

type rowsMock struct {
	mock.Mock
}

func (m *rowsMock) Columns() []string              { return m.Called().Get(0).([]string) }
func (m *rowsMock) Close() error                   { return m.Called().Error(0) }
func (m *rowsMock) Next(dest []driver.Value) error { return m.Called(dest).Error(0) }

type rowsRowsNextResultSetMock struct {
	rowsMock
}

func (m *rowsRowsNextResultSetMock) HasNextResultSet() bool { return m.Called().Get(0).(bool) }
func (m *rowsRowsNextResultSetMock) NextResultSet() error   { return m.Called().Error(0) }

type rowsRowsColumnTypeScanTypeMock struct {
	rowsMock
}

func (m *rowsRowsColumnTypeScanTypeMock) ColumnTypeScanType(index int) reflect.Type {
	return m.Called(index).Get(0).(reflect.Type)
}

type rowsRowsColumnTypeDatabaseTypeNameMock struct {
	rowsMock
}

func (m *rowsRowsColumnTypeDatabaseTypeNameMock) ColumnTypeDatabaseTypeName(index int) string {
	return m.Called(index).Get(0).(string)
}

type rowsRowsColumnTypeLengthMock struct {
	rowsMock
}

func (m *rowsRowsColumnTypeLengthMock) ColumnTypeLength(index int) (length int64, ok bool) {
	c := m.Called(index)

	return c.Get(0).(int64), c.Get(1).(bool)
}

type rowsRowsColumnTypeNullableMock struct {
	rowsMock
}

func (m *rowsRowsColumnTypeNullableMock) ColumnTypeNullable(index int) (nullable, ok bool) {
	c := m.Called(index)

	return c.Get(0).(bool), c.Get(1).(bool)
}

type rowsRowsColumnTypePrecisionScaleMock struct {
	rowsMock
}

func (m *rowsRowsColumnTypePrecisionScaleMock) ColumnTypePrecisionScale(index int) (precision, scale int64, ok bool) {
	c := m.Called(index)

	return c.Get(0).(int64), c.Get(1).(int64), c.Get(2).(bool)
}
