package gen

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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
		return nil, fmt.Errorf("unsupported database type: %s", info.DBType)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return db, nil
}

// GetTables retrieves all tables and their column information from the database.
func GetTables(db *gorm.DB) ([]TableInfo, error) {
	tableNames, err := db.Migrator().GetTables()
	if err != nil {
		return nil, fmt.Errorf("failed to get tables from database: %w", err)
	}

	var tables []TableInfo
	for _, name := range tableNames {
		columns, err := db.Migrator().ColumnTypes(name)
		if err != nil {
			return nil, fmt.Errorf("failed to get columns for table %s: %w", name, err)
		}

		tables = append(tables, TableInfo{
			Name:    name,
			Columns: columns,
		})
	}

	return tables, nil
}
