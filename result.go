package sqldblogger

import (
	"context"
	"database/sql/driver"
	"time"
)

// result is a wrapper for driver.Result.
// result wrapper will only log on error.
type result struct {
	driver.Result
	logger    *logger
	connID    string
	stmtID    string
	query     string
	args      []driver.Value
	namedArgs []driver.NamedValue
}

// LastInsertId implement driver.Result
func (r *result) LastInsertId() (int64, error) {
	start := time.Now()
	id, err := r.Result.LastInsertId()

	if err != nil {
		r.logger.log(context.Background(), LevelError, "ResultLastInsertId", start, err, r.logData()...)
	}

	return id, err
}

// RowsAffected implement driver.Result
func (r *result) RowsAffected() (int64, error) {
	start := time.Now()
	num, err := r.Result.RowsAffected()

	if err != nil {
		r.logger.log(context.Background(), LevelError, "ResultRowsAffected", start, err, r.logData()...)
	}

	return num, err
}

// logData default log data for result.
func (r *result) logData() []dataFunc {
	return []dataFunc{
		r.logger.withUID(connID, r.connID),
		r.logger.withUID(stmtID, r.stmtID),
		r.logger.withQuery(r.query),
		r.logger.withArgs(r.args),
		r.logger.withNamedArgs(r.namedArgs),
	}
}
