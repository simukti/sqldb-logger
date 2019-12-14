package sqldblogger

import (
	"context"
	"database/sql/driver"
	"time"
)

// result is a wrapper for driver.Result.
type result struct {
	driver.Result
	logger *logger
	connID string
	stmtID string
	query  string
	args   []driver.Value
}

// LastInsertId implement driver.Result
func (r *result) LastInsertId() (int64, error) {
	lvl, start := LevelTrace, time.Now()
	id, err := r.Result.LastInsertId()

	if err != nil {
		lvl = LevelError
	}

	r.logger.log(context.Background(), lvl, "ResultLastInsertId", start, err, r.logData()...)

	return id, err
}

// RowsAffected implement driver.Result
func (r *result) RowsAffected() (int64, error) {
	lvl, start := LevelTrace, time.Now()
	num, err := r.Result.RowsAffected()

	if err != nil {
		lvl = LevelError
	}

	r.logger.log(context.Background(), lvl, "ResultRowsAffected", start, err, r.logData()...)

	return num, err
}

// logData default log data for result.
func (r *result) logData() []dataFunc {
	return []dataFunc{
		r.logger.withUID(r.logger.opt.connIDFieldname, r.connID),
		r.logger.withUID(r.logger.opt.stmtIDFieldname, r.stmtID),
		r.logger.withQuery(r.query),
		r.logger.withArgs(r.args),
	}
}
