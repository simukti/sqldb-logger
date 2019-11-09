// +build unit

package sqldblogger

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	testConfig = &config{}
	bufLogger  = &bufferTestLogger{}
	testLogger *logger
)

func init() {
	setDefaultConfig(testConfig)
	testConfig.minimumLogLevel = LevelDebug
	testLogger = &logger{
		logger: bufLogger,
		cfg:    testConfig,
	}
}

func TestConnection_Begin(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		driverConnMock := &driverConnMock{}
		var txMock *transactionMock
		driverConnMock.On("Begin").Return(txMock, driver.ErrBadConn)

		conn := &connection{driverConn: driverConnMock, logger: testLogger}
		_, err := conn.Begin()
		assert.Error(t, err)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, LevelError.String(), output.Level)
	})

	t.Run("Success", func(t *testing.T) {
		driverConnMock := &driverConnMock{}
		txMock := &transactionMock{}
		driverConnMock.On("Begin").Return(txMock, nil)

		conn := &connection{driverConn: driverConnMock, logger: testLogger}
		tx, err := conn.Begin()
		assert.NoError(t, err)
		assert.Implements(t, (*driver.Tx)(nil), tx)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "Begin", output.Message)
		assert.Equal(t, LevelDebug.String(), output.Level)
	})
}

func TestConnection_Prepare(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		driverConnMock := &driverConnMock{}
		var stmtMock *statementMock
		driverConnMock.On("Prepare", mock.Anything).Return(stmtMock, driver.ErrBadConn)
		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{driverConn: driverConnMock, logger: testLogger}
		_, err := conn.Prepare(q)
		assert.Error(t, err)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "Prepare", output.Message)
		assert.Equal(t, LevelError.String(), output.Level)
		assert.Equal(t, q, output.Data[testConfig.sqlQueryFieldname])
	})

	t.Run("Success", func(t *testing.T) {
		driverConnMock := &driverConnMock{}
		stmtMock := &statementMock{}
		driverConnMock.On("Prepare", mock.Anything).Return(stmtMock, nil)
		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{driverConn: driverConnMock, logger: testLogger}
		stmt, err := conn.Prepare(q)
		assert.NoError(t, err)
		assert.Implements(t, (*driver.Stmt)(nil), stmt)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "Prepare", output.Message)
		assert.Equal(t, LevelDebug.String(), output.Level)
		assert.Equal(t, q, output.Data[testConfig.sqlQueryFieldname])
	})
}

func TestConnection_Close(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		driverConnMock := &driverConnMock{}
		driverConnMock.On("Close").Return(driver.ErrBadConn)
		conn := &connection{driverConn: driverConnMock, logger: testLogger}
		err := conn.Close()
		assert.Error(t, err)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "Close", output.Message)
		assert.Equal(t, LevelError.String(), output.Level)
		assert.Equal(t, driver.ErrBadConn.Error(), output.Data[testConfig.errorFieldname])
	})

	t.Run("Success", func(t *testing.T) {
		driverConnMock := &driverConnMock{}
		driverConnMock.On("Close").Return(nil)
		conn := &connection{driverConn: driverConnMock, logger: testLogger}
		err := conn.Close()
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "Close", output.Message)
		assert.Equal(t, LevelDebug.String(), output.Level)
	})
}

func TestConnection_BeginTx(t *testing.T) {
	t.Run("Non driver.ConnBeginTx", func(t *testing.T) {
		driverConnMock := &driverConnMock{}
		txMock := &transactionMock{}
		driverConnMock.On("Begin").Return(txMock, nil)

		conn := &connection{driverConn: driverConnMock, logger: testLogger}
		_, err := conn.BeginTx(context.TODO(), driver.TxOptions{
			Isolation: 1,
			ReadOnly:  true,
		})
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "Begin", output.Message)
		assert.Equal(t, LevelDebug.String(), output.Level)
	})

	t.Run("With driver.ConnBeginTx Error", func(t *testing.T) {
		driverConnMock := &driverConnWithContextMock{}
		var txMock *transactionMock
		driverConnMock.On("BeginTx", mock.Anything, mock.Anything).Return(txMock, driver.ErrBadConn)

		conn := &connection{driverConn: driverConnMock, logger: testLogger}
		_, err := conn.BeginTx(context.TODO(), driver.TxOptions{
			Isolation: 1,
			ReadOnly:  true,
		})
		assert.Error(t, err)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "BeginTx", output.Message)
		assert.Equal(t, LevelError.String(), output.Level)
	})

	t.Run("With driver.ConnBeginTx Success", func(t *testing.T) {
		driverConnMock := &driverConnWithContextMock{}
		txMock := &transactionMock{}
		driverConnMock.On("BeginTx", mock.Anything, mock.Anything).Return(txMock, nil)

		conn := &connection{driverConn: driverConnMock, logger: testLogger}
		tx, err := conn.BeginTx(context.TODO(), driver.TxOptions{
			Isolation: 1,
			ReadOnly:  true,
		})
		assert.NoError(t, err)
		assert.Implements(t, (*driver.Tx)(nil), tx)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "BeginTx", output.Message)
		assert.Equal(t, LevelDebug.String(), output.Level)
	})
}

func TestConnection_PrepareContext(t *testing.T) {
	t.Run("Non driver.ConnPrepareContext", func(t *testing.T) {
		driverConnMock := &driverConnMock{}
		stmtMock := &statementMock{}
		driverConnMock.On("Prepare", mock.Anything).Return(stmtMock, nil)

		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{driverConn: driverConnMock, logger: testLogger}
		_, err := conn.PrepareContext(context.TODO(), q)
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "Prepare", output.Message)
		assert.Equal(t, LevelDebug.String(), output.Level)
		assert.Equal(t, q, output.Data[testConfig.sqlQueryFieldname])
	})

	t.Run("With driver.ConnPrepareContext Error", func(t *testing.T) {
		driverConnMock := &driverConnWithContextMock{}
		var stmtMock *statementMock
		driverConnMock.On("PrepareContext", mock.Anything).Return(stmtMock, driver.ErrBadConn)

		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{driverConn: driverConnMock, logger: testLogger}
		_, err := conn.PrepareContext(context.TODO(), q)
		assert.Error(t, err)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "PrepareContext", output.Message)
		assert.Equal(t, LevelError.String(), output.Level)
		assert.Equal(t, driver.ErrBadConn.Error(), output.Data[testConfig.errorFieldname])
		assert.Equal(t, q, output.Data[testConfig.sqlQueryFieldname])
	})

	t.Run("With driver.ConnBeginTx Success", func(t *testing.T) {
		driverConnMock := &driverConnWithContextMock{}
		stmtMock := &statementMock{}
		driverConnMock.On("PrepareContext", mock.Anything).Return(stmtMock, nil)

		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{driverConn: driverConnMock, logger: testLogger}
		stmt, err := conn.PrepareContext(context.TODO(), q)
		assert.NoError(t, err)
		assert.Implements(t, (*driver.Stmt)(nil), stmt)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "PrepareContext", output.Message)
		assert.Equal(t, LevelDebug.String(), output.Level)
		assert.Equal(t, q, output.Data[testConfig.sqlQueryFieldname])
	})
}

func TestConnection_Ping(t *testing.T) {
	t.Run("Non driver.Pinger", func(t *testing.T) {
		driverConnMock := &driverConnMock{}
		conn := &connection{driverConn: driverConnMock, logger: testLogger}
		err := conn.Ping(context.TODO())
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "Ping", output.Message)
		assert.Equal(t, LevelInfo.String(), output.Level)
	})

	t.Run("driver.Pinger With Error", func(t *testing.T) {
		driverConnMock := &driverConnPingerMock{}
		driverConnMock.On("Ping", mock.Anything).Return(driver.ErrBadConn)
		conn := &connection{driverConn: driverConnMock, logger: testLogger}
		err := conn.Ping(context.TODO())
		assert.Error(t, err)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "Ping", output.Message)
		assert.Equal(t, LevelError.String(), output.Level)
		assert.Equal(t, driver.ErrBadConn.Error(), output.Data[testConfig.errorFieldname])
	})
}

func TestConnection_Exec(t *testing.T) {
	t.Run("Non driver.Execer Will Return Error", func(t *testing.T) {
		driverConnMock := &driverConnMock{}

		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{driverConn: driverConnMock, logger: testLogger}
		res, err := conn.Exec(q, []driver.Value{1})
		assert.Nil(t, res)
		assert.Error(t, err)
		assert.Equal(t, interface{}(driver.ErrSkip), err)
	})

	t.Run("driver.Execer Return Error", func(t *testing.T) {
		driverConnMock := &driverConnExecerMock{}
		resultMock := driver.ResultNoRows
		driverConnMock.On("Exec", mock.Anything, mock.Anything).Return(resultMock, driver.ErrBadConn)

		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{driverConn: driverConnMock, logger: testLogger}
		_, err := conn.Exec(q, []driver.Value{1})
		assert.Error(t, err)
		assert.Equal(t, interface{}(driver.ErrBadConn), err)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "Exec", output.Message)
		assert.Equal(t, LevelError.String(), output.Level)
		assert.Equal(t, driver.ErrBadConn.Error(), output.Data[testConfig.errorFieldname])
		assert.Equal(t, q, output.Data[testConfig.sqlQueryFieldname])
	})

	t.Run("driver.Execer Success", func(t *testing.T) {
		driverConnMock := &driverConnExecerMock{}
		resultMock := driver.ResultNoRows
		driverConnMock.On("Exec", mock.Anything, mock.Anything).Return(resultMock, nil)

		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{driverConn: driverConnMock, logger: testLogger}
		_, err := conn.Exec(q, []driver.Value{"testid"})
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "Exec", output.Message)
		assert.Equal(t, LevelInfo.String(), output.Level)
		assert.Equal(t, q, output.Data[testConfig.sqlQueryFieldname])
		assert.Equal(t, []interface{}{"testid"}, output.Data[testConfig.sqlArgsFieldname])
	})
}

func TestConnection_ExecContext(t *testing.T) {
	t.Run("Non driver.ExecerContext Return Error args", func(t *testing.T) {
		driverConnMock := &driverConnExecerMock{}
		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{driverConn: driverConnMock, logger: testLogger}
		_, err := conn.ExecContext(context.TODO(), q, []driver.NamedValue{
			{Name: "errrrr", Ordinal: 0, Value: 1},
		})
		assert.Error(t, err)
	})

	t.Run("Non driver.ExecerContext Return Error Context Done", func(t *testing.T) {
		driverConnMock := &driverConnExecerMock{}
		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{driverConn: driverConnMock, logger: testLogger}
		ctx, cancel := context.WithDeadline(context.TODO(), time.Now().Add(-1*time.Hour))
		cancel()
		_, err := conn.ExecContext(ctx, q, []driver.NamedValue{{Name: "", Ordinal: 0, Value: 1}})
		assert.Error(t, err)
		assert.Equal(t, interface{}(context.DeadlineExceeded), err)
	})

	t.Run("Non driver.ExecerContext Return Success", func(t *testing.T) {
		driverConnMock := &driverConnExecerMock{}
		resultMock := driver.ResultNoRows
		driverConnMock.On("Exec", mock.Anything, mock.Anything).Return(resultMock, nil)

		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{driverConn: driverConnMock, logger: testLogger}
		_, err := conn.ExecContext(context.TODO(), q, []driver.NamedValue{{Name: "", Ordinal: 0, Value: "testid"}})
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "Exec", output.Message)
		assert.Equal(t, LevelInfo.String(), output.Level)
		assert.Equal(t, q, output.Data[testConfig.sqlQueryFieldname])
		assert.Equal(t, []interface{}{"testid"}, output.Data[testConfig.sqlArgsFieldname])
	})

	t.Run("driver.ExecerContext Return Error", func(t *testing.T) {
		driverConnMock := &driverConnExecerContextMock{}
		resultMock := driver.ResultNoRows
		driverConnMock.On("ExecContext", mock.Anything, mock.Anything, mock.Anything).Return(resultMock, driver.ErrBadConn)

		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{driverConn: driverConnMock, logger: testLogger}
		_, err := conn.ExecContext(context.TODO(), q, []driver.NamedValue{{Name: "", Ordinal: 0, Value: "testid"}})
		assert.Error(t, err)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "ExecContext", output.Message)
		assert.Equal(t, LevelError.String(), output.Level)
		assert.Equal(t, driver.ErrBadConn.Error(), output.Data[testConfig.errorFieldname])
		assert.Equal(t, q, output.Data[testConfig.sqlQueryFieldname])
		assert.Equal(t, []interface{}{"testid"}, output.Data[testConfig.sqlArgsFieldname])
	})

	t.Run("driver.ExecerContext Success", func(t *testing.T) {
		driverConnMock := &driverConnExecerContextMock{}
		resultMock := driver.ResultNoRows
		driverConnMock.On("ExecContext", mock.Anything, mock.Anything, mock.Anything).Return(resultMock, nil)

		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{driverConn: driverConnMock, logger: testLogger}
		_, err := conn.ExecContext(context.TODO(), q, []driver.NamedValue{{Name: "", Ordinal: 0, Value: "testid"}})
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "ExecContext", output.Message)
		assert.Equal(t, LevelInfo.String(), output.Level)
		assert.Equal(t, q, output.Data[testConfig.sqlQueryFieldname])
		assert.Equal(t, []interface{}{"testid"}, output.Data[testConfig.sqlArgsFieldname])
	})
}

func TestConnection_Query(t *testing.T) {
	t.Run("Non driver.Queryer Will Return Error", func(t *testing.T) {
		driverConnMock := &driverConnMock{}

		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{driverConn: driverConnMock, logger: testLogger}
		res, err := conn.Query(q, []driver.Value{1})
		assert.Nil(t, res)
		assert.Error(t, err)
		assert.Equal(t, interface{}(driver.ErrSkip), err)
	})

	t.Run("driver.Queryer Return Error", func(t *testing.T) {
		driverConnMock := &driverConnQueryerMock{}
		resultMock := &rowsMock{}
		driverConnMock.On("Query", mock.Anything, mock.Anything).Return(resultMock, driver.ErrBadConn)

		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{driverConn: driverConnMock, logger: testLogger}
		_, err := conn.Query(q, []driver.Value{"testid"})
		assert.Error(t, err)
		assert.Equal(t, interface{}(driver.ErrBadConn), err)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "Query", output.Message)
		assert.Equal(t, LevelError.String(), output.Level)
		assert.Equal(t, driver.ErrBadConn.Error(), output.Data[testConfig.errorFieldname])
		assert.Equal(t, q, output.Data[testConfig.sqlQueryFieldname])
		assert.Equal(t, []interface{}{"testid"}, output.Data[testConfig.sqlArgsFieldname])
	})

	t.Run("driver.Queryer Success", func(t *testing.T) {
		driverConnMock := &driverConnQueryerMock{}
		resultMock := &rowsMock{}
		driverConnMock.On("Query", mock.Anything, mock.Anything).Return(resultMock, nil)

		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{driverConn: driverConnMock, logger: testLogger}
		_, err := conn.Query(q, []driver.Value{"testid"})
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "Query", output.Message)
		assert.Equal(t, LevelInfo.String(), output.Level)
		assert.Equal(t, q, output.Data[testConfig.sqlQueryFieldname])
		assert.Equal(t, []interface{}{"testid"}, output.Data[testConfig.sqlArgsFieldname])
	})
}

func TestConnection_QueryContext(t *testing.T) {
	t.Run("Non driver.QueryerContext Return Error args", func(t *testing.T) {
		driverConnMock := &driverConnQueryerMock{}
		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{driverConn: driverConnMock, logger: testLogger}
		_, err := conn.QueryContext(context.TODO(), q, []driver.NamedValue{
			{Name: "errrrr", Ordinal: 0, Value: 1},
		})
		assert.Error(t, err)
	})

	t.Run("Non driver.QueryerContext Return Error Context Done", func(t *testing.T) {
		driverConnMock := &driverConnQueryerMock{}
		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{driverConn: driverConnMock, logger: testLogger}
		ctx, cancel := context.WithDeadline(context.TODO(), time.Now().Add(-1*time.Hour))
		cancel()
		_, err := conn.QueryContext(ctx, q, []driver.NamedValue{{Name: "", Ordinal: 0, Value: 1}})
		assert.Error(t, err)
		assert.Equal(t, interface{}(context.DeadlineExceeded), err)
	})

	t.Run("Non driver.QueryerContext Return Success", func(t *testing.T) {
		driverConnMock := &driverConnQueryerMock{}
		resultMock := &rowsMock{}
		driverConnMock.On("Query", mock.Anything, mock.Anything).Return(resultMock, nil)

		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{driverConn: driverConnMock, logger: testLogger}
		_, err := conn.QueryContext(context.TODO(), q, []driver.NamedValue{{Name: "", Ordinal: 0, Value: "testid"}})
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "Query", output.Message)
		assert.Equal(t, LevelInfo.String(), output.Level)
		assert.Equal(t, q, output.Data[testConfig.sqlQueryFieldname])
		assert.Equal(t, []interface{}{"testid"}, output.Data[testConfig.sqlArgsFieldname])
	})

	t.Run("driver.QueryerContext Return Error", func(t *testing.T) {
		driverConnMock := &driverConnQueryerContextMock{}
		resultMock := &rowsMock{}
		driverConnMock.On("QueryContext", mock.Anything, mock.Anything, mock.Anything).Return(resultMock, driver.ErrBadConn)

		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{driverConn: driverConnMock, logger: testLogger}
		_, err := conn.QueryContext(context.TODO(), q, []driver.NamedValue{{Name: "", Ordinal: 0, Value: "testid"}})
		assert.Error(t, err)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "QueryContext", output.Message)
		assert.Equal(t, LevelError.String(), output.Level)
		assert.Equal(t, driver.ErrBadConn.Error(), output.Data[testConfig.errorFieldname])
		assert.Equal(t, q, output.Data[testConfig.sqlQueryFieldname])
		assert.Equal(t, []interface{}{"testid"}, output.Data[testConfig.sqlArgsFieldname])
	})

	t.Run("driver.QueryerContext Success", func(t *testing.T) {
		driverConnMock := &driverConnQueryerContextMock{}
		resultMock := &rowsMock{}
		driverConnMock.On("QueryContext", mock.Anything, mock.Anything, mock.Anything).Return(resultMock, nil)

		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{driverConn: driverConnMock, logger: testLogger}
		_, err := conn.QueryContext(context.TODO(), q, []driver.NamedValue{{Name: "", Ordinal: 0, Value: "testid"}})
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "QueryContext", output.Message)
		assert.Equal(t, LevelInfo.String(), output.Level)
		assert.Equal(t, q, output.Data[testConfig.sqlQueryFieldname])
		assert.Equal(t, []interface{}{"testid"}, output.Data[testConfig.sqlArgsFieldname])
	})
}

func TestConnection_ResetSession(t *testing.T) {
	t.Run("Non driver.SessionResetter", func(t *testing.T) {
		driverConnMock := &driverConnMock{}
		conn := &connection{driverConn: driverConnMock, logger: testLogger}
		err := conn.ResetSession(context.TODO())
		assert.NoError(t, err)
	})

	t.Run("driver.SessionResetter Return Error", func(t *testing.T) {
		driverConnMock := &driverConnResetterMock{}
		driverConnMock.On("ResetSession", mock.Anything).Return(driver.ErrBadConn)

		conn := &connection{driverConn: driverConnMock, logger: testLogger}
		err := conn.ResetSession(context.TODO())
		assert.Error(t, err)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "ResetSession", output.Message)
		assert.Equal(t, LevelError.String(), output.Level)
	})

	t.Run("driver.SessionResetter Success", func(t *testing.T) {
		driverConnMock := &driverConnResetterMock{}
		driverConnMock.On("ResetSession", mock.Anything).Return(nil)

		conn := &connection{driverConn: driverConnMock, logger: testLogger}
		err := conn.ResetSession(context.TODO())
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "ResetSession", output.Message)
		assert.Equal(t, LevelDebug.String(), output.Level)
	})
}

func TestConnection_CheckNamedValue(t *testing.T) {
	t.Run("Non driver.NamedValueChecker", func(t *testing.T) {
		driverConnMock := &driverConnMock{}
		conn := &connection{driverConn: driverConnMock, logger: testLogger}
		err := conn.CheckNamedValue(&driver.NamedValue{
			Name:    "",
			Ordinal: 0,
			Value:   "testid",
		})
		assert.NoError(t, err)
	})

	t.Run("driver.NamedValueChecker Return Error", func(t *testing.T) {
		driverConnMock := &driverConnNameValueCheckerMock{}
		driverConnMock.On("CheckNamedValue", mock.Anything).Return(driver.ErrBadConn)

		conn := &connection{driverConn: driverConnMock, logger: testLogger}
		err := conn.CheckNamedValue(&driver.NamedValue{
			Name:    "",
			Ordinal: 0,
			Value:   "testid",
		})
		assert.Error(t, err)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "CheckNamedValue", output.Message)
		assert.Equal(t, LevelError.String(), output.Level)
	})

	t.Run("driver.NamedValueChecker Success", func(t *testing.T) {
		driverConnMock := &driverConnNameValueCheckerMock{}
		driverConnMock.On("CheckNamedValue", mock.Anything).Return(nil)

		conn := &connection{driverConn: driverConnMock, logger: testLogger}
		err := conn.CheckNamedValue(&driver.NamedValue{
			Name:    "",
			Ordinal: 0,
			Value:   "testid",
		})
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "CheckNamedValue", output.Message)
		assert.Equal(t, LevelDebug.String(), output.Level)
	})
}

type driverConnMock struct {
	mock.Mock
}

func (m *driverConnMock) Prepare(query string) (driver.Stmt, error) {
	args := m.Called(query)

	return args.Get(0).(driver.Stmt), args.Error(1)
}
func (m *driverConnMock) Close() error { return m.Called().Error(0) }
func (m *driverConnMock) Begin() (driver.Tx, error) {
	return m.Called().Get(0).(driver.Tx), m.Called().Error(1)
}

type driverConnExecerMock struct {
	driverConnMock
}

func (m *driverConnExecerMock) Exec(query string, args []driver.Value) (driver.Result, error) {
	arg := m.Called(query, args)

	return arg.Get(0).(driver.Result), arg.Error(1)
}

type driverConnExecerContextMock struct {
	driverConnExecerMock
}

func (m *driverConnExecerContextMock) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	arg := m.Called(ctx, query, args)

	return arg.Get(0).(driver.Result), arg.Error(1)
}

type driverConnQueryerMock struct {
	driverConnMock
}

func (m *driverConnQueryerMock) Query(query string, args []driver.Value) (driver.Rows, error) {
	arg := m.Called(query, args)

	return arg.Get(0).(driver.Rows), arg.Error(1)
}

type driverConnQueryerContextMock struct {
	driverConnExecerMock
}

func (m *driverConnQueryerContextMock) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	arg := m.Called(ctx, query, args)

	return arg.Get(0).(driver.Rows), arg.Error(1)
}

type driverConnWithContextMock struct {
	driverConnMock
}

func (m *driverConnWithContextMock) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	args := m.Called(ctx, opts)

	return args.Get(0).(driver.Tx), args.Error(1)
}

func (m *driverConnWithContextMock) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	args := m.Called(query)

	return args.Get(0).(driver.Stmt), args.Error(1)
}

type driverConnPingerMock struct {
	driverConnMock
}

func (m *driverConnPingerMock) Ping(ctx context.Context) error { return m.Called().Error(0) }

type driverConnResetterMock struct {
	driverConnMock
}

func (m *driverConnResetterMock) ResetSession(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

type driverConnNameValueCheckerMock struct {
	driverConnMock
}

func (m *driverConnNameValueCheckerMock) CheckNamedValue(nm *driver.NamedValue) error {
	return m.Called(nm).Error(0)
}

type rowsMock struct {
	mock.Mock
}

func (m *rowsMock) Columns() []string              { return m.Called().Get(0).([]string) }
func (m *rowsMock) Close() error                   { return m.Called().Error(0) }
func (m *rowsMock) Next(dest []driver.Value) error { return m.Called(dest).Error(0) }
