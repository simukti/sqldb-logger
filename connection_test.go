package sqldblogger

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConnection_Begin(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		ml := newMockLogger()

		var txMock *transactionMock
		driverConnMock := &driverConnMock{}
		driverConnMock.On("Begin").Return(txMock, driver.ErrBadConn)

		id := ml.testLogger.opt.uidGenerator.UniqueID()
		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: id}
		_, err := conn.Begin()
		assert.Error(t, err)
		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, LevelError.String(), output.Level)
		assert.Equal(t, id, output.Data[ml.testOpts.connIDFieldname])
	})

	t.Run("Success", func(t *testing.T) {
		ml := newMockLogger()
		txMock := &transactionMock{}
		driverConnMock := &driverConnMock{}
		driverConnMock.On("Begin").Return(txMock, nil)

		id := ml.testLogger.opt.uidGenerator.UniqueID()
		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: id}
		tx, err := conn.Begin()
		assert.NoError(t, err)
		assert.Implements(t, (*driver.Tx)(nil), tx)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "Begin", output.Message)
		assert.Equal(t, LevelDebug.String(), output.Level)
		assert.Equal(t, id, output.Data[ml.testOpts.connIDFieldname])
	})
}

func TestConnection_Prepare(t *testing.T) {

	t.Run("Error", func(t *testing.T) {
		ml := newMockLogger()
		var stmtMock *statementMock
		driverConnMock := &driverConnMock{}
		driverConnMock.On("Prepare", mock.Anything).Return(stmtMock, driver.ErrBadConn)
		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID()}
		_, err := conn.Prepare(q)
		assert.Error(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "Prepare", output.Message)
		assert.Equal(t, LevelError.String(), output.Level)
		assert.Equal(t, q, output.Data[ml.testOpts.sqlQueryFieldname])
		assert.Equal(t, conn.id, output.Data[ml.testOpts.connIDFieldname])
	})

	t.Run("Success", func(t *testing.T) {
		ml := newMockLogger()
		driverConnMock := &driverConnMock{}
		stmtMock := &statementMock{}
		driverConnMock.On("Prepare", mock.Anything).Return(stmtMock, nil)
		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID()}
		stmt, err := conn.Prepare(q)
		assert.NoError(t, err)
		assert.Implements(t, (*driver.Stmt)(nil), stmt)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, false, isAbleToPrint(ml.testOpts, MessagePrepare, LevelDebug))
	})

	t.Run("Success with cusmlm level", func(t *testing.T) {
		ml := newMockLogger()
		driverConnMock := &driverConnMock{}
		stmtMock := &statementMock{}
		driverConnMock.On("Prepare", mock.Anything).Return(stmtMock, nil)
		q := "SELECT * FROM tt WHERE id = ?"

		ml.testOpts.preparerLevel = LevelDebug

		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID()}
		stmt, err := conn.Prepare(q)
		assert.NoError(t, err)
		assert.Implements(t, (*driver.Stmt)(nil), stmt)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "Prepare", output.Message)
		assert.Equal(t, getDefaultLevelByMessage(output.Message).String(), output.Level)
		assert.Equal(t, q, output.Data[ml.testOpts.sqlQueryFieldname])
		assert.Equal(t, conn.id, output.Data[ml.testOpts.connIDFieldname])
		assert.NotEmpty(t, output.Data[ml.testOpts.stmtIDFieldname])
	})
}

func TestConnection_Close(t *testing.T) {
	ml := newMockLogger()

	t.Run("Error", func(t *testing.T) {
		driverConnMock := &driverConnMock{}
		driverConnMock.On("Close").Return(driver.ErrBadConn)
		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID()}
		err := conn.Close()
		assert.Error(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "Close", output.Message)
		assert.Equal(t, LevelError.String(), output.Level)
		assert.Equal(t, driver.ErrBadConn.Error(), output.Data[ml.testOpts.errorFieldname])
		assert.Equal(t, conn.id, output.Data[ml.testOpts.connIDFieldname])
	})

	t.Run("Success", func(t *testing.T) {
		driverConnMock := &driverConnMock{}
		driverConnMock.On("Close").Return(nil)
		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID()}
		err := conn.Close()
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "Close", output.Message)
		assert.Equal(t, LevelDebug.String(), output.Level)
		assert.Equal(t, conn.id, output.Data[ml.testOpts.connIDFieldname])
	})
}

func TestConnection_BeginTx(t *testing.T) {

	t.Run("Non driver.ConnBeginTx", func(t *testing.T) {
		ml := newMockLogger()
		driverConnMock := &driverConnMock{}
		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID()}
		_, err := conn.BeginTx(context.TODO(), driver.TxOptions{
			Isolation: 1,
			ReadOnly:  true,
		})
		assert.Error(t, err)
		assert.Equal(t, driver.ErrSkip, err)
	})

	t.Run("With driver.ConnBeginTx Error", func(t *testing.T) {
		ml := newMockLogger()
		driverConnMock := &driverConnWithContextMock{}
		var txMock *transactionMock
		driverConnMock.On("BeginTx", mock.Anything, mock.Anything).Return(txMock, driver.ErrBadConn)

		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID()}
		_, err := conn.BeginTx(context.TODO(), driver.TxOptions{
			Isolation: 1,
			ReadOnly:  true,
		})
		assert.Error(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "BeginTx", output.Message)
		assert.Equal(t, LevelError.String(), output.Level)
		assert.Equal(t, conn.id, output.Data[ml.testOpts.connIDFieldname])
		assert.NotEmpty(t, output.Data[ml.testOpts.txIDFieldname])
	})

	t.Run("With driver.ConnBeginTx Success", func(t *testing.T) {
		ml := newMockLogger()
		driverConnMock := &driverConnWithContextMock{}
		txMock := &transactionMock{}
		driverConnMock.On("BeginTx", mock.Anything, mock.Anything).Return(txMock, nil)

		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID()}
		tx, err := conn.BeginTx(context.TODO(), driver.TxOptions{
			Isolation: 1,
			ReadOnly:  true,
		})
		assert.NoError(t, err)
		assert.Implements(t, (*driver.Tx)(nil), tx)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "BeginTx", output.Message)
		assert.Equal(t, LevelDebug.String(), output.Level)
		assert.Equal(t, conn.id, output.Data[ml.testOpts.connIDFieldname])
		assert.NotEmpty(t, output.Data[ml.testOpts.txIDFieldname])
	})
}

func TestConnection_PrepareContext(t *testing.T) {

	t.Run("Non driver.ConnPrepareContext", func(t *testing.T) {
		ml := newMockLogger()
		driverConnMock := &driverConnMock{}

		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID()}
		_, err := conn.PrepareContext(context.TODO(), q)
		assert.Error(t, err)
		assert.Equal(t, driver.ErrSkip, err)
	})

	t.Run("With driver.ConnPrepareContext Error", func(t *testing.T) {
		ml := newMockLogger()
		driverConnMock := &driverConnWithContextMock{}
		var stmtMock *statementMock
		driverConnMock.On("PrepareContext", mock.Anything).Return(stmtMock, driver.ErrBadConn)

		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID()}
		_, err := conn.PrepareContext(context.TODO(), q)
		assert.Error(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "PrepareContext", output.Message)
		assert.Equal(t, LevelError.String(), output.Level)
		assert.Equal(t, driver.ErrBadConn.Error(), output.Data[ml.testOpts.errorFieldname])
		assert.Equal(t, q, output.Data[ml.testOpts.sqlQueryFieldname])
		assert.Equal(t, conn.id, output.Data[ml.testOpts.connIDFieldname])
		assert.NotEmpty(t, output.Data[ml.testOpts.stmtIDFieldname])
	})

	t.Run("With driver.ConnBeginTx Success", func(t *testing.T) {
		ml := newMockLogger()
		driverConnMock := &driverConnWithContextMock{}
		stmtMock := &statementMock{}
		driverConnMock.On("PrepareContext", mock.Anything).Return(stmtMock, nil)

		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID()}
		stmt, err := conn.PrepareContext(context.TODO(), q)
		assert.NoError(t, err)
		assert.Implements(t, (*driver.Stmt)(nil), stmt)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, false, isAbleToPrint(ml.testOpts, MessagePrepareContext, LevelDebug))
	})

	t.Run("With driver.ConnBeginTx Success and cusmlm preparer level", func(t *testing.T) {
		ml := newMockLogger()
		ml.testOpts.preparerLevel = LevelDebug

		driverConnMock := &driverConnWithContextMock{}
		stmtMock := &statementMock{}
		driverConnMock.On("PrepareContext", mock.Anything).Return(stmtMock, nil)

		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: ml.testOpts.uidGenerator.UniqueID()}
		stmt, err := conn.PrepareContext(context.TODO(), q)
		assert.NoError(t, err)
		assert.Implements(t, (*driver.Stmt)(nil), stmt)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "PrepareContext", output.Message)
		assert.Equal(t, LevelDebug.String(), output.Level)
		assert.Equal(t, q, output.Data[ml.testOpts.sqlQueryFieldname])
		assert.Equal(t, conn.id, output.Data[ml.testOpts.connIDFieldname])
		assert.NotEmpty(t, output.Data[ml.testOpts.stmtIDFieldname])
	})
}

func TestConnection_Ping(t *testing.T) {

	t.Run("Non driver.Pinger", func(t *testing.T) {
		ml := newMockLogger()
		driverConnMock := &driverConnMock{}
		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID()}
		err := conn.Ping(context.TODO())
		assert.Error(t, err)
		assert.Equal(t, driver.ErrSkip, err)
	})

	t.Run("driver.Pinger With Error", func(t *testing.T) {
		ml := newMockLogger()
		driverConnMock := &driverConnPingerMock{}
		driverConnMock.On("Ping", mock.Anything).Return(driver.ErrBadConn)
		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID()}
		err := conn.Ping(context.TODO())
		assert.Error(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "Ping", output.Message)
		assert.Equal(t, LevelError.String(), output.Level)
		assert.Equal(t, driver.ErrBadConn.Error(), output.Data[ml.testOpts.errorFieldname])
		assert.Equal(t, conn.id, output.Data[ml.testOpts.connIDFieldname])
	})

	t.Run("driver.Pinger Success", func(t *testing.T) {
		ml := newMockLogger()
		driverConnMock := &driverConnPingerMock{}
		driverConnMock.On("Ping", mock.Anything).Return(nil)
		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID()}
		err := conn.Ping(context.TODO())
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "Ping", output.Message)
		assert.Equal(t, LevelDebug.String(), output.Level)
		assert.Equal(t, conn.id, output.Data[ml.testOpts.connIDFieldname])
	})
}

func TestConnection_Exec(t *testing.T) {

	t.Run("Non driver.Execer Will Return Error", func(t *testing.T) {
		ml := newMockLogger()
		driverConnMock := &driverConnMock{}

		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID()}
		res, err := conn.Exec(q, []driver.Value{1})
		assert.Nil(t, res)
		assert.Error(t, err)
		assert.Equal(t, interface{}(driver.ErrSkip), err)
	})

	t.Run("driver.Execer Return Error", func(t *testing.T) {
		ml := newMockLogger()
		driverConnMock := &driverConnExecerMock{}
		resultMock := driver.ResultNoRows
		driverConnMock.On("Exec", mock.Anything, mock.Anything).Return(resultMock, driver.ErrBadConn)

		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID()}
		_, err := conn.Exec(q, []driver.Value{1})
		assert.Error(t, err)
		assert.Equal(t, interface{}(driver.ErrBadConn), err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "Exec", output.Message)
		assert.Equal(t, LevelError.String(), output.Level)
		assert.Equal(t, driver.ErrBadConn.Error(), output.Data[ml.testOpts.errorFieldname])
		assert.Equal(t, q, output.Data[ml.testOpts.sqlQueryFieldname])
		assert.Equal(t, conn.id, output.Data[ml.testOpts.connIDFieldname])
	})

	t.Run("driver.Execer Success", func(t *testing.T) {
		ml := newMockLogger()
		driverConnMock := &driverConnExecerMock{}
		resultMock := driver.ResultNoRows
		driverConnMock.On("Exec", mock.Anything, mock.Anything).Return(resultMock, nil)

		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID()}
		_, err := conn.Exec(q, []driver.Value{"testid"})
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "Exec", output.Message)
		assert.Equal(t, LevelInfo.String(), output.Level)
		assert.Equal(t, q, output.Data[ml.testOpts.sqlQueryFieldname])
		assert.Equal(t, []interface{}{"testid"}, output.Data[ml.testOpts.sqlArgsFieldname])
		assert.Equal(t, conn.id, output.Data[ml.testOpts.connIDFieldname])
	})

	t.Run("driver.Execer Success With Custom Level", func(t *testing.T) {
		ml := newMockLogger()
		ml.testOpts.execerLevel = LevelDebug

		driverConnMock := &driverConnExecerMock{}
		resultMock := driver.ResultNoRows
		driverConnMock.On("Exec", mock.Anything, mock.Anything).Return(resultMock, nil)

		q := "SELECT * FROM tt WHERE id = ?"

		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: ml.testOpts.uidGenerator.UniqueID()}
		_, err := conn.Exec(q, []driver.Value{"testid"})
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "Exec", output.Message)
		if _, ok := output.Data[ml.testOpts.errorFieldname]; ok {
			fmt.Println(output.Data[ml.testOpts.errorFieldname])
			assert.Equal(t, LevelError.String(), output.Level)
		} else {
			assert.Equal(t, LevelInfo.String(), output.Level)
		}
		assert.Equal(t, q, output.Data[ml.testOpts.sqlQueryFieldname])
		assert.Equal(t, []interface{}{"testid"}, output.Data[ml.testOpts.sqlArgsFieldname])
		assert.Equal(t, conn.id, output.Data[ml.testOpts.connIDFieldname])
	})
}

func TestConnection_ExecContext(t *testing.T) {
	t.Run("Non driver.ExecerContext Return Error args", func(t *testing.T) {
		ml := newMockLogger()

		driverConnMock := &driverConnExecerMock{}
		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID()}
		_, err := conn.ExecContext(context.TODO(), q, []driver.NamedValue{
			{Name: "errrrr", Ordinal: 0, Value: 1},
		})
		assert.Error(t, err)
	})

	t.Run("driver.ExecerContext Return Error", func(t *testing.T) {
		ml := newMockLogger()

		driverConnMock := &driverConnExecerContextMock{}
		resultMock := driver.ResultNoRows
		driverConnMock.On("ExecContext", mock.Anything, mock.Anything, mock.Anything).Return(resultMock, driver.ErrBadConn)

		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID()}
		_, err := conn.ExecContext(context.TODO(), q, []driver.NamedValue{{Name: "", Ordinal: 0, Value: "testid"}})
		assert.Error(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "ExecContext", output.Message)
		assert.Equal(t, LevelError.String(), output.Level)
		assert.Equal(t, driver.ErrBadConn.Error(), output.Data[ml.testOpts.errorFieldname])
		assert.Equal(t, q, output.Data[ml.testOpts.sqlQueryFieldname])
		assert.Equal(t, []interface{}{"testid"}, output.Data[ml.testOpts.sqlArgsFieldname])
		assert.Equal(t, conn.id, output.Data[ml.testOpts.connIDFieldname])
	})

	t.Run("driver.ExecerContext Success", func(t *testing.T) {
		ml := newMockLogger()

		driverConnMock := &driverConnExecerContextMock{}
		resultMock := driver.ResultNoRows
		driverConnMock.On("ExecContext", mock.Anything, mock.Anything, mock.Anything).Return(resultMock, nil)

		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID()}
		_, err := conn.ExecContext(context.TODO(), q, []driver.NamedValue{{Name: "", Ordinal: 0, Value: "testid"}})
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, false, isAbleToPrint(ml.testOpts, MessageExecContext, LevelDebug))
	})

	t.Run("driver.ExecerContext Success", func(t *testing.T) {
		ml := newMockLogger()
		ml.testOpts.execerLevel = LevelDebug

		driverConnMock := &driverConnExecerContextMock{}
		resultMock := driver.ResultNoRows
		driverConnMock.On("ExecContext", mock.Anything, mock.Anything, mock.Anything).Return(resultMock, nil)

		q := "SELECT * FROM tt WHERE id = ?"

		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: ml.testOpts.uidGenerator.UniqueID()}
		_, err := conn.ExecContext(context.TODO(), q, []driver.NamedValue{{Name: "", Ordinal: 0, Value: "testid"}})
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "ExecContext", output.Message)
		assert.Equal(t, LevelDebug.String(), output.Level)
		assert.Equal(t, q, output.Data[ml.testOpts.sqlQueryFieldname])
		assert.Equal(t, []interface{}{"testid"}, output.Data[ml.testOpts.sqlArgsFieldname])
		assert.Equal(t, conn.id, output.Data[ml.testOpts.connIDFieldname])
	})
}

func TestConnection_Query(t *testing.T) {

	t.Run("Non driver.Queryer Will Return Error", func(t *testing.T) {
		ml := newMockLogger()

		driverConnMock := &driverConnMock{}

		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID()}
		res, err := conn.Query(q, []driver.Value{1})
		assert.Nil(t, res)
		assert.Error(t, err)
		assert.Equal(t, driver.ErrSkip, err)
	})

	t.Run("driver.Queryer Return Error", func(t *testing.T) {
		ml := newMockLogger()

		driverConnMock := &driverConnQueryerMock{}
		resultMock := &rowsMock{}
		driverConnMock.On("Query", mock.Anything, mock.Anything).Return(resultMock, driver.ErrBadConn)

		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID()}
		_, err := conn.Query(q, []driver.Value{"testid"})
		assert.Error(t, err)
		assert.Equal(t, interface{}(driver.ErrBadConn), err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "Query", output.Message)
		assert.Equal(t, LevelError.String(), output.Level)
		assert.Equal(t, driver.ErrBadConn.Error(), output.Data[ml.testOpts.errorFieldname])
		assert.Equal(t, q, output.Data[ml.testOpts.sqlQueryFieldname])
		assert.Equal(t, []interface{}{"testid"}, output.Data[ml.testOpts.sqlArgsFieldname])
		assert.Equal(t, conn.id, output.Data[ml.testOpts.connIDFieldname])
	})

	t.Run("driver.Queryer Success", func(t *testing.T) {
		ml := newMockLogger()

		driverConnMock := &driverConnQueryerMock{}
		resultMock := &rowsMock{}
		driverConnMock.On("Query", mock.Anything, mock.Anything).Return(resultMock, nil)

		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID()}
		_, err := conn.Query(q, []driver.Value{"testid"})
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, false, isAbleToPrint(ml.testOpts, "Query", LevelDebug))
	})

	t.Run("driver.Queryer Success With Custom Level", func(t *testing.T) {
		ml := newMockLogger()
		ml.testOpts.queryerLevel = LevelDebug

		driverConnMock := &driverConnQueryerMock{}
		resultMock := &rowsMock{}
		driverConnMock.On("Query", mock.Anything, mock.Anything).Return(resultMock, nil)

		q := "SELECT * FROM tt WHERE id = ?"

		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: ml.testOpts.uidGenerator.UniqueID()}
		_, err := conn.Query(q, []driver.Value{"testid"})
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "Query", output.Message)
		assert.Equal(t, LevelDebug.String(), output.Level)
		assert.Equal(t, q, output.Data[ml.testOpts.sqlQueryFieldname])
		assert.Equal(t, []interface{}{"testid"}, output.Data[ml.testOpts.sqlArgsFieldname])
		assert.Equal(t, conn.id, output.Data[ml.testOpts.connIDFieldname])
	})
}

func TestConnection_QueryContext(t *testing.T) {

	t.Run("Non driver.QueryerContext Return Error args", func(t *testing.T) {
		ml := newMockLogger()

		driverConnMock := &driverConnQueryerMock{}
		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID()}
		_, err := conn.QueryContext(context.TODO(), q, []driver.NamedValue{
			{Name: "errrrr", Ordinal: 0, Value: 1},
		})
		assert.Error(t, err)
		assert.Equal(t, driver.ErrSkip, err)
	})

	t.Run("driver.QueryerContext Return Error", func(t *testing.T) {
		ml := newMockLogger()

		driverConnMock := &driverConnQueryerContextMock{}
		resultMock := &rowsMock{}
		driverConnMock.On("QueryContext", mock.Anything, mock.Anything, mock.Anything).Return(resultMock, driver.ErrBadConn)

		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID()}
		_, err := conn.QueryContext(context.TODO(), q, []driver.NamedValue{{Name: "", Ordinal: 0, Value: "testid"}})
		assert.Error(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "QueryContext", output.Message)
		assert.Equal(t, LevelError.String(), output.Level)
		assert.Equal(t, driver.ErrBadConn.Error(), output.Data[ml.testOpts.errorFieldname])
		assert.Equal(t, q, output.Data[ml.testOpts.sqlQueryFieldname])
		assert.Equal(t, []interface{}{"testid"}, output.Data[ml.testOpts.sqlArgsFieldname])
		assert.Equal(t, conn.id, output.Data[ml.testOpts.connIDFieldname])
	})

	t.Run("driver.QueryerContext Success", func(t *testing.T) {
		ml := newMockLogger()

		driverConnMock := &driverConnQueryerContextMock{}
		resultMock := &rowsMock{}
		driverConnMock.On("QueryContext", mock.Anything, mock.Anything, mock.Anything).Return(resultMock, nil)

		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID()}
		_, err := conn.QueryContext(context.TODO(), q, []driver.NamedValue{{Name: "", Ordinal: 0, Value: "testid"}})
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, false, isAbleToPrint(ml.testOpts, MessageQueryContext, LevelDebug))
	})

	t.Run("driver.QueryerContext Success With Custom Level", func(t *testing.T) {
		ml := newMockLogger()
		ml.testOpts.queryerLevel = LevelDebug

		driverConnMock := &driverConnQueryerContextMock{}
		resultMock := &rowsMock{}
		driverConnMock.On("QueryContext", mock.Anything, mock.Anything, mock.Anything).Return(resultMock, nil)

		q := "SELECT * FROM tt WHERE id = ?"
		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: ml.testOpts.uidGenerator.UniqueID()}
		_, err := conn.QueryContext(context.TODO(), q, []driver.NamedValue{{Name: "", Ordinal: 0, Value: "testid"}})
		assert.NoError(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "QueryContext", output.Message)
		assert.Equal(t, LevelDebug.String(), output.Level)
		assert.Equal(t, q, output.Data[ml.testOpts.sqlQueryFieldname])
		assert.Equal(t, []interface{}{"testid"}, output.Data[ml.testOpts.sqlArgsFieldname])
		assert.Equal(t, conn.id, output.Data[ml.testOpts.connIDFieldname])
	})
}

func TestConnection_ResetSession(t *testing.T) {
	t.Run("Non driver.SessionResetter", func(t *testing.T) {
		ml := newMockLogger()

		driverConnMock := &driverConnMock{}
		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID()}
		err := conn.ResetSession(context.TODO())
		assert.Error(t, err)
		assert.Error(t, driver.ErrSkip, err)
	})

	t.Run("driver.SessionResetter Return Error", func(t *testing.T) {
		ml := newMockLogger()

		driverConnMock := &driverConnResetterMock{}
		driverConnMock.On("ResetSession", mock.Anything).Return(driver.ErrBadConn)

		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID()}
		err := conn.ResetSession(context.TODO())
		assert.Error(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "ResetSession", output.Message)
		assert.Equal(t, LevelError.String(), output.Level)
		assert.Equal(t, conn.id, output.Data[ml.testOpts.connIDFieldname])
	})
}

func TestConnection_CheckNamedValue(t *testing.T) {
	t.Run("Non driver.NamedValueChecker", func(t *testing.T) {
		ml := newMockLogger()

		driverConnMock := &driverConnMock{}
		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID()}
		err := conn.CheckNamedValue(&driver.NamedValue{
			Name:    "",
			Ordinal: 0,
			Value:   "testid",
		})
		assert.Error(t, err)
		assert.Equal(t, driver.ErrSkip, err)
	})

	t.Run("driver.NamedValueChecker Return Error", func(t *testing.T) {
		ml := newMockLogger()

		driverConnMock := &driverConnNameValueCheckerMock{}
		driverConnMock.On("CheckNamedValue", mock.Anything).Return(driver.ErrBadConn)

		conn := &connection{Conn: driverConnMock, logger: ml.testLogger, id: ml.testLogger.opt.uidGenerator.UniqueID()}
		err := conn.CheckNamedValue(&driver.NamedValue{
			Name:    "",
			Ordinal: 0,
			Value:   "testid",
		})
		assert.Error(t, err)

		var output bufLog
		err = json.Unmarshal(ml.bufLogger.Bytes(), &output)
		assert.NoError(t, err)
		assert.Equal(t, "CheckNamedValue", output.Message)
		assert.Equal(t, LevelError.String(), output.Level)
		assert.NotEmpty(t, output.Data[ml.testOpts.connIDFieldname])
		assert.Equal(t, conn.id, output.Data[ml.testOpts.connIDFieldname])
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
