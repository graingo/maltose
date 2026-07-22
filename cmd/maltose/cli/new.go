package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/graingo/maltose/errors/merror"
	"github.com/spf13/cobra"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
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
		if repoURLFlag != "" {
			repoURL = repoURLFlag
		}
		modulePath := moduleFlag
		if modulePath == "" {
			modulePath = filepath.ToSlash(filepath.Clean(projectName))
		}
		if err := module.CheckPath(modulePath); err != nil {
			return merror.Wrap(err, "invalid Go module path")
		}
		cwd, err := os.Getwd()
		if err != nil {
			return merror.Wrap(err, "failed to get current working directory")
		}
		target, err := projectTarget(cwd, projectName)
		if err != nil {
			return err
		}

		utils.PrintInfo("🚀 Creating new Maltose project at './{{.ProjectName}}'...", utils.TplData{"ProjectName": projectName})
		utils.PrintInfo("📥 Cloning project template from {{.RepoURL}}...", utils.TplData{"RepoURL": repoURL})

		// 1. Clone the repository
		cloneCmd := exec.Command("git", "clone", "--", repoURL, target)
		cloneCmd.Stdout = cmd.OutOrStdout()
		cloneCmd.Stderr = cmd.ErrOrStderr()
		if err := cloneCmd.Run(); err != nil {
			return merror.Wrap(err, "cloning template repository failed")
		}

		// 2. Remove the .git directory
		if err := os.RemoveAll(filepath.Join(target, ".git")); err != nil {
			return merror.Wrap(err, "failed to remove .git directory")
		}

		// Update go.mod
		gomodPath := filepath.Join(target, "go.mod")
		content, err := os.ReadFile(gomodPath)
		if err != nil {
			return merror.Wrap(err, "failed to read go.mod")
		}
		goMod, err := modfile.Parse(gomodPath, content, nil)
		if err != nil {
			return merror.Wrap(err, "failed to parse go.mod")
		}
		if err := goMod.AddModuleStmt(modulePath); err != nil {
			return merror.Wrap(err, "failed to update module path")
		}
		newContent, err := goMod.Format()
		if err != nil {
			return merror.Wrap(err, "failed to format go.mod")
		}
		if err := os.WriteFile(gomodPath, newContent, 0644); err != nil {
			return merror.Wrap(err, "failed to write updated go.mod")
		}

		utils.PrintSuccess("✅ Successfully created project '{{.ProjectName}}'.", utils.TplData{"ProjectName": projectName})
		utils.PrintInfo("🔧 Module path has been set to '{{.ModulePath}}'.", utils.TplData{"ModulePath": modulePath})
		fmt.Println(utils.Print("\n👉 To get started, run:\n"))
		fmt.Println(utils.Printf("  cd {{.ProjectName}}", utils.TplData{"ProjectName": projectName}))
		fmt.Println(utils.Print("  go mod tidy"))
		fmt.Println(utils.Print("  go run main.go"))

		return nil
	},
}

func init() {
	rootCmd.AddCommand(newCmd)
	newCmd.Flags().StringVar(&moduleFlag, "module", "", "Specify the Go module path for the new project.")
	newCmd.Flags().StringVar(&repoURLFlag, "repo-url", "", "Specify a custom git repository URL for the project template.")
}

func projectTarget(cwd, projectName string) (string, error) {
	if strings.TrimSpace(projectName) == "" || filepath.IsAbs(projectName) {
		return "", merror.New("project name must be a non-empty relative path")
	}
	cleanName := filepath.Clean(projectName)
	if cleanName == "." || cleanName == ".." || strings.HasPrefix(cleanName, ".."+string(os.PathSeparator)) {
		return "", merror.New("project path must stay inside the current directory")
	}
	target := filepath.Join(cwd, cleanName)
	relative, err := filepath.Rel(cwd, target)
	if err != nil || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		return "", merror.New("project path must stay inside the current directory")
	}
	return target, nil
}
