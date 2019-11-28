package sqldblogger

import (
	"context"
	"database/sql/driver"
	"time"
)

type result struct {
	driver.Result
	logger *logger
	connID string
}

// LastInsertId implement driver.Result
func (r *result) LastInsertId() (int64, error) {
	start := time.Now()
	id, err := r.Result.LastInsertId()

	if err != nil {
		r.logger.log(context.Background(), LevelError, "ResultLastInsertId", start, err, r.logIDs()...)
	}

	return id, err
}

// RowsAffected implement driver.Result
func (r *result) RowsAffected() (int64, error) {
	start := time.Now()
	num, err := r.Result.RowsAffected()

	if err != nil {
		r.logger.log(context.Background(), LevelError, "ResultRowsAffected", start, err, r.logIDs()...)
	}

	return num, err
}

func (r *result) logIDs() []dataFunc {
	return []dataFunc{
		r.logger.withUID(connID, r.connID),
	}
}
