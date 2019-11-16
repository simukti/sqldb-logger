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
	driverConn driver.Conn
	logger     *logger
}

// Begin implements driver.Conn
func (c *connection) Begin() (driver.Tx, error) {
	lvl, start := LevelDebug, time.Now()
	connTx, err := c.driverConn.Begin() // nolint: staticcheck

	if err != nil {
		lvl = LevelError
	}

	c.logger.log(context.Background(), lvl, "Begin", start, err)

	if err != nil {
		return connTx, err
	}

	return &transaction{tx: connTx, logger: c.logger}, nil
}

// Prepare implements driver.Conn
func (c *connection) Prepare(query string) (driver.Stmt, error) {
	lvl, start := LevelDebug, time.Now()
	driverStmt, err := c.driverConn.Prepare(query)

	if err != nil {
		lvl = LevelError
	}

	c.logger.log(context.Background(), lvl, "Prepare", start, err, c.logger.withQuery(query))

	if err != nil {
		return driverStmt, err
	}

	return &statement{query: query, driverStmt: driverStmt, logger: c.logger}, nil
}

// Prepare implements driver.Conn
func (c *connection) Close() error {
	var err error

	lvl, start := LevelDebug, time.Now()

	if err = c.driverConn.Close(); err != nil {
		lvl = LevelError
	}

	c.logger.log(context.Background(), lvl, "Close", start, err)

	return err
}

// BeginTx implements driver.ConnBeginTx
func (c *connection) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	drvTx, ok := c.driverConn.(driver.ConnBeginTx)
	if !ok {
		return nil, driver.ErrSkip
	}

	lvl, start := LevelDebug, time.Now()
	connTx, err := drvTx.BeginTx(ctx, opts)

	if err != nil {
		lvl = LevelError
	}

	c.logger.log(ctx, lvl, "BeginTx", start, err)

	if err != nil {
		return connTx, err
	}

	return &transaction{tx: connTx, logger: c.logger}, nil
}

// PrepareContext implements driver.ConnPrepareContext
func (c *connection) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	driverPrep, ok := c.driverConn.(driver.ConnPrepareContext)
	if !ok {
		return nil, driver.ErrSkip
	}

	lvl, start := LevelDebug, time.Now()
	driverStmt, err := driverPrep.PrepareContext(ctx, query)

	if err != nil {
		lvl = LevelError
	}

	c.logger.log(ctx, lvl, "PrepareContext", start, err, c.logger.withQuery(query))

	if err != nil {
		return driverStmt, err
	}

	return &statement{query: query, driverStmt: driverStmt, logger: c.logger}, nil
}

// Ping implements driver.Pinger
func (c *connection) Ping(ctx context.Context) error {
	var err error

	lvl, start := LevelInfo, time.Now()

	if connPing, ok := c.driverConn.(driver.Pinger); ok {
		if pingErr := connPing.Ping(ctx); pingErr != nil {
			lvl, err = LevelError, pingErr
		}
	}

	c.logger.log(ctx, lvl, "Ping", start, err)

	return err
}

// Exec implements driver.Execer
// Deprecated: use ExecContext() instead
func (c *connection) Exec(query string, args []driver.Value) (driver.Result, error) {
	driverExecer, ok := c.driverConn.(driver.Execer) // nolint: staticcheck
	if !ok {
		return nil, driver.ErrSkip
	}

	lvl, start := LevelInfo, time.Now()
	drvResult, err := driverExecer.Exec(query, args)

	if err != nil {
		lvl = LevelError
	}

	c.logger.log(context.Background(), lvl, "Exec", start, err, c.logger.withQuery(query), c.logger.withArgs(args))

	return drvResult, err
}

// ExecContext implements driver.ExecerContext
func (c *connection) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	driverExecerContext, ok := c.driverConn.(driver.ExecerContext)
	if !ok {
		return nil, driver.ErrSkip
	}

	lvl, start := LevelInfo, time.Now()
	drvResult, err := driverExecerContext.ExecContext(ctx, query, args)

	if err != nil {
		lvl = LevelError
	}

	c.logger.log(ctx, lvl, "ExecContext", start, err, c.logger.withQuery(query), c.logger.withNamedArgs(args))

	return drvResult, err
}

// Query implements driver.Queryer
// Deprecated: use QueryContext() instead
func (c *connection) Query(query string, args []driver.Value) (driver.Rows, error) {
	driverQueryer, ok := c.driverConn.(driver.Queryer) // nolint: staticcheck
	if !ok {
		return nil, driver.ErrSkip
	}

	lvl, start := LevelInfo, time.Now()
	drvResult, err := driverQueryer.Query(query, args)

	if err != nil {
		lvl = LevelError
	}

	c.logger.log(context.Background(), lvl, "Query", start, err, c.logger.withQuery(query), c.logger.withArgs(args))

	return drvResult, err
}

// QueryContext implements driver.QueryerContext
func (c *connection) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	driverQueryerContext, ok := c.driverConn.(driver.QueryerContext)
	if !ok {
		return nil, driver.ErrSkip
	}

	lvl, start := LevelInfo, time.Now()
	drvResult, err := driverQueryerContext.QueryContext(ctx, query, args)

	if err != nil {
		lvl = LevelError
	}

	c.logger.log(ctx, lvl, "QueryContext", start, err, c.logger.withQuery(query), c.logger.withNamedArgs(args))

	return drvResult, err
}

// ResetSession implements driver.SessionResetter
func (c *connection) ResetSession(ctx context.Context) error {
	driverSessionResetter, ok := c.driverConn.(driver.SessionResetter)
	if !ok {
		return driver.ErrSkip
	}

	lvl, start := LevelDebug, time.Now()
	err := driverSessionResetter.ResetSession(ctx)

	if err != nil {
		lvl = LevelError
	}

	c.logger.log(ctx, lvl, "ResetSession", start, err)

	return err
}

// CheckNamedValue implements driver.NamedValueChecker
func (c *connection) CheckNamedValue(nm *driver.NamedValue) error {
	driverNamedValueChecker, ok := c.driverConn.(driver.NamedValueChecker)
	if !ok {
		return driver.ErrSkip
	}

	lvl, start := LevelDebug, time.Now()
	err := driverNamedValueChecker.CheckNamedValue(nm)

	if err != nil {
		lvl = LevelError
	}

	c.logger.log(context.Background(), lvl, "CheckNamedValue", start, err)

	return err
}
