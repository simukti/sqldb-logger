package sqldblogger

import (
	"database/sql"
	"database/sql/driver"
)

// Open wrap given driver with logger and return sql.DB
func Open(dsn string, drv driver.Driver, lg Logger, opts ...Option) (*sql.DB, error) {
	cfg := &config{}
	setDefaultConfig(cfg)

	for _, opt := range opts {
		opt(cfg)
	}

	loggedConnector := &connector{
		dsn:    dsn,
		driver: drv,
		logger: &logger{logger: lg, cfg: cfg},
	}

	return sql.OpenDB(loggedConnector), nil
}
