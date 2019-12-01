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
	lvl, start, id := LevelDebug, time.Now(), uniqueID()
	logs := append(c.logData(), c.logger.withUID(txID, id))
	connTx, err := c.Conn.Begin() // nolint: staticcheck

	if err != nil {
		lvl = LevelError
	}

	c.logger.log(context.Background(), lvl, "Begin", start, err, logs...)

	if err != nil {
		return connTx, err
	}

	return &transaction{Tx: connTx, logger: c.logger, connID: c.id, id: id}, nil
}

// Prepare implements driver.Conn
func (c *connection) Prepare(query string) (driver.Stmt, error) {
	lvl, start, id := LevelInfo, time.Now(), uniqueID()
	logs := append(c.logData(), c.logger.withQuery(query), c.logger.withUID(stmtID, id))
	driverStmt, err := c.Conn.Prepare(query)

	if err != nil {
		lvl = LevelError
	}

	c.logger.log(context.Background(), lvl, "Prepare", start, err, logs...)

	if err != nil {
		return driverStmt, err
	}

	return &statement{query: query, Stmt: driverStmt, logger: c.logger, connID: c.id, id: id}, nil
}

// Prepare implements driver.Conn
func (c *connection) Close() error {
	var err error

	lvl, start := LevelDebug, time.Now()

	if err = c.Conn.Close(); err != nil {
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

	lvl, start, id := LevelDebug, time.Now(), uniqueID()
	logs := append(c.logData(), c.logger.withUID(txID, id))
	connTx, err := drvTx.BeginTx(ctx, opts)

	if err != nil {
		lvl = LevelError
	}

	c.logger.log(ctx, lvl, "BeginTx", start, err, logs...)

	if err != nil {
		return connTx, err
	}

	return &transaction{Tx: connTx, logger: c.logger, connID: c.id, id: id}, nil
}

// PrepareContext implements driver.ConnPrepareContext
func (c *connection) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	driverPrep, ok := c.Conn.(driver.ConnPrepareContext)
	if !ok {
		return nil, driver.ErrSkip
	}

	lvl, start, id := LevelInfo, time.Now(), uniqueID()
	logs := append(c.logData(), c.logger.withQuery(query), c.logger.withUID(stmtID, id))
	driverStmt, err := driverPrep.PrepareContext(ctx, query)

	if err != nil {
		lvl = LevelError
	}

	c.logger.log(ctx, lvl, "PrepareContext", start, err, logs...)

	if err != nil {
		return driverStmt, err
	}

	return &statement{query: query, Stmt: driverStmt, logger: c.logger, connID: c.id, id: id}, nil
}

// Ping implements driver.Pinger
func (c *connection) Ping(ctx context.Context) error {
	driverPinger, ok := c.Conn.(driver.Pinger)
	if !ok {
		return driver.ErrSkip
	}

	var err error

	lvl, start := LevelDebug, time.Now()

	if err = driverPinger.Ping(ctx); err != nil {
		lvl = LevelError
	}

	c.logger.log(ctx, lvl, "Ping", start, err, c.logData()...)

	return err
}

// Exec implements driver.Execer
// Deprecated: use ExecContext() instead
func (c *connection) Exec(query string, args []driver.Value) (driver.Result, error) {
	driverExecer, ok := c.Conn.(driver.Execer) // nolint: staticcheck
	if !ok {
		return nil, driver.ErrSkip
	}

	logs := append(c.logData(), c.logger.withQuery(query), c.logger.withArgs(args))
	lvl, start := LevelInfo, time.Now()
	res, err := driverExecer.Exec(query, args)

	if err != nil {
		lvl = LevelError
	}

	c.logger.log(context.Background(), lvl, "Exec", start, err, logs...)

	if err != nil {
		return res, err
	}

	return &result{Result: res, logger: c.logger, connID: c.id, query: query, args: args}, nil
}

// ExecContext implements driver.ExecerContext
func (c *connection) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	driverExecerContext, ok := c.Conn.(driver.ExecerContext)
	if !ok {
		return nil, driver.ErrSkip
	}

	logs := append(c.logData(), c.logger.withQuery(query), c.logger.withNamedArgs(args))
	lvl, start := LevelInfo, time.Now()
	res, err := driverExecerContext.ExecContext(ctx, query, args)

	if err != nil {
		lvl = LevelError
	}

	c.logger.log(ctx, lvl, "ExecContext", start, err, logs...)

	if err != nil {
		return res, err
	}

	return &result{Result: res, logger: c.logger, connID: c.id, query: query, namedArgs: args}, nil
}

// Query implements driver.Queryer
// Deprecated: use QueryContext() instead
func (c *connection) Query(query string, args []driver.Value) (driver.Rows, error) {
	driverQueryer, ok := c.Conn.(driver.Queryer) // nolint: staticcheck
	if !ok {
		return nil, driver.ErrSkip
	}

	logs := append(c.logData(), c.logger.withQuery(query), c.logger.withArgs(args))
	lvl, start := LevelInfo, time.Now()
	res, err := driverQueryer.Query(query, args)

	if err != nil {
		lvl = LevelError
	}

	c.logger.log(context.Background(), lvl, "Query", start, err, logs...)

	if err != nil {
		return res, err
	}

	return &rows{Rows: res, logger: c.logger, connID: c.id, query: query, args: args}, nil
}

// QueryContext implements driver.QueryerContext
func (c *connection) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	driverQueryerContext, ok := c.Conn.(driver.QueryerContext)
	if !ok {
		return nil, driver.ErrSkip
	}

	logs := append(c.logData(), c.logger.withQuery(query), c.logger.withNamedArgs(args))
	lvl, start := LevelInfo, time.Now()
	res, err := driverQueryerContext.QueryContext(ctx, query, args)

	if err != nil {
		lvl = LevelError
	}

	c.logger.log(ctx, lvl, "QueryContext", start, err, logs...)

	if err != nil {
		return res, err
	}

	return &rows{Rows: res, logger: c.logger, connID: c.id, query: query, namedArgs: args}, nil
}

// ResetSession implements driver.SessionResetter
func (c *connection) ResetSession(ctx context.Context) error {
	resetter, ok := c.Conn.(driver.SessionResetter)
	if !ok {
		return driver.ErrSkip
	}

	start := time.Now()
	err := resetter.ResetSession(ctx)

	if err != nil {
		c.logger.log(context.Background(), LevelError, "ResetSession", start, err, c.logData()...)
	}

	return err
}

// CheckNamedValue implements driver.NamedValueChecker
func (c *connection) CheckNamedValue(nm *driver.NamedValue) error {
	checker, ok := c.Conn.(driver.NamedValueChecker)
	if !ok {
		return driver.ErrSkip
	}

	start := time.Now()
	err := checker.CheckNamedValue(nm)

	if err != nil {
		c.logger.log(context.Background(), LevelError, "ConnCheckNamedValue", start, err, c.logData()...)
	}

	return err
}

// connID connection log key id
const connID = "conn_id"

// logData default log data for connection.
func (c *connection) logData() []dataFunc {
	return []dataFunc{
		c.logger.withUID(connID, c.id),
	}
}
