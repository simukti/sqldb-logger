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

	tx.logger.log(context.Background(), lvl, "Commit", start, err, tx.txIDs()...)

	return err
}

// Rollback implement driver.Tx
func (tx *transaction) Rollback() error {
	lvl, start := LevelDebug, time.Now()
	err := tx.Tx.Rollback()

	if err != nil {
		lvl = LevelError
	}

	tx.logger.log(context.Background(), lvl, "Rollback", start, err, tx.txIDs()...)

	return err
}

const txID = "tx.id"

func (tx *transaction) txIDs() []dataFunc {
	return []dataFunc{
		tx.logger.withUID(connID, tx.connID),
		tx.logger.withUID(txID, tx.id),
	}
}
