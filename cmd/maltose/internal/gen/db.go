package gen

import (
	"errors"
	"fmt"
	"os"

	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/graingo/maltose/errors/merror"
	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ErrEnvFileNeedUpdate indicates that the .env file is missing and needs to be created from .env.example.
var ErrEnvFileNeedUpdate = errors.New("env file need update")

// shared state for generation
var (
	db     *gorm.DB
	tables []TableInfo
)

// initDB ensures the database connection is initialized.
func initDB() error {
	if db != nil {
		return nil // Already initialized
	}

	// Load .env file if it exists
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		if err := createEnvExample(); err != nil {
			return err
		}
		utils.PrintNotice("'.env' file not found. Creating a '.env.example' for you.\nPlease copy '.env.example' to '.env' and fill in your database credentials", nil)
		return ErrEnvFileNeedUpdate
	}

	utils.PrintInfo("🔎 Loading .env file...", nil)
	if err := godotenv.Load(); err != nil {
		return merror.Wrap(err, "error loading .env file")
	}

	dbInfo := DBInfo{
		DBType: os.Getenv("DB_TYPE"),
		Host:   os.Getenv("DB_HOST"),
		Port:   os.Getenv("DB_PORT"),
		User:   os.Getenv("DB_USER"),
		Pass:   os.Getenv("DB_PASS"),
		Name:   os.Getenv("DB_NAME"),
	}

	var err error
	utils.PrintInfo("⚡ Connecting to the database...", nil)
	db, err = GetDBConnection(dbInfo)
	if err != nil {
		return err
	}

	// Inspect the database schema
	utils.PrintInfo("🔍 Inspecting database schema...", nil)
	tables, err = GetTables(db)
	if err != nil {
		return err
	}
	utils.PrintInfo("✔ Found {{.Count}} tables.", utils.TplData{"Count": len(tables)})
	return nil
}

func createEnvExample() error {
	content := `# General Database Settings (数据库通用设置)
DB_TYPE=mysql

# MySQL Settings (MySQL 配置)
DB_HOST=127.0.0.1
DB_PORT=3306
DB_USER=root
DB_PASS=
DB_NAME=your_database_name
`
	return os.WriteFile(".env.example", []byte(content), 0644)
}

// DBInfo holds all the necessary information for a database connection.
type DBInfo struct {
	DBType string
	Host   string
	Port   string
	User   string
	Pass   string
	Name   string
}

// TableInfo holds information about a database table.
type TableInfo struct {
	Name    string
	Columns []gorm.ColumnType
}

// GetDBConnection creates and returns a GORM DB instance.
func GetDBConnection(info DBInfo) (*gorm.DB, error) {
	var dialector gorm.Dialector

	switch info.DBType {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			info.User, info.Pass, info.Host, info.Port, info.Name)
		dialector = mysql.Open(dsn)
	case "pg", "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Shanghai",
			info.Host, info.User, info.Pass, info.Name, info.Port)
		dialector = postgres.Open(dsn)
	default:
		return nil, merror.Newf("unsupported database type: %s", info.DBType)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, merror.Wrap(err, "failed to connect to database")
	}
	return db, nil
}

// GetTables retrieves all tables and their column information from the database.
func GetTables(db *gorm.DB) ([]TableInfo, error) {
	tableNames, err := db.Migrator().GetTables()
	if err != nil {
		return nil, merror.Wrap(err, "failed to get tables from database")
	}

	var tables []TableInfo
	for _, name := range tableNames {
		columns, err := db.Migrator().ColumnTypes(name)
		if err != nil {
			return nil, merror.Wrapf(err, "failed to get columns for table %s", name)
		}

		tables = append(tables, TableInfo{
			Name:    name,
			Columns: columns,
		})
	}

	return tables, nil
}
