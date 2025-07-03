package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/graingo/maltose/errors/merror"
	"github.com/spf13/cobra"
)

var moduleFlag string
var repoURLFlag string

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new [project-name]",
	Short: "Create a new Maltose project.",
	Long:  "Creates a new Maltose project by cloning the quickstart template repository.\nIt automatically replaces the module path in the new project's go.mod file.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]
		repoURL := "https://github.com/graingo/maltose-quickstart.git"

		utils.PrintInfo("Creating a new Maltose project in './{{.ProjectName}}'...", utils.TplData{"ProjectName": projectName})
		utils.PrintInfo("Using template from: {{.RepoURL}}", utils.TplData{"RepoURL": repoURL})

		// 1. Clone the repository
		cloneCmd := exec.Command("git", "clone", repoURL, projectName)
		if err := cloneCmd.Run(); err != nil {
			return merror.Wrap(err, "cloning template repository failed")
		}

		// 2. Remove the .git directory
		if err := os.RemoveAll(filepath.Join(projectName, ".git")); err != nil {
			return merror.Wrap(err, "failed to remove .git directory")
		}

		// 3. Determine and update the module path
		// Get current working directory
		cwd, err := os.Getwd()
		if err != nil {
			return merror.Wrap(err, "failed to get current working directory")
		}
		// The new module path is the current working directory + project name
		modulePath := filepath.Join(filepath.Base(cwd), projectName)

		// Update go.mod
		gomodPath := filepath.Join(projectName, "go.mod")
		content, err := os.ReadFile(gomodPath)
		if err != nil {
			return merror.Wrap(err, "failed to read go.mod")
		}
		re := regexp.MustCompile(`module\s+\S+`)
		newContent := re.ReplaceAllString(string(content), "module "+modulePath)
		if err := os.WriteFile(gomodPath, []byte(newContent), 0644); err != nil {
			return merror.Wrap(err, "failed to write updated go.mod")
		}

		utils.PrintSuccess("Project '{{.ProjectName}}' created successfully.", utils.TplData{"ProjectName": projectName})
		utils.PrintInfo("Module path set to '{{.ModulePath}}'.", utils.TplData{"ModulePath": modulePath})
		fmt.Println(utils.Print("\nTo get started:"))
		fmt.Println(utils.Printf("  cd {{.ProjectName}}", utils.TplData{"ProjectName": projectName}))
		fmt.Println(utils.Print("  go mod tidy"))
		fmt.Println(utils.Print("go run main.go"))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().StringVar(&moduleFlag, "module", "", "Specify the Go module path for the new project.")
	newCmd.Flags().StringVar(&repoURLFlag, "repo-url", "", "Specify a custom git repository URL for the project template.")
}

// runCommand is a helper to execute external commands
func runCommand(name string, arg ...string) error {
	cmd := exec.Command(name, arg...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
