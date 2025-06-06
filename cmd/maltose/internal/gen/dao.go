package gen

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/jinzhu/inflection"
)

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
