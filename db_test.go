package sqldblogger

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestConnector_Connect(t *testing.T) {
	cfg := &config{}
	setDefaultConfig(cfg)
	dbmock, _, _ := sqlmock.New()
	con := &connector{
		dsn:    "sqlmock_db_0", // this 0 to use sql mock counter
		driver: dbmock.Driver(),
		logger: &logger{
			logger: &NullLogger{},
			cfg:    cfg,
		},
	}

	c, err := con.Connect(context.TODO())
	assert.NoError(t, err)
	_, ok := c.(driver.Conn)
	assert.True(t, ok)
	_ = dbmock.Close()
}

func TestConnector_ConnectError(t *testing.T) {
	cfg := &config{}
	setDefaultConfig(cfg)
	dbmock, _, _ := sqlmock.New()
	con := &connector{
		dsn:    "sqlmock_err_always",
		driver: dbmock.Driver(),
		logger: &logger{
			logger: &NullLogger{},
			cfg:    cfg,
		},
	}

	_, err := con.Connect(context.TODO())
	assert.Error(t, err)
}

func TestConnector_Driver(t *testing.T) {
	dbmock, _, _ := sqlmock.New()
	db, err := Open("sqlmock_err", dbmock.Driver(), &NullLogger{})
	assert.NoError(t, err)
	_, ok := interface{}(db).(*sql.DB)
	assert.True(t, ok)
	assert.Equal(t, dbmock.Driver(), db.Driver())
	_ = dbmock.Close()
}

func TestOpen(t *testing.T) {
	dbmock, _, _ := sqlmock.New()
	db, err := Open("sqlmock_db_0", dbmock.Driver(), &NullLogger{})
	assert.NoError(t, err)
	_, ok := interface{}(db).(*sql.DB)
	assert.True(t, ok)
	_ = dbmock.Close()
}

func TestOpenWithOptions(t *testing.T) {
	dbmock, _, _ := sqlmock.New()
	db, err := Open("sqlmock_db_0", dbmock.Driver(), &NullLogger{}, WithErrorFieldname("errtest"), WithMinimumLevel(LevelNotice))
	assert.NoError(t, err)
	_, ok := interface{}(db).(*sql.DB)
	assert.True(t, ok)
	_ = dbmock.Close()
}
