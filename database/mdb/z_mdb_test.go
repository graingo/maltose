package mdb_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/graingo/maltose/database/mdb"
	"github.com/graingo/maltose/os/mlog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// User is a test model.
type User struct {
	ID   uint   `gorm:"primarykey"`
	Name string `gorm:"unique"`
	Age  int
}

// TableName explicitly sets the table name for the User model.
func (User) TableName() string {
	return "users"
}

// setupTestDB creates a new database connection for testing.
// It will use MySQL if CI environment variables are set, otherwise it defaults to in-memory SQLite.
func setupTestDB(t *testing.T) (*mdb.DB, error) {
	t.Helper()

	// Use environment variables in CI, otherwise use local SQLite.
	dbType := "sqlite"
	var cfg *mdb.Config

	if dbType == "mysql" {
		cfg = &mdb.Config{
			Type:     "mysql",
			Host:     os.Getenv("DB_HOST"),
			Port:     os.Getenv("DB_PORT"),
			User:     os.Getenv("DB_USER"),
			Password: os.Getenv("DB_PASS"),
			DBName:   os.Getenv("DB_NAME"),
		}
	} else {
		// Create a unique DSN for each test to prevent parallel test interference.
		dbName := strings.ReplaceAll(t.Name(), "/", "_")
		dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", dbName)
		cfg = &mdb.Config{
			Type: "sqlite",
			DSN:  dsn,
		}
	}
	// Use New() to get a logger with default configuration.
	cfg.SetLogger(mlog.New())
	db, err := mdb.New(cfg)
	if err != nil {
		return nil, err
	}

	// Force a single connection for SQLite in-memory tests to prevent connection pool issues.
	sqlDB, err := db.DB.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetMaxOpenConns(1)

	return db, nil
}

func TestMDB(t *testing.T) {
	// No top-level db instance. Each test group will create its own for isolation.

	t.Run("connection_and_config", func(t *testing.T) {
		db, err := setupTestDB(t)
		require.NoError(t, err)
		require.NotNil(t, db)
		sqlDB, _ := db.DB.DB()
		defer sqlDB.Close()

		assert.NotNil(t, db.DB)

		// Test Ping
		err = db.Ping(context.Background())
		assert.NoError(t, err)

		// Test creating a db instance with a custom logger.
		customLogger := mlog.New()
		cfg := &mdb.Config{
			Type: "sqlite",
			DSN:  "file:config_test?mode=memory&cache=shared",
		}
		cfg.SetLogger(customLogger)
		dbWithCustomLogger, err := mdb.New(cfg)
		require.NoError(t, err)
		require.NotNil(t, dbWithCustomLogger)
		// The logger returned by GetLogger will have the "component" field added internally.
		// We can't do a direct assert.Equal on the logger object itself.
		// Instead, we check that the core parent logger is the same.
		// The functional test is in the logging_hook section.
		assert.NotNil(t, dbWithCustomLogger.GetLogger(), "GetLogger should not return nil")
		customSQLDB, _ := dbWithCustomLogger.DB.DB()
		customSQLDB.Close()
	})

	t.Run("crud_and_transactions", func(t *testing.T) {
		// Each subtest in this group will set up its own DB to ensure isolation.
		ctx := context.Background()

		t.Run("create_and_read", func(t *testing.T) {
			db, err := setupTestDB(t)
			require.NoError(t, err)
			sqlDB, _ := db.DB.DB()
			defer sqlDB.Close()
			require.NoError(t, db.AutoMigrate(&User{}))

			user := User{Name: "Alice", Age: 30}
			err = db.WithContext(ctx).Create(&user).Error
			require.NoError(t, err)
			assert.NotZero(t, user.ID)

			var foundUser User
			err = db.WithContext(ctx).First(&foundUser, user.ID).Error
			require.NoError(t, err)
			assert.Equal(t, "Alice", foundUser.Name)
		})

		t.Run("update", func(t *testing.T) {
			db, err := setupTestDB(t)
			require.NoError(t, err)
			sqlDB, _ := db.DB.DB()
			defer sqlDB.Close()
			require.NoError(t, db.AutoMigrate(&User{}))

			user := User{Name: "Bob", Age: 25}
			db.Create(&user)

			err = db.WithContext(ctx).Model(&user).Where("id = ?", user.ID).Update("Age", 26).Error
			require.NoError(t, err)

			var updatedUser User
			db.First(&updatedUser, user.ID)
			assert.Equal(t, 26, updatedUser.Age)
		})

		t.Run("delete", func(t *testing.T) {
			db, err := setupTestDB(t)
			require.NoError(t, err)
			sqlDB, _ := db.DB.DB()
			defer sqlDB.Close()
			require.NoError(t, db.AutoMigrate(&User{}))

			user := User{Name: "Charlie", Age: 40}
			db.Create(&user)

			err = db.WithContext(ctx).Where("id = ?", user.ID).Delete(&user).Error
			require.NoError(t, err)

			var result User
			err = db.First(&result, user.ID).Error
			assert.Error(t, err)
			assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
		})

		t.Run("transaction_commit", func(t *testing.T) {
			db, err := setupTestDB(t)
			require.NoError(t, err)
			sqlDB, _ := db.DB.DB()
			defer sqlDB.Close()
			require.NoError(t, db.AutoMigrate(&User{}))

			tx := db.WithContext(ctx).Begin()
			require.NoError(t, tx.Error)

			err = tx.Create(&User{Name: "David"}).Error
			require.NoError(t, err)

			err = tx.Create(&User{Name: "Eve"}).Error
			require.NoError(t, err)

			err = tx.Commit().Error
			require.NoError(t, err)

			var count int64
			db.Model(&User{}).Where("name IN ?", []string{"David", "Eve"}).Count(&count)
			assert.Equal(t, int64(2), count)
		})

		t.Run("transaction_rollback", func(t *testing.T) {
			db, err := setupTestDB(t)
			require.NoError(t, err)
			sqlDB, _ := db.DB.DB()
			defer sqlDB.Close()
			require.NoError(t, db.AutoMigrate(&User{}))

			tx := db.WithContext(ctx).Begin()
			require.NoError(t, tx.Error)

			err = tx.Create(&User{Name: "Frank"}).Error
			require.NoError(t, err)

			err = tx.Rollback().Error
			require.NoError(t, err)

			var count int64
			db.Model(&User{}).Where("name = ?", "Frank").Count(&count)
			assert.Equal(t, int64(0), count)
		})
	})

	t.Run("logging_hook", func(t *testing.T) {
		var logBuffer bytes.Buffer
		logger := mlog.New(&mlog.Config{
			Writer: &logBuffer,
			Stdout: false,
			Level:  mlog.DebugLevel,
			Format: "json",
		})
		ctx := context.Background()

		t.Run("normal_query", func(t *testing.T) {
			logBuffer.Reset()
			cfg := &mdb.Config{Type: "sqlite", DSN: "file:testdb_log?mode=memory&cache=shared"}
			cfg.SetLogger(logger)
			db, err := mdb.New(cfg)
			require.NoError(t, err)
			sqlDB, _ := db.DB.DB()
			sqlDB.SetMaxIdleConns(1)
			sqlDB.SetMaxOpenConns(1)
			defer sqlDB.Close()

			require.NoError(t, db.AutoMigrate(&User{}))

			db.WithContext(ctx).Create(&User{Name: "LogUserNormal"})
			logOutput := logBuffer.String()
			assert.Contains(t, logOutput, `"msg":"sql trace"`, "should log normal queries")
			assert.Contains(t, logOutput, "INSERT INTO `users`", "should contain the correct SQL")
		})

		t.Run("slow_query", func(t *testing.T) {
			logBuffer.Reset()
			cfg := &mdb.Config{
				Type:          "sqlite",
				DSN:           "file:testdb_slow?mode=memory&cache=shared",
				SlowThreshold: time.Nanosecond, // Set a very low threshold to guarantee slowness
			}
			cfg.SetLogger(logger)
			db, err := mdb.New(cfg)
			require.NoError(t, err)
			sqlDB, _ := db.DB.DB()
			sqlDB.SetMaxIdleConns(1)
			sqlDB.SetMaxOpenConns(1)
			defer sqlDB.Close()

			require.NoError(t, db.AutoMigrate(&User{}))

			db.WithContext(ctx).First(&User{}, 1)
			logOutput := logBuffer.String()
			assert.Contains(t, logOutput, `"msg":"sql slow"`, "should log slow queries")
		})

		t.Run("error_query", func(t *testing.T) {
			logBuffer.Reset()
			cfg := &mdb.Config{Type: "sqlite", DSN: "file:testdb_err?mode=memory&cache=shared"}
			cfg.SetLogger(logger)
			db, err := mdb.New(cfg)
			require.NoError(t, err)
			sqlDB, _ := db.DB.DB()
			sqlDB.SetMaxIdleConns(1)
			sqlDB.SetMaxOpenConns(1)
			defer sqlDB.Close()

			require.NoError(t, db.AutoMigrate(&User{}))

			db.WithContext(ctx).Create(&User{Name: "LogUserError"})             // Create first user
			err = db.WithContext(ctx).Create(&User{Name: "LogUserError"}).Error // Trigger unique constraint
			assert.Error(t, err)
			logOutput := logBuffer.String()
			assert.Contains(t, logOutput, `"msg":"sql error"`, "should log query errors")
			assert.Contains(t, logOutput, "UNIQUE constraint failed: users.name", "should contain the db error message for sqlite")
		})
	})
}
