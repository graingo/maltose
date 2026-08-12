package mdb

import (
	"net/url"
	"testing"

	driverMySQL "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
	gormMySQL "gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
)

func TestCreateDriverEscapesConnectionParameters(t *testing.T) {
	t.Run("mysql", func(t *testing.T) {
		driver, err := createDriver(&Config{
			Type:     "mysql",
			Host:     "2001:db8::1",
			Port:     "3306",
			User:     "user@tenant",
			Password: "p@ss/word",
			DBName:   "db/name",
		})
		require.NoError(t, err)

		dialector := driver.(*gormMySQL.Dialector)
		parsed, err := driverMySQL.ParseDSN(dialector.Config.DSN)
		require.NoError(t, err)
		require.Equal(t, "[2001:db8::1]:3306", parsed.Addr)
		require.Equal(t, "user@tenant", parsed.User)
		require.Equal(t, "p@ss/word", parsed.Passwd)
		require.Equal(t, "db/name", parsed.DBName)
	})

	t.Run("postgres", func(t *testing.T) {
		driver, err := createDriver(&Config{
			Type:     "postgres",
			Host:     "2001:db8::1",
			Port:     "5432",
			User:     "user@tenant",
			Password: "p@ss word",
			DBName:   "db name",
		})
		require.NoError(t, err)

		dialector := driver.(*postgres.Dialector)
		parsed, err := url.Parse(dialector.Config.DSN)
		require.NoError(t, err)
		require.Equal(t, "user@tenant", parsed.User.Username())
		password, ok := parsed.User.Password()
		require.True(t, ok)
		require.Equal(t, "p@ss word", password)
		require.Equal(t, "db name", parsed.Path[1:])
		require.Equal(t, "disable", parsed.Query().Get("sslmode"))
	})
}
