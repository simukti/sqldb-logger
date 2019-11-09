package sqldblogger

import (
	"context"
	"database/sql/driver"
	"time"
)

// statement should implements:
// - driver.Stmt
// - driver.StmtExecContext
// - driver.StmtQueryContext
// - driver.NamedValueChecker
// - driver.ColumnConverter
type statement struct {
	query      string
	driverStmt driver.Stmt
	logger     *logger
}

// Close implements driver.Stmt
func (s *statement) Close() error {
	return s.driverStmt.Close()
}

// NumInput implements driver.Stmt
func (s *statement) NumInput() int {
	return s.driverStmt.NumInput()
}

// Exec implements driver.Stmt
func (s *statement) Exec(args []driver.Value) (driver.Result, error) {
	lvl, start := LevelInfo, time.Now()
	res, err := s.driverStmt.Exec(args) // nolint: staticcheck

	if err != nil {
		lvl = LevelError
	}

	s.logger.log(context.Background(), lvl, "StmtExec", start, err, s.logger.withQuery(s.query), s.logger.withArgs(args))

	return res, err
}

// Query implements driver.Stmt
func (s *statement) Query(args []driver.Value) (driver.Rows, error) {
	lvl, start := LevelInfo, time.Now()
	res, err := s.driverStmt.Query(args) // nolint: staticcheck

	if err != nil {
		lvl = LevelError
	}

	s.logger.log(context.Background(), lvl, "StmtQuery", start, err, s.logger.withQuery(s.query), s.logger.withArgs(args))

	return res, err
}

// ExecContext implements driver.StmtExecContext
func (s *statement) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	stmtExecer, ok := s.driverStmt.(driver.StmtExecContext)
	if !ok {
		dargs, err := namedValueToValue(args)
		if err != nil {
			return nil, err
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		return s.Exec(dargs)
	}

	lvl, start := LevelInfo, time.Now()
	res, err := stmtExecer.ExecContext(ctx, args)

	if err != nil {
		lvl = LevelError
	}

	s.logger.log(ctx, lvl, "StmtExecContext", start, err, s.logger.withQuery(s.query), s.logger.withNamedArgs(args))

	return res, err
}

// QueryContext implements driver.StmtQueryContext
func (s *statement) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	stmtQueryer, ok := s.driverStmt.(driver.StmtQueryContext)
	if !ok {
		dargs, err := namedValueToValue(args)
		if err != nil {
			return nil, err
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		return s.Query(dargs)
	}

	lvl, start := LevelInfo, time.Now()
	res, err := stmtQueryer.QueryContext(ctx, args)

	if err != nil {
		lvl = LevelError
	}

	s.logger.log(ctx, lvl, "StmtQueryContext", start, err, s.logger.withQuery(s.query), s.logger.withNamedArgs(args))

	return res, err
}

// CheckNamedValue implements driver.NamedValueChecker
func (s *statement) CheckNamedValue(nm *driver.NamedValue) error {
	if checker, ok := s.driverStmt.(driver.NamedValueChecker); ok {
		return checker.CheckNamedValue(nm)
	}

	return nil
}

// QueryContext implements driver.ColumnConverter
func (s *statement) ColumnConverter(idx int) driver.ValueConverter {
	// nolint: staticcheck
	if converter, ok := s.driverStmt.(driver.ColumnConverter); ok {
		return converter.ColumnConverter(idx)
	}

	return nil
}
