package sqldblogger

import (
	"context"
	"database/sql/driver"
	"time"
)

type transaction struct {
	driver.Tx
	logger *logger
}

// Commit implement driver.Tx
func (tx *transaction) Commit() error {
	lvl, start := LevelDebug, time.Now()
	err := tx.Tx.Commit()

	if err != nil {
		lvl = LevelError
	}

	tx.logger.log(context.Background(), lvl, "Commit", start, err)

	return err
}

// Rollback implement driver.Tx
func (tx *transaction) Rollback() error {
	lvl, start := LevelDebug, time.Now()
	err := tx.Tx.Rollback()

	if err != nil {
		lvl = LevelError
	}

	tx.logger.log(context.Background(), lvl, "Rollback", start, err)

	return err
}
