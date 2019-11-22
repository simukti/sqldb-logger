package sqldblogger

import (
	"context"
	"database/sql/driver"
	"time"
)

type transaction struct {
	tx     driver.Tx
	logger *logger
}

func (t *transaction) Commit() error {
	lvl, start := LevelDebug, time.Now()
	err := t.tx.Commit()

	if err != nil {
		lvl = LevelError
	}

	t.logger.log(context.Background(), lvl, "Commit", start, err)

	return err
}

func (t *transaction) Rollback() error {
	lvl, start := LevelDebug, time.Now()
	err := t.tx.Rollback()

	if err != nil {
		lvl = LevelError
	}

	t.logger.log(context.Background(), lvl, "Rollback", start, err)

	return err
}
