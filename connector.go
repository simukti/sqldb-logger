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
	start, id := time.Now(), c.logger.opt.uidGenerator.UniqueID()
	logID := c.logger.withUID(c.logger.opt.connIDFieldname, id)
	conn, err := c.driver.Open(c.dsn)

	if err != nil {
		c.logger.log(ctx, LevelError, "Connect", start, err, logID)
		return nil, err
	}

	// here, we're using the overriding log level to guarantee that it's included
	// despite really being informational... this will allow us to mute other stuff
	// we don't want to listen to
	c.logger.log(ctx, LevelInfo, "Connect", start, err, logID)

	return &connection{Conn: conn, logger: c.logger, id: id}, nil
}

// Driver implement driver.Connector
func (c *connector) Driver() driver.Driver { return c.driver }
