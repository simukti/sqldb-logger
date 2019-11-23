package sqldblogger

import (
	"context"
	"database/sql/driver"
	"time"
)

type result struct {
	driver.Result
	logger *logger
}

// LastInsertId implement driver.Result
func (r *result) LastInsertId() (int64, error) {
	start := time.Now()
	id, err := r.Result.LastInsertId()

	if err != nil {
		r.logger.log(context.Background(), LevelError, "ResultLastInsertId", start, err)
	}

	return id, err
}

// RowsAffected implement driver.Result
func (r *result) RowsAffected() (int64, error) {
	start := time.Now()
	num, err := r.Result.RowsAffected()

	if err != nil {
		r.logger.log(context.Background(), LevelError, "ResultRowsAffected", start, err)
	}

	return num, err
}
