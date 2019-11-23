// +build unit

package sqldblogger

import (
	"database/sql/driver"
	"reflect"

	"github.com/stretchr/testify/mock"
)

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
