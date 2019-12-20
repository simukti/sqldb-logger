// Package sqldblogger act as thin and transparent logger without having to change existing *sql.DB usage.
package sqldblogger

import (
	"database/sql"
	"database/sql/driver"
)

// OpenDriver wrap given driver with logger and return *sql.DB.
func OpenDriver(dsn string, drv driver.Driver, lg Logger, opt ...Option) *sql.DB {
	opts := &options{}
	setDefaultOptions(opts)

	for _, o := range opt {
		o(opts)
	}

	conn := &connector{
		dsn:    dsn,
		driver: drv,
		logger: &logger{logger: lg, opt: opts},
	}

	return sql.OpenDB(conn)
}
