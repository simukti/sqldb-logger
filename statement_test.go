package sqldblogger

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStatement_Close(t *testing.T) {

	t.Run("Error", func(t *testing.T) {
		ml := newMockLogger()
		q := "SELECT * FROM tt WHERE id = ?"
		stmtMock := &statementMock{}
		stmtMock.On("Close").Return(driver.ErrBadConn)

		stmt := &statement{query: q, Stmt: stmtMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID(), connID: ml.testLogger.opt.uidGenerator.UniqueID()}
		err := stmt.Close()
		assert.Error(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "StmtClose", output.Message)
		assert.Equal(t, LevelError.String(), output.Level)
		assert.Equal(t, driver.ErrBadConn.Error(), output.Data[ml.testOpts.errorFieldname])
		assert.Equal(t, q, output.Data[ml.testOpts.sqlQueryFieldname])
		assert.Equal(t, stmt.connID, output.Data[ml.testOpts.connIDFieldname])
		assert.Equal(t, stmt.id, output.Data[ml.testOpts.stmtIDFieldname])
	})

	t.Run("Success", func(t *testing.T) {
		ml := newMockLogger()
		q := "SELECT * FROM tt WHERE id = ?"
		stmtMock := &statementMock{}
		stmtMock.On("Close").Return(nil)

		stmt := &statement{query: q, Stmt: stmtMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID(), connID: ml.testLogger.opt.uidGenerator.UniqueID()}
		err := stmt.Close()
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, false, isAbleToPrint(ml.testOpts, MessageStmtClose, LevelDebug))
	})
}

func TestStatement_NumInput(t *testing.T) {
	ml := newMockLogger()

	q := "SELECT * FROM tt WHERE id = ?"
	stmtMock := &statementMock{}
	stmtMock.On("NumInput").Return(1)

	stmt := &statement{query: q, Stmt: stmtMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID(), connID: ml.testLogger.opt.uidGenerator.UniqueID()}
	input := stmt.NumInput()
	assert.Equal(t, 1, input)
}

func TestStatement_Exec(t *testing.T) {

	t.Run("Error", func(t *testing.T) {
		ml := newMockLogger()

		q := "SELECT * FROM tt WHERE id = ?"
		stmtMock := &statementMock{}
		stmtMock.On("Exec", mock.Anything).Return(driver.ResultNoRows, driver.ErrBadConn)

		stmt := &statement{query: q, Stmt: stmtMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID(), connID: ml.testLogger.opt.uidGenerator.UniqueID()}
		_, err := stmt.Exec([]driver.Value{"testid"})
		assert.Error(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "StmtExec", output.Message)
		assert.Equal(t, LevelError.String(), output.Level)
		assert.Equal(t, driver.ErrBadConn.Error(), output.Data[ml.testOpts.errorFieldname])
		assert.Equal(t, q, output.Data[ml.testOpts.sqlQueryFieldname])
		assert.Equal(t, []interface{}{"testid"}, output.Data[ml.testOpts.sqlArgsFieldname])
	})

	t.Run("Success", func(t *testing.T) {
		ml := newMockLogger()

		q := "SELECT * FROM tt WHERE id = ?"
		stmtMock := &statementMock{}
		stmtMock.On("Exec", mock.Anything).Return(&resultMock{}, nil)

		stmt := &statement{query: q, Stmt: stmtMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID(), connID: ml.testLogger.opt.uidGenerator.UniqueID()}
		_, err := stmt.Exec([]driver.Value{"testid"})
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, false, isAbleToPrint(ml.testOpts, MessagePrepare, LevelDebug))
	})

	t.Run("Success With Custom Level", func(t *testing.T) {
		ml := newMockLogger()
		ml.testOpts.execerLevel = LevelDebug

		q := "SELECT * FROM tt WHERE id = ?"
		stmtMock := &statementMock{}
		stmtMock.On("Exec", mock.Anything).Return(&resultMock{}, nil)

		stmt := &statement{query: q, Stmt: stmtMock, logger: ml.testLogger, id: ml.testOpts.uidGenerator.UniqueID(), connID: ml.testOpts.uidGenerator.UniqueID()}
		_, err := stmt.Exec([]driver.Value{"testid"})
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, false, isAbleToPrint(ml.testOpts, MessageStmtExec, LevelDebug))
	})
}

func TestStatement_Query(t *testing.T) {

	t.Run("Error", func(t *testing.T) {
		ml := newMockLogger()

		q := "SELECT * FROM tt WHERE id = ?"
		stmtMock := &statementMock{}
		stmtMock.On("Query", mock.Anything).Return(&rowsMock{}, driver.ErrBadConn)

		stmt := &statement{query: q, Stmt: stmtMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID(), connID: ml.testLogger.opt.uidGenerator.UniqueID()}
		_, err := stmt.Query([]driver.Value{"testid"})
		assert.Error(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "StmtQuery", output.Message)
		assert.Equal(t, LevelError.String(), output.Level)
		assert.Equal(t, driver.ErrBadConn.Error(), output.Data[ml.testOpts.errorFieldname])
		assert.Equal(t, q, output.Data[ml.testOpts.sqlQueryFieldname])
		assert.Equal(t, []interface{}{"testid"}, output.Data[ml.testOpts.sqlArgsFieldname])
	})

	t.Run("Success", func(t *testing.T) {
		ml := newMockLogger()

		q := "SELECT * FROM tt WHERE id = ?"
		stmtMock := &statementMock{}
		stmtMock.On("Query", mock.Anything).Return(&rowsMock{}, nil)

		stmt := &statement{query: q, Stmt: stmtMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID(), connID: ml.testLogger.opt.uidGenerator.UniqueID()}
		_, err := stmt.Query([]driver.Value{"testid"})
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, false, isAbleToPrint(ml.testOpts, MessagePrepare, LevelDebug))
	})

	t.Run("Success With Custom Level", func(t *testing.T) {
		ml := newMockLogger()
		ml.testOpts.queryerLevel = LevelDebug

		q := "SELECT * FROM tt WHERE id = ?"
		stmtMock := &statementMock{}
		stmtMock.On("Query", mock.Anything).Return(&rowsMock{}, nil)

		stmt := &statement{query: q, Stmt: stmtMock, logger: ml.testLogger, id: ml.testOpts.uidGenerator.UniqueID(), connID: ml.testOpts.uidGenerator.UniqueID()}
		_, err := stmt.Query([]driver.Value{"testid"})
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, false, isAbleToPrint(ml.testOpts, MessageStmtQuery, LevelDebug))
	})
}

func TestStatement_ExecContext(t *testing.T) {

	t.Run("Not implement driver.StmtExecContext", func(t *testing.T) {
		ml := newMockLogger()

		q := "SELECT * FROM tt WHERE id = ?"
		stmtMock := &statementMock{}
		stmt := &statement{query: q, Stmt: stmtMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID(), connID: ml.testLogger.opt.uidGenerator.UniqueID()}

		_, err := stmt.ExecContext(context.TODO(), []driver.NamedValue{{Name: "", Ordinal: 0, Value: "testid"}})
		assert.Error(t, err)
		assert.Equal(t, driver.ErrSkip, err)
	})

	t.Run("Error", func(t *testing.T) {
		ml := newMockLogger()

		q := "SELECT * FROM tt WHERE id = ?"
		stmtMock := &statementExecerContextMock{}
		stmtMock.On("ExecContext", mock.Anything, mock.Anything).Return(&resultMock{}, driver.ErrBadConn)

		stmt := &statement{query: q, Stmt: stmtMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID(), connID: ml.testLogger.opt.uidGenerator.UniqueID()}
		_, err := stmt.ExecContext(context.TODO(), []driver.NamedValue{{Name: "", Ordinal: 0, Value: "testid"}})
		assert.Error(t, err)
		assert.Equal(t, driver.ErrBadConn, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "StmtExecContext", output.Message)
		assert.Equal(t, LevelError.String(), output.Level)
		assert.Equal(t, driver.ErrBadConn.Error(), output.Data[ml.testOpts.errorFieldname])
		assert.Equal(t, q, output.Data[ml.testOpts.sqlQueryFieldname])
		assert.Equal(t, []interface{}{"testid"}, output.Data[ml.testOpts.sqlArgsFieldname])
	})

	t.Run("Success", func(t *testing.T) {
		ml := newMockLogger()

		q := "SELECT * FROM tt WHERE id = ?"
		stmtMock := &statementExecerContextMock{}
		stmtMock.On("ExecContext", mock.Anything, mock.Anything).Return(&resultMock{}, nil)

		stmt := &statement{query: q, Stmt: stmtMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID(), connID: ml.testLogger.opt.uidGenerator.UniqueID()}
		_, err := stmt.ExecContext(context.TODO(), []driver.NamedValue{{Name: "", Ordinal: 0, Value: "testid"}})
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, false, isAbleToPrint(ml.testOpts, MessageStmtExecContext, LevelDebug))
	})

	t.Run("Success With Custom Level", func(t *testing.T) {
		ml := newMockLogger()
		ml.testOpts.execerLevel = LevelDebug

		q := "SELECT * FROM tt WHERE id = ?"
		stmtMock := &statementExecerContextMock{}
		stmtMock.On("ExecContext", mock.Anything, mock.Anything).Return(&resultMock{}, nil)

		stmt := &statement{query: q, Stmt: stmtMock, logger: ml.testLogger, id: ml.testOpts.uidGenerator.UniqueID(), connID: ml.testOpts.uidGenerator.UniqueID()}
		_, err := stmt.ExecContext(context.TODO(), []driver.NamedValue{{Name: "", Ordinal: 0, Value: "testid"}})
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, false, isAbleToPrint(ml.testOpts, MessageStmtExecContext, LevelDebug))
	})
}

func TestStatement_QueryContext(t *testing.T) {

	t.Run("Not implement driver.StmtQueryContext", func(t *testing.T) {
		ml := newMockLogger()

		q := "SELECT * FROM tt WHERE id = ?"
		stmtMock := &statementMock{}
		stmt := &statement{query: q, Stmt: stmtMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID(), connID: ml.testLogger.opt.uidGenerator.UniqueID()}

		_, err := stmt.QueryContext(context.TODO(), []driver.NamedValue{{Name: "", Ordinal: 0, Value: "testid"}})
		assert.Error(t, err)
		assert.Equal(t, driver.ErrSkip, err)
	})

	t.Run("Error", func(t *testing.T) {
		ml := newMockLogger()

		q := "SELECT * FROM tt WHERE id = ?"
		stmtMock := &statementQueryerContextMock{}
		stmtMock.On("QueryContext", mock.Anything, mock.Anything).Return(&rowsMock{}, driver.ErrBadConn)

		stmt := &statement{query: q, Stmt: stmtMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID(), connID: ml.testLogger.opt.uidGenerator.UniqueID()}
		_, err := stmt.QueryContext(context.TODO(), []driver.NamedValue{{Name: "", Ordinal: 0, Value: "testid"}})
		assert.Error(t, err)
		assert.Equal(t, driver.ErrBadConn, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "StmtQueryContext", output.Message)
		assert.Equal(t, LevelError.String(), output.Level)
		assert.Equal(t, driver.ErrBadConn.Error(), output.Data[ml.testOpts.errorFieldname])
		assert.Equal(t, q, output.Data[ml.testOpts.sqlQueryFieldname])
		assert.Equal(t, []interface{}{"testid"}, output.Data[ml.testOpts.sqlArgsFieldname])
	})

	t.Run("Success", func(t *testing.T) {
		ml := newMockLogger()

		q := "SELECT * FROM tt WHERE id = ?"
		stmtMock := &statementQueryerContextMock{}
		stmtMock.On("QueryContext", mock.Anything, mock.Anything).Return(&rowsMock{}, nil)

		stmt := &statement{query: q, Stmt: stmtMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID(), connID: ml.testLogger.opt.uidGenerator.UniqueID()}
		_, err := stmt.QueryContext(context.TODO(), []driver.NamedValue{{Name: "", Ordinal: 0, Value: "testid"}})
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, false, isAbleToPrint(ml.testOpts, MessagePrepare, LevelDebug))
	})

	t.Run("Success With Custom Level", func(t *testing.T) {
		ml := newMockLogger()
		ml.testOpts.queryerLevel = LevelDebug

		q := "SELECT * FROM tt WHERE id = ?"
		stmtMock := &statementQueryerContextMock{}
		stmtMock.On("QueryContext", mock.Anything, mock.Anything).Return(&rowsMock{}, nil)

		stmt := &statement{query: q, Stmt: stmtMock, logger: ml.testLogger, id: ml.testOpts.uidGenerator.UniqueID(), connID: ml.testOpts.uidGenerator.UniqueID()}
		_, err := stmt.QueryContext(context.TODO(), []driver.NamedValue{{Name: "", Ordinal: 0, Value: "testid"}})
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, false, isAbleToPrint(ml.testOpts, MessageStmtQueryContext, LevelDebug))
	})
}

func TestStatement_QueryContext2(t *testing.T) {
	ml := newMockLogger()

	// make sure conn id flow into statement
	driverConnMock := &driverConnMock{}
	stmtMock := &statementMock{}
	stmtMock.On("Query", mock.Anything).Return(&rowsMock{}, nil)
	driverConnMock.On("Prepare", mock.Anything).Return(stmtMock, nil)
	q := "SELECT * FROM tt WHERE id = ?"
	conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID()}
	stmt, err := conn.Prepare(q)
	assert.NoError(t, err)

	var connOutput bufLog
	err = json.Unmarshal(ml.bufLogger.Bytes(), &connOutput)
	assert.NoError(t, err)
	assert.Equal(t, false, isAbleToPrint(ml.testOpts, MessagePrepare, LevelDebug))

	_, rsErr := stmt.Query([]driver.Value{1})
	assert.NoError(t, rsErr)
	var stmtOutput bufLog
	err = json.Unmarshal(ml.bufLogger.Bytes(), &stmtOutput)
	assert.NoError(t, err)
	assert.Equal(t, false, isAbleToPrint(ml.testOpts, MessagePrepare, LevelDebug))
}

func TestStatement_CheckNamedValue(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		ml := newMockLogger()

		stmtMock := &statementNamedValueCheckerMock{}
		stmtMock.On("CheckNamedValue", mock.Anything).Return(driver.ErrBadConn)

		stmt := &statement{Stmt: stmtMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID(), connID: ml.testLogger.opt.uidGenerator.UniqueID()}
		err := stmt.CheckNamedValue(&driver.NamedValue{Name: "", Ordinal: 0, Value: "testid"})
		assert.Error(t, err)

		var stmtOutput bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &stmtOutput)
		assert.NoError(t, err)
		assert.Equal(t, LevelError.String(), stmtOutput.Level)
		assert.Equal(t, "StmtCheckNamedValue", stmtOutput.Message)
		assert.NotEmpty(t, stmtOutput.Data[ml.testOpts.stmtIDFieldname])
		assert.NotEmpty(t, stmtOutput.Data[ml.testOpts.connIDFieldname])
	})

	t.Run("Not implement driver.NamedValueChecker", func(t *testing.T) {
		ml := newMockLogger()

		stmtMock := &statementMock{}

		stmt := &statement{Stmt: stmtMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID(), connID: ml.testLogger.opt.uidGenerator.UniqueID()}
		err := stmt.CheckNamedValue(&driver.NamedValue{Name: "", Ordinal: 0, Value: "testid"})
		assert.Error(t, err)
		assert.Equal(t, driver.ErrSkip, err)
	})
}

func TestStatement_ColumnConverter(t *testing.T) {

	t.Run("Return as is", func(t *testing.T) {
		ml := newMockLogger()

		stmtMock := &statementValueConverterMock{}
		stmtMock.On("ColumnConverter", mock.Anything).Return(driver.NotNull{Converter: driver.DefaultParameterConverter})

		stmt := &statement{Stmt: stmtMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID(), connID: ml.testLogger.opt.uidGenerator.UniqueID()}
		cnv := stmt.ColumnConverter(1)
		val, err := cnv.ConvertValue(1)
		assert.NoError(t, err)
		intVal, ok := val.(int64)
		assert.True(t, ok)
		assert.Equal(t, int64(1), intVal)
	})

	t.Run("Not implement driver.ColumnConverter", func(t *testing.T) {
		ml := newMockLogger()

		stmtMock := &statementMock{}
		stmt := &statement{Stmt: stmtMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID(), connID: ml.testLogger.opt.uidGenerator.UniqueID()}
		cnv := stmt.ColumnConverter(1)
		assert.Equal(t, driver.DefaultParameterConverter, cnv)
	})
}

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

type statementExecerContextMock struct {
	statementMock
}

func (m *statementExecerContextMock) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	arg := m.Called(ctx, args)

	return arg.Get(0).(driver.Result), arg.Error(1)
}

type statementQueryerContextMock struct {
	statementMock
}

func (m *statementQueryerContextMock) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	arg := m.Called(ctx, args)

	return arg.Get(0).(driver.Rows), arg.Error(1)
}

type statementNamedValueCheckerMock struct {
	statementMock
}

func (m *statementNamedValueCheckerMock) CheckNamedValue(nm *driver.NamedValue) error {
	return m.Called().Error(0)
}

type statementValueConverterMock struct {
	statementMock
}

func (m *statementValueConverterMock) ColumnConverter(idx int) driver.ValueConverter {
	return m.Called(idx).Get(0).(driver.ValueConverter)
}
