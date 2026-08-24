// Package cli provides the command-line interface for the application.
package cli

import (
	"github.com/graingo/maltose"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "maltose",
	Short: "Maltose is a lightweight and powerful Go framework for building modern web applications.",
	Long: `Maltose provides an elegant and concise way to build web services, 
with a focus on high performance, scalability, and developer experience.
It includes features like routing, middleware, configuration management, and more.`,
	RunE: func(cmd *cobra.Command, _ []string) error { return cmd.Help() },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Version = maltose.VERSION
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}
