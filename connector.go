package sqldblogger

import (
	"context"
	"database/sql/driver"
	"time"
)

// connector is a wrapped connector to a given driver and should implements:
// - driver.Connector
type connector struct {
	dsn    string
	driver driver.Driver
	logger *logger
}

// Connect implement driver.Connector which will open new db connection if none exist
func (c *connector) Connect(ctx context.Context) (driver.Conn, error) {
	start := time.Now()
	driverConn, err := c.driver.Open(c.dsn)

	if err != nil {
		c.logger.log(ctx, LevelError, "Connect", start, err)
		return nil, err
	}

	c.logger.log(ctx, LevelDebug, "Connect", start, err)

	return &connection{driverConn: driverConn, logger: c.logger}, nil
}

// Driver implement driver.Connector
func (c *connector) Driver() driver.Driver { return c.driver }
