package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/spf13/cobra"
)

var moduleFlag string
var repoURLFlag string

// newCmd represents the new command
var newCmd = &cobra.Command{
	Use:   "new <project-name>",
	Short: utils.Print("new_cmd_short"),
	Long:  utils.Print("new_cmd_long"),
	Args:  cobra.ExactArgs(1), // Requires exactly one argument: the project name
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]

		// Determine the repository URL to use
		repoURL := repoURLFlag
		if repoURL == "" {
			repoURL = "https://github.com/graingo/maltose-quickstart.git"
		}

		utils.PrintInfo("new_project_creation_start", utils.TplData{"ProjectName": projectName})
		utils.PrintInfo("new_project_template", utils.TplData{"RepoURL": repoURL})

		// 1. Clone the repository
		if err := runCommand("git", "clone", repoURL, projectName); err != nil {
			utils.PrintError("new_project_clone_failed", utils.TplData{"Error": err})
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
			// This is not a critical error, so we can just warn the user.
			// The user can remove it manually.
			utils.PrintError("new_project_git_remove_failed", utils.TplData{"Error": err})
			// os.Exit(1)
		}

		// 4. Read, replace, and write go.mod
		if newModulePath != "" {
			goModPath := filepath.Join(projectName, "go.mod")
			content, err := os.ReadFile(goModPath)
			if err != nil {
				utils.PrintError("new_project_go_mod_read_failed", utils.TplData{"Error": err})
				os.Exit(1)
			}

			// This is a simplistic replacement and might not cover all edge cases
			// It assumes the template's module path is known and consistent.
			// A more robust solution might use go mod edit.
			updatedContent := strings.Replace(string(content), "github.com/graingo/maltose-quickstart", newModulePath, 1)
			if err := os.WriteFile(goModPath, []byte(updatedContent), 0644); err != nil {
				utils.PrintError("new_project_go_mod_write_failed", utils.TplData{"Error": err})
				os.Exit(1)
			}
		}

		utils.PrintSuccess("new_project_success", utils.TplData{"ProjectName": projectName})
		if newModulePath != "" {
			utils.PrintInfo("new_project_module_path_set", utils.TplData{"ModulePath": newModulePath})
		}
		utils.PrintInfo("new_project_get_started", nil)
		utils.PrintInfo("new_project_get_started_cd", utils.TplData{"ProjectName": projectName})
		utils.PrintInfo("new_project_get_started_tidy", nil)
		utils.PrintInfo("new_project_get_started_run", nil)
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
