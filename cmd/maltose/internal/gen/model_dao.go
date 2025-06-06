// Package gen handles the logic for DAO layer generation.
package gen

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"
	"github.com/joho/godotenv"
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

// GenerateModel generates only the entity files.
func GenerateModel() error {
	if err := initDB(); err != nil {
		return err
	}

	fmt.Println("H Generating entity files...")
	for _, table := range tables {
		structName := strcase.ToCamel(inflection.Singular(table.Name))
		data := DaoTplData{
			TableName:  table.Name,
			StructName: structName,
			Columns:    table.Columns,
		}

		outputPath := filepath.Join("internal", "model", "entity", fmt.Sprintf("%s.go", table.Name))

		fmt.Printf("  -> Generating %s\n", outputPath)
		if err := generateFile(outputPath, "entity", TplGenEntity, data); err != nil {
			return err
		}
	}
	return nil
}

// GenerateDao generates only the DAO files.
func GenerateDao() error {
	if err := initDB(); err != nil {
		return err
	}

	fmt.Println("H Generating dao files...")
	modulePath, err := getGoModulePath()
	if err != nil {
		return fmt.Errorf("failed to get go module path: %w", err)
	}

	for _, table := range tables {
		structName := strcase.ToCamel(inflection.Singular(table.Name))
		daoName := structName + "Dao"
		data := DaoTplData{
			TableName:   table.Name,
			StructName:  structName,
			PackageName: modulePath,
			DaoName:     daoName,
		}

		internalPath := filepath.Join("internal", "dao", "internal", fmt.Sprintf("%s.go", table.Name))
		fmt.Printf("  -> Generating %s\n", internalPath)
		if err := generateFile(internalPath, "daoInternal", TplGenDaoInternal, data); err != nil {
			return err
		}

		daoPath := filepath.Join("internal", "dao", fmt.Sprintf("%s.go", table.Name))
		if _, err := os.Stat(daoPath); os.IsNotExist(err) {
			fmt.Printf("  -> Generating %s\n", daoPath)
			if err := generateFile(daoPath, "dao", TplGenDao, data); err != nil {
				return err
			}
		} else {
			fmt.Printf("  -> Skipping %s (already exists)\n", daoPath)
		}
	}
	return nil
}

func getGoModulePath() (string, error) {
	cmd := exec.Command("go", "list", "-m")
	var out strings.Builder
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}
