// Package gendao handles the logic for DAO layer generation.
package gendao

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

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
	fmt.Println("🔎 Searching for .env file...")
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf(".env file not found or failed to load: %w", err)
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

// GenerateModel generates only the entity files.
func GenerateModel() error {
	if err := initDB(); err != nil {
		return err
	}

	fmt.Println(" H Generating entity files...")
	tpl, err := template.New("entity").Funcs(funcMap).Parse(entityTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse entity template: %w", err)
	}

	for _, table := range tables {
		data := templateData{
			TableName:  table.Name,
			StructName: toCamel(inflection.Singular(table.Name)),
			Columns:    table.Columns,
		}

		outputPath := filepath.Join("internal", "model", "entity", fmt.Sprintf("%s.go", table.Name))

		fmt.Printf("  -> Generating %s\n", outputPath)
		if err := generateFile(data, tpl, outputPath); err != nil {
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

	fmt.Println(" H Generating dao files...")
	modulePath, err := getGoModulePath()
	if err != nil {
		return fmt.Errorf("failed to get go module path: %w", err)
	}

	internalTpl, err := template.New("daoInternal").Funcs(funcMap).Parse(daoInternalTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse dao internal template: %w", err)
	}
	daoTpl, err := template.New("dao").Funcs(funcMap).Parse(daoTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse dao template: %w", err)
	}

	for _, table := range tables {
		structName := toCamel(inflection.Singular(table.Name))
		data := templateData{
			TableName:       table.Name,
			StructName:      structName,
			PackageName:     modulePath,
			InternalDaoName: "s" + structName + "Dao",
			DaoName:         structName + "Dao",
		}

		internalPath := filepath.Join("internal", "dao", "internal", fmt.Sprintf("%s.go", table.Name))
		fmt.Printf("  -> Generating %s\n", internalPath)
		if err := generateFile(data, internalTpl, internalPath); err != nil {
			return err
		}

		daoPath := filepath.Join("internal", "dao", fmt.Sprintf("%s.go", table.Name))
		if _, err := os.Stat(daoPath); os.IsNotExist(err) {
			fmt.Printf("  -> Generating %s\n", daoPath)
			if err := generateFile(data, daoTpl, daoPath); err != nil {
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
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}
