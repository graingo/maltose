package cli

import (
	"context"
	"fmt"
	"go/parser"
	"go/token"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/graingo/maltose/cmd/maltose/utils"
	"github.com/graingo/maltose/errors/merror"
	"github.com/spf13/cobra"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
)

var moduleFlag string
var repoURLFlag string

const defaultQuickstartRepository = "https://github.com/graingo/maltose-quickstart.git"

// newCmd creates a project from the Maltose quickstart repository.
var newCmd = &cobra.Command{
	Use:   "new [project-name]",
	Short: "Create a new Maltose project.",
	Long:  "Creates a new Maltose project from the quickstart repository, rewrites its module imports, and prepares its dependencies.",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]
		repoURL := defaultQuickstartRepository
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

		utils.PrintInfo("🚀 Creating new Maltose project at './{{.ProjectName}}'...", utils.TplData{"ProjectName": projectName})
		utils.PrintInfo("📥 Cloning project template from {{.RepoURL}}...", utils.TplData{"RepoURL": repoURL})
		if err := createProject(cmd.Context(), cmd.OutOrStdout(), cmd.ErrOrStderr(), cwd, projectName, modulePath, repoURL); err != nil {
			return err
		}

		utils.PrintSuccess("✅ Successfully created project '{{.ProjectName}}'.", utils.TplData{"ProjectName": projectName})
		utils.PrintInfo("🔧 Module path has been set to '{{.ModulePath}}'.", utils.TplData{"ModulePath": modulePath})
		fmt.Fprintln(cmd.OutOrStdout(), utils.Print("\n👉 To get started, run:\n"))
		fmt.Fprintln(cmd.OutOrStdout(), utils.Printf("  cd {{.ProjectName}}", utils.TplData{"ProjectName": projectName}))
		fmt.Fprintln(cmd.OutOrStdout(), utils.Print("  go run main.go"))

		return nil
	},
}

// createProject clones, customizes, and prepares a new Maltose project.
func createProject(ctx context.Context, stdout, stderr io.Writer, cwd, projectName, modulePath, repoURL string) (err error) {
	target, err := projectTarget(cwd, projectName)
	if err != nil {
		return err
	}

	cloneCmd := exec.CommandContext(ctx, "git", "clone", "--", repoURL, target)
	cloneCmd.Stdout = stdout
	cloneCmd.Stderr = stderr
	if err := cloneCmd.Run(); err != nil {
		return merror.Wrap(err, "cloning template repository failed")
	}

	completed := false
	defer func() {
		if !completed {
			_ = os.RemoveAll(target)
		}
	}()

	if err := os.RemoveAll(filepath.Join(target, ".git")); err != nil {
		return merror.Wrap(err, "failed to remove template Git metadata")
	}
	if err := rewriteProjectModule(target, modulePath); err != nil {
		return err
	}

	tidyCmd := exec.CommandContext(ctx, "go", "mod", "tidy")
	tidyCmd.Dir = target
	tidyCmd.Stdout = stdout
	tidyCmd.Stderr = stderr
	if err := tidyCmd.Run(); err != nil {
		return merror.Wrap(err, "failed to prepare project dependencies")
	}

	completed = true
	return nil
}

// rewriteProjectModule updates go.mod and imports that reference the template module.
func rewriteProjectModule(projectRoot, modulePath string) error {
	gomodPath := filepath.Join(projectRoot, "go.mod")
	content, err := os.ReadFile(gomodPath)
	if err != nil {
		return merror.Wrap(err, "failed to read go.mod")
	}

	goMod, err := modfile.Parse(gomodPath, content, nil)
	if err != nil {
		return merror.Wrap(err, "failed to parse go.mod")
	}
	if goMod.Module == nil || goMod.Module.Mod.Path == "" {
		return merror.New("template go.mod does not declare a module path")
	}

	templateModule := goMod.Module.Mod.Path
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

	return filepath.Walk(projectRoot, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}
		return rewriteGoImports(path, templateModule, modulePath)
	})
}

// rewriteGoImports replaces imports rooted at oldModule without changing other strings.
func rewriteGoImports(path, oldModule, newModule string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return merror.Wrapf(err, "failed to read Go source %s", path)
	}

	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, path, content, parser.ImportsOnly)
	if err != nil {
		return merror.Wrapf(err, "failed to parse Go imports in %s", path)
	}

	type replacement struct {
		start int
		end   int
		value string
	}
	replacements := make([]replacement, 0)
	for _, importSpec := range file.Imports {
		importPath, unquoteErr := strconv.Unquote(importSpec.Path.Value)
		if unquoteErr != nil {
			return merror.Wrapf(unquoteErr, "failed to parse import path in %s", path)
		}
		if importPath != oldModule && !strings.HasPrefix(importPath, oldModule+"/") {
			continue
		}

		start := fileSet.Position(importSpec.Path.Pos()).Offset
		end := fileSet.Position(importSpec.Path.End()).Offset
		replacements = append(replacements, replacement{
			start: start,
			end:   end,
			value: strconv.Quote(newModule + strings.TrimPrefix(importPath, oldModule)),
		})
	}

	for i := len(replacements) - 1; i >= 0; i-- {
		replacement := replacements[i]
		content = append(content[:replacement.start], append([]byte(replacement.value), content[replacement.end:]...)...)
	}
	if len(replacements) == 0 {
		return nil
	}
	if err := os.WriteFile(path, content, infoMode(path)); err != nil {
		return merror.Wrapf(err, "failed to rewrite imports in %s", path)
	}
	return nil
}

func infoMode(path string) os.FileMode {
	info, err := os.Stat(path)
	if err != nil {
		return 0644
	}
	return info.Mode().Perm()
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
