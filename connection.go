package sqldblogger

import (
	"context"
	"database/sql/driver"
	"time"
)

// connection is a database connection wrapper which implements following interfaces:
// - driver.Conn
// - driver.ConnBeginTx
// - driver.ConnPrepareContext
// - driver.Pinger
// - driver.Execer
// - driver.ExecerContext
// - driver.Queryer
// - driver.QueryerContext
// - driver.SessionResetter
// - driver.NamedValueChecker
type connection struct {
	driver.Conn
	id     string
	logger *logger
}

// Begin implements driver.Conn
func (c *connection) Begin() (driver.Tx, error) {
	lvl, start, id := LevelDebug, time.Now(), c.logger.opt.uidGenerator.UniqueID()
	logs := append(c.logData(), c.logger.withUID(c.logger.opt.txIDFieldname, id))
	connTx, err := c.Conn.Begin() // nolint // disable static check on deprecated driver method

	if err != nil {
		lvl = LevelError
	}

	c.logger.log(context.Background(), lvl, "Begin", start, err, logs...)

	return c.transaction(connTx, err, id)
}

// Prepare implements driver.Conn
func (c *connection) Prepare(query string) (driver.Stmt, error) {
	lvl, start, id := c.logger.opt.preparerLevel, time.Now(), c.logger.opt.uidGenerator.UniqueID()
	logs := append(c.logData(), c.logger.withQuery(query), c.logger.withUID(c.logger.opt.stmtIDFieldname, id))
	driverStmt, err := c.Conn.Prepare(query)

	if err != nil {
		lvl = LevelError
	}

	c.logger.log(context.Background(), lvl, "Prepare", start, err, logs...)

	return c.statement(driverStmt, err, id, query)
}

// Prepare implements driver.Conn
func (c *connection) Close() error {
	lvl, start := LevelDebug, time.Now()
	err := c.Conn.Close()

	if err != nil {
		lvl = LevelError
	}

	c.logger.log(context.Background(), lvl, "Close", start, err, c.logData()...)

	return err
}

// BeginTx implements driver.ConnBeginTx
func (c *connection) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	drvTx, ok := c.Conn.(driver.ConnBeginTx)
	if !ok {
		return nil, driver.ErrSkip
	}

	lvl, start, id := LevelDebug, time.Now(), c.logger.opt.uidGenerator.UniqueID()
	logs := append(c.logData(), c.logger.withUID(c.logger.opt.txIDFieldname, id))
	connTx, err := drvTx.BeginTx(ctx, opts)

	if err != nil {
		lvl = LevelError
	}

	c.logger.log(ctx, lvl, "BeginTx", start, err, logs...)

	return c.transaction(connTx, err, id)
}

// PrepareContext implements driver.ConnPrepareContext
func (c *connection) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	driverPrep, ok := c.Conn.(driver.ConnPrepareContext)
	if !ok {
		return nil, driver.ErrSkip
	}

	lvl, start, id := c.logger.opt.preparerLevel, time.Now(), c.logger.opt.uidGenerator.UniqueID()
	logs := append(c.logData(), c.logger.withQuery(query), c.logger.withUID(c.logger.opt.stmtIDFieldname, id))
	driverStmt, err := driverPrep.PrepareContext(ctx, query)

	if err != nil {
		lvl = LevelError
	}

	c.logger.log(ctx, lvl, "PrepareContext", start, err, logs...)

	return c.statement(driverStmt, err, id, query)
}

// Ping implements driver.Pinger
func (c *connection) Ping(ctx context.Context) error {
	driverPinger, ok := c.Conn.(driver.Pinger)
	if !ok {
		return driver.ErrSkip
	}

	lvl, start := LevelDebug, time.Now()
	err := driverPinger.Ping(ctx)

	if err != nil {
		lvl = LevelError
	}

	c.logger.log(ctx, lvl, "Ping", start, err, c.logData()...)

	return err
}

// Exec implements driver.Execer
// Deprecated: use ExecContext() instead
func (c *connection) Exec(query string, args []driver.Value) (driver.Result, error) {
	driverExecer, ok := c.Conn.(driver.Execer) // nolint // disable static check on deprecated driver method
	if !ok {
		return nil, driver.ErrSkip
	}

	logs := append(c.logData(), c.logger.withQuery(query), c.logger.withArgs(args))
	lvl, start := c.logger.opt.execerLevel, time.Now()
	res, err := driverExecer.Exec(query, args)

	if err != nil {
		lvl = LevelError
	}

	c.logger.log(context.Background(), lvl, "Exec", start, err, logs...)

	return c.result(res, err, query, args)
}

// ExecContext implements driver.ExecerContext
func (c *connection) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	driverExecerContext, ok := c.Conn.(driver.ExecerContext)
	if !ok {
		return nil, driver.ErrSkip
	}

	logArgs := namedValuesToValues(args)
	logs := append(c.logData(), c.logger.withQuery(query), c.logger.withArgs(logArgs))
	lvl, start := c.logger.opt.execerLevel, time.Now()
	res, err := driverExecerContext.ExecContext(ctx, query, args)

	if err != nil {
		lvl = LevelError
	}

	c.logger.log(ctx, lvl, "ExecContext", start, err, logs...)

	return c.result(res, err, query, logArgs)
}

// Query implements driver.Queryer
// Deprecated: use QueryContext() instead
func (c *connection) Query(query string, args []driver.Value) (driver.Rows, error) {
	driverQueryer, ok := c.Conn.(driver.Queryer) // nolint // disable static check on deprecated driver method
	if !ok {
		return nil, driver.ErrSkip
	}

	logs := append(c.logData(), c.logger.withQuery(query), c.logger.withArgs(args))
	lvl, start := c.logger.opt.queryerLevel, time.Now()
	res, err := driverQueryer.Query(query, args)

	if err != nil {
		lvl = LevelError
	}

	c.logger.log(context.Background(), lvl, "Query", start, err, logs...)

	return c.rows(res, err, query, args)
}

// QueryContext implements driver.QueryerContext
func (c *connection) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	driverQueryerContext, ok := c.Conn.(driver.QueryerContext)
	if !ok {
		return nil, driver.ErrSkip
	}

	logArgs := namedValuesToValues(args)
	logs := append(c.logData(), c.logger.withQuery(query), c.logger.withArgs(logArgs))
	lvl, start := c.logger.opt.queryerLevel, time.Now()
	res, err := driverQueryerContext.QueryContext(ctx, query, args)

	if err != nil {
		lvl = LevelError
	}

	c.logger.log(ctx, lvl, "QueryContext", start, err, logs...)

	return c.rows(res, err, query, logArgs)
}

// ResetSession implements driver.SessionResetter
func (c *connection) ResetSession(ctx context.Context) error {
	resetter, ok := c.Conn.(driver.SessionResetter)
	if !ok {
		return driver.ErrSkip
	}

	lvl, start := LevelTrace, time.Now()
	err := resetter.ResetSession(ctx)

	if err != nil {
		lvl = LevelError
	}

	c.logger.log(context.Background(), lvl, "ResetSession", start, err, c.logData()...)

	return err
}

// CheckNamedValue implements driver.NamedValueChecker
func (c *connection) CheckNamedValue(nm *driver.NamedValue) error {
	checker, ok := c.Conn.(driver.NamedValueChecker)
	if !ok {
		return driver.ErrSkip
	}

	lvl, start := LevelTrace, time.Now()
	err := checker.CheckNamedValue(nm)

	if err != nil {
		lvl = LevelError
	}

	c.logger.log(context.Background(), lvl, "CheckNamedValue", start, err, c.logData()...)

	return err
}

func (c *connection) transaction(tx driver.Tx, err error, id string) (driver.Tx, error) {
	if err != nil {
		return tx, err
	}

	return &transaction{Tx: tx, logger: c.logger, connID: c.id, id: id}, nil
}

func (c *connection) statement(stmt driver.Stmt, err error, id, query string) (driver.Stmt, error) {
	if err != nil {
		return stmt, err
	}

	return &statement{Stmt: stmt, query: query, logger: c.logger, connID: c.id, id: id}, nil
}

func (c *connection) rows(res driver.Rows, err error, query string, args []driver.Value) (driver.Rows, error) {
	if !c.logger.opt.wrapResult || err != nil {
		return res, err
	}

	return &rows{Rows: res, logger: c.logger, connID: c.id, query: query, args: args}, nil
}

func (c *connection) result(res driver.Result, err error, query string, args []driver.Value) (driver.Result, error) {
	if !c.logger.opt.wrapResult || err != nil {
		return res, err
	}

	return &result{Result: res, logger: c.logger, connID: c.id, query: query, args: args}, nil
}

// logData default log data for connection.
func (c *connection) logData() []dataFunc {
	return []dataFunc{
		c.logger.withUID(c.logger.opt.connIDFieldname, c.id),
	}
}
