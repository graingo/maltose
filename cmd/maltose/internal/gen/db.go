package gen

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// shared state for generation
var (
	db     *gorm.DB
	tables []TableInfo
)

// initDB loads .env, connects to the database, and gets table schemas.
// It uses a shared state to avoid reconnecting if called multiple times.
func initDB() error {
	if db != nil {
		return nil // Already initialized
	}

	// Check for .env file and create an example if it doesn't exist.
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		fmt.Println("'.env' file not found. Creating a '.env.example' for you.")
		err := createEnvExample()
		if err != nil {
			return fmt.Errorf("failed to create .env.example: %w", err)
		}
		return fmt.Errorf("please copy '.env.example' to '.env' and fill in your database credentials")
	}

	fmt.Println("🔎 Loading .env file...")
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf(".env file found but failed to load: %w", err)
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
	fmt.Println("⚡ Connecting to the database...")
	db, err = GetDBConnection(dbInfo)
	if err != nil {
		return err
	}

	fmt.Println("🔍 Inspecting database schema...")
	tables, err = GetTables(db)
	if err != nil {
		return err
	}
	fmt.Printf("✔ Found %d tables.\n", len(tables))
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
