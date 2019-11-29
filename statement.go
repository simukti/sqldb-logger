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
	driver.Stmt
	query  string
	logger *logger
	id     string
	connID string
}

// Close implements driver.Stmt
func (s *statement) Close() error {
	return s.Stmt.Close()
}

// NumInput implements driver.Stmt
func (s *statement) NumInput() int {
	return s.Stmt.NumInput()
}

// Exec implements driver.Stmt
func (s *statement) Exec(args []driver.Value) (driver.Result, error) {
	logs := append(s.logData(), s.logger.withArgs(args))
	lvl, start := LevelInfo, time.Now()
	res, err := s.Stmt.Exec(args) // nolint: staticcheck

	if err != nil {
		lvl = LevelError
	}

	s.logger.log(context.Background(), lvl, "StmtExec", start, err, logs...)

	if err != nil {
		return res, err
	}

	return &result{Result: res, logger: s.logger, connID: s.connID, stmtID: s.id, query: s.query, args: args}, nil
}

// Query implements driver.Stmt
func (s *statement) Query(args []driver.Value) (driver.Rows, error) {
	logs := append(s.logData(), s.logger.withArgs(args))
	lvl, start := LevelInfo, time.Now()
	res, err := s.Stmt.Query(args) // nolint: staticcheck

	if err != nil {
		lvl = LevelError
	}

	s.logger.log(context.Background(), lvl, "StmtQuery", start, err, logs...)

	if err != nil {
		return res, err
	}

	return &rows{Rows: res, logger: s.logger, connID: s.connID, stmtID: s.id, query: s.query, args: args}, nil
}

// ExecContext implements driver.StmtExecContext
func (s *statement) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	stmtExecer, ok := s.Stmt.(driver.StmtExecContext)
	if !ok {
		return nil, driver.ErrSkip
	}

	logs := append(s.logData(), s.logger.withNamedArgs(args))
	lvl, start := LevelInfo, time.Now()
	res, err := stmtExecer.ExecContext(ctx, args)

	if err != nil {
		lvl = LevelError
	}

	s.logger.log(ctx, lvl, "StmtExecContext", start, err, logs...)

	if err != nil {
		return res, err
	}

	return &result{Result: res, logger: s.logger, connID: s.connID, stmtID: s.id, query: s.query, namedArgs: args}, nil
}

// QueryContext implements driver.StmtQueryContext
func (s *statement) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	stmtQueryer, ok := s.Stmt.(driver.StmtQueryContext)
	if !ok {
		return nil, driver.ErrSkip
	}

	logs := append(s.logData(), s.logger.withNamedArgs(args))
	lvl, start := LevelInfo, time.Now()
	res, err := stmtQueryer.QueryContext(ctx, args)

	if err != nil {
		lvl = LevelError
	}

	s.logger.log(ctx, lvl, "StmtQueryContext", start, err, logs...)

	if err != nil {
		return res, err
	}

	return &rows{Rows: res, logger: s.logger, connID: s.connID, stmtID: s.id, query: s.query, namedArgs: args}, nil
}

// CheckNamedValue implements driver.NamedValueChecker
func (s *statement) CheckNamedValue(nm *driver.NamedValue) error {
	checker, ok := s.Stmt.(driver.NamedValueChecker)
	if !ok {
		return driver.ErrSkip
	}

	start := time.Now()
	err := checker.CheckNamedValue(nm)

	if err != nil {
		s.logger.log(context.Background(), LevelError, "StmtCheckNamedValue", start, err, s.logData()...)
	}

	return err
}

// ColumnConverter implements driver.ColumnConverter
func (s *statement) ColumnConverter(idx int) driver.ValueConverter {
	// nolint: staticcheck
	if converter, ok := s.Stmt.(driver.ColumnConverter); ok {
		return converter.ColumnConverter(idx)
	}

	return driver.DefaultParameterConverter
}

// stmtID prepared statement log key id
const stmtID = "stmt_id"

// logData default log data for statement log.
func (s *statement) logData() []dataFunc {
	return []dataFunc{
		s.logger.withUID(connID, s.connID),
		s.logger.withUID(stmtID, s.id),
		s.logger.withQuery(s.query),
	}
}
