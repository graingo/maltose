// Package cli provides the command-line interface for the application.
package cli

import (
	"fmt"

	"github.com/graingo/maltose"
	"github.com/spf13/cobra"
)

var (
	versionFlag bool
)

var rootCmd = &cobra.Command{
	Use:   "maltose",
	Short: "Maltose is a lightweight and powerful Go framework for building modern web applications.",
	Long: `Maltose provides an elegant and concise way to build web services, 
with a focus on high performance, scalability, and developer experience.
It includes features like routing, middleware, configuration management, and more.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if versionFlag {
			fmt.Printf("Maltose CLI version: %s\n", maltose.VERSION)
			return nil
		}
		// Default action when no subcommand is provided
		return cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Version = maltose.VERSION
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.maltose.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
