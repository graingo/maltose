package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/spf13/cobra"
)

var moduleFlag string
var repoURLFlag string

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new [project-name]",
	Short: utils.Print("new_cmd_short"),
	Long:  utils.Print("new_cmd_long"),
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]
		repoURL := "https://github.com/graingo/maltose-quickstart.git"

		utils.PrintInfo("new_project_creation_start", utils.TplData{"ProjectName": projectName})
		utils.PrintInfo("new_project_template", utils.TplData{"RepoURL": repoURL})

		// 1. Clone the repository
		cloneCmd := exec.Command("git", "clone", repoURL, projectName)
		if err := cloneCmd.Run(); err != nil {
			return fmt.Errorf("cloning template repository failed: %w", err)
		}

		// 2. Remove the .git directory
		if err := os.RemoveAll(filepath.Join(projectName, ".git")); err != nil {
			return fmt.Errorf("failed to remove .git directory: %w", err)
		}

		// 3. Determine and update the module path
		// Get current working directory
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current working directory: %w", err)
		}
		// The new module path is the current working directory + project name
		modulePath := filepath.Join(filepath.Base(cwd), projectName)

		// Update go.mod
		gomodPath := filepath.Join(projectName, "go.mod")
		content, err := os.ReadFile(gomodPath)
		if err != nil {
			return fmt.Errorf("failed to read go.mod: %w", err)
		}
		re := regexp.MustCompile(`module\s+\S+`)
		newContent := re.ReplaceAllString(string(content), "module "+modulePath)
		if err := os.WriteFile(gomodPath, []byte(newContent), 0644); err != nil {
			return fmt.Errorf("failed to write updated go.mod: %w", err)
		}

		utils.PrintSuccess("new_project_success", utils.TplData{"ProjectName": projectName})
		utils.PrintInfo("new_project_module_path_set", utils.TplData{"ModulePath": modulePath})
		fmt.Println(utils.Print("new_project_get_started"))
		fmt.Println(utils.Printf("new_project_get_started_cd", utils.TplData{"ProjectName": projectName}))
		fmt.Println(utils.Print("new_project_get_started_tidy"))
		fmt.Println(utils.Print("new_project_get_started_run"))

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
