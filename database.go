package sqldblogger

import (
	"database/sql"
	"database/sql/driver"
)

// Open wrap given driver with logger and return sql.DB
func Open(dsn string, drv driver.Driver, lg Logger, opt ...Option) (*sql.DB, error) {
	optObj := &options{}
	setDefaultOptions(optObj)

	for _, o := range opt {
		o(optObj)
	}

	loggedConnector := &connector{
		dsn:    dsn,
		driver: drv,
		logger: &logger{logger: lg, opt: optObj},
	}

	return sql.OpenDB(loggedConnector), nil
}
