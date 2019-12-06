package sqldblogger

import (
	"context"
	"database/sql/driver"
	"time"
)

type transaction struct {
	driver.Tx
	id     string
	connID string
	logger *logger
}

// Commit implement driver.Tx
func (tx *transaction) Commit() error {
	lvl, start := LevelDebug, time.Now()
	err := tx.Tx.Commit()

	if err != nil {
		lvl = LevelError
	}

	tx.logger.log(context.Background(), lvl, "Commit", start, err, tx.logData()...)

	return err
}

// Rollback implement driver.Tx
func (tx *transaction) Rollback() error {
	lvl, start := LevelDebug, time.Now()
	err := tx.Tx.Rollback()

	if err != nil {
		lvl = LevelError
	}

	tx.logger.log(context.Background(), lvl, "Rollback", start, err, tx.logData()...)

	return err
}

// logData default log data for transaction.
func (tx *transaction) logData() []dataFunc {
	return []dataFunc{
		tx.logger.withUID(tx.logger.opt.connIDFieldname, tx.connID),
		tx.logger.withUID(tx.logger.opt.txIDFieldname, tx.id),
	}
}
