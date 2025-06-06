package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var moduleFlag string

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new <project-name>",
	Short: "Create a new Maltose project.",
	Long: `Creates a new Maltose project by cloning the quickstart template repository.
It automatically replaces the module path in the new project's go.mod file.
`,
	Args: cobra.ExactArgs(1), // Requires exactly one argument: the project name
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		repoURL := "https://github.com/graingo/maltose-quickstart.git"

		fmt.Printf("Creating a new Maltose project in './%s'...\n", projectName)

		// 1. Clone the repository
		cloneCmd := exec.Command("git", "clone", repoURL, projectName)
		if err := cloneCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error cloning template repository: %v\n", err)
			os.Exit(1)
		}

		// 2. Determine the new module path
		newModulePath := moduleFlag
		if newModulePath == "" {
			newModulePath = projectName
		}

		// 3. Remove the .git directory
		gitPath := filepath.Join(projectName, ".git")
		if err := os.RemoveAll(gitPath); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to remove .git directory: %v\n", err)
			os.Exit(1)
		}

		// 4. Read, replace, and write go.mod
		goModPath := filepath.Join(projectName, "go.mod")
		oldModulePath := "github.com/graingo/maltose-quickstart"

		input, err := os.ReadFile(goModPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read go.mod: %v\n", err)
			os.Exit(1)
		}

		output := strings.Replace(string(input), oldModulePath, newModulePath, 1)
		if err = os.WriteFile(goModPath, []byte(output), 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write updated go.mod: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Project '%s' created successfully.\n", projectName)
		fmt.Printf("Module path set to '%s'.\n", newModulePath)
		fmt.Println("\nTo get started:")
		fmt.Printf("  cd %s\n", projectName)
		fmt.Println("  go mod tidy")
		fmt.Println("  go run .")
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().StringVar(&moduleFlag, "module", "", "Specify the Go module path for the new project.")
}
