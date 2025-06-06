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
)

// Generate is the main function that orchestrates the DAO generation process.
func Generate() error {
	// The command is run from the project root (e.g. maltose-quickstart),
	// so it will find the .env file in the current working directory.
	fmt.Println("🔎 Searching for .env file in the current directory...")
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf(".env file not found or failed to load. Please ensure a .env file with DB credentials exists in your project root. Error: %w", err)
	}
	fmt.Println("✔ .env file loaded successfully.")

	// Read DB config from environment variables
	dbInfo := DBInfo{
		DBType: os.Getenv("DB_TYPE"),
		Host:   os.Getenv("DB_HOST"),
		Port:   os.Getenv("DB_PORT"),
		User:   os.Getenv("DB_USER"),
		Pass:   os.Getenv("DB_PASS"),
		Name:   os.Getenv("DB_NAME"),
	}

	if dbInfo.DBType == "" || dbInfo.Host == "" || dbInfo.User == "" || dbInfo.Name == "" {
		return fmt.Errorf("one or more required database environment variables are not set in .env file (DB_TYPE, DB_HOST, DB_USER, DB_NAME, DB_PORT, DB_PASS)")
	}

	fmt.Printf("⚙️ Database config: [Type: %s, Host: %s, User: %s, DB: %s]\n", dbInfo.DBType, dbInfo.Host, dbInfo.User, dbInfo.Name)

	// Step 1: Connect to DB
	fmt.Println("⚡ Connecting to the database...")
	db, err := GetDBConnection(dbInfo)
	if err != nil {
		return err
	}
	fmt.Println("✔ Database connection successful.")

	// Step 2: Get all table schemas
	fmt.Println("🔍 Inspecting database schema...")
	tables, err := GetTables(db)
	if err != nil {
		return err
	}
	fmt.Printf("✔ Found %d tables.\n", len(tables))

	// Step 3: Generate files for each table
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

		// Define output path for the entity file
		outputPath := filepath.Join("internal", "model", "entity", fmt.Sprintf("%s.go", table.Name))

		fmt.Printf("  -> Generating %s\n", outputPath)
		if err := generateFile(data, tpl, outputPath); err != nil {
			return err
		}
	}
	fmt.Println("✔ Entity files generated successfully.")

	// Step 4: Generate DAO files
	fmt.Println(" H Generating dao files...")

	// Get go module path
	modulePath, err := getGoModulePath()
	if err != nil {
		return fmt.Errorf("failed to get go module path: %w", err)
	}

	// Parse templates
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

		// Generate internal dao file
		internalPath := filepath.Join("internal", "dao", "internal", fmt.Sprintf("%s.go", table.Name))
		fmt.Printf("  -> Generating %s\n", internalPath)
		if err := generateFile(data, internalTpl, internalPath); err != nil {
			return err
		}

		// Generate public dao file (if it doesn't exist)
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
	fmt.Println("✔ DAO files generated successfully.")

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
