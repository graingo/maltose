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
	Short: "Create a new Maltose project.",
	Long: `Creates a new Maltose project by cloning the quickstart template repository.
It automatically replaces the module path in the new project's go.mod file.
`,
	Args: cobra.ExactArgs(1), // Requires exactly one argument: the project name
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]

		// Determine the repository URL to use
		repoURL := repoURLFlag
		if repoURL == "" {
			repoURL = "https://github.com/graingo/maltose-quickstart.git"
		}

		utils.PrintInfo("newProjectCreationStart", map[string]interface{}{"ProjectName": projectName})
		utils.PrintInfo("newProjectTemplate", map[string]interface{}{"RepoURL": repoURL})

		// 1. Clone the repository
		if err := runCommand("git", "clone", repoURL, projectName); err != nil {
			utils.PrintError("newProjectCloneFailed", map[string]interface{}{"Error": err})
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
			utils.PrintError("newProjectGitRemoveFailed", map[string]interface{}{"Error": err})
			// os.Exit(1)
		}

		// 4. Read, replace, and write go.mod
		if newModulePath != "" {
			goModPath := filepath.Join(projectName, "go.mod")
			content, err := os.ReadFile(goModPath)
			if err != nil {
				utils.PrintError("newProjectGoModReadFailed", map[string]interface{}{"Error": err})
				os.Exit(1)
			}

			// This is a simplistic replacement and might not cover all edge cases
			// It assumes the template's module path is known and consistent.
			// A more robust solution might use go mod edit.
			updatedContent := strings.Replace(string(content), "github.com/graingo/maltose-quickstart", newModulePath, 1)
			if err := os.WriteFile(goModPath, []byte(updatedContent), 0644); err != nil {
				utils.PrintError("newProjectGoModWriteFailed", map[string]interface{}{"Error": err})
				os.Exit(1)
			}
		}

		utils.PrintSuccess("newProjectSuccess", map[string]interface{}{"ProjectName": projectName})
		if newModulePath != "" {
			utils.PrintInfo("newProjectModulePathSet", map[string]interface{}{"ModulePath": newModulePath})
		}
		utils.PrintInfo("newProjectGetStarted", nil)
		utils.PrintInfo("newProjectGetStartedCD", map[string]interface{}{"ProjectName": projectName})
		utils.PrintInfo("newProjectGetStartedTidy", nil)
		utils.PrintInfo("newProjectGetStartedRun", nil)
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
