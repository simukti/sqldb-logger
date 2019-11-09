// +build unit

package sqldblogger

import (
	"context"
	"database/sql/driver"

	"github.com/stretchr/testify/mock"
)

type statementMock struct {
	mock.Mock
}

func (m *statementMock) Close() error {
	return m.Called().Error(0)
}
func (m *statementMock) NumInput() int {
	return m.Called().Int(0)
}
func (m *statementMock) Exec(args []driver.Value) (driver.Result, error) {
	arg := m.Called(args)

	return arg.Get(0).(driver.Result), arg.Error(1)
}

func (m *statementMock) Query(args []driver.Value) (driver.Rows, error) {
	arg := m.Called(args)

	return arg.Get(0).(driver.Rows), arg.Error(1)
}

func (m *statementMock) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	arg := m.Called(ctx, args)

	return arg.Get(0).(driver.Result), arg.Error(1)
}

func (m *statementMock) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	arg := m.Called(ctx, args)

	return arg.Get(0).(driver.Rows), arg.Error(1)
}

func (m *statementMock) CheckNamedValue(nm *driver.NamedValue) error {
	return m.Called().Error(0)
}

func (m *statementMock) ColumnConverter(idx int) driver.ValueConverter {
	arg := m.Called(idx)

	return arg.Get(0).(driver.ValueConverter)
}
