package cli

import (
	"fmt"
	"os"

	"github.com/Xuanwo/go-locale"
	"github.com/graingo/maltose"
	"github.com/graingo/maltose/cmd/maltose/i18n"
	"github.com/spf13/cobra"
)

var versionFlag bool

var rootCmd = &cobra.Command{
	Use:   "maltose",
	Short: "Maltose CLI application",
	Long: `Maltose CLI is a powerful tool for the Maltose framework.

It provides a collection of commands to boost your development efficiency,
including creating new projects, generating code (models, DAO), 
and generating OpenAPI documentation.`,
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
func Execute() {
	// Initialize i18n after flags are parsed but before command execution
	cobra.OnInitialize(i18n.Init)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

// getSystemLang detects the OS language and returns 'zh' for Chinese, otherwise 'en'.
func getSystemLang() string {
	lang, err := locale.Detect()
	if err != nil {
		return "en" // Default to English on error
	}
	// We only care about the primary language tag (e.g., "zh" from "zh-CN").
	base, _ := lang.Base()
	if base.String() == "zh" {
		return "zh"
	}
	return "en"
}

func init() {
	// Here you will define your flags and configuration settings.
	// cobra.OnInitialize(initConfig) // Example: if you have a config file

	rootCmd.PersistentFlags().BoolVarP(&versionFlag, "version", "v", false, "Print the version number of Maltose")
	rootCmd.PersistentFlags().StringVar(&i18n.Lang, "lang", getSystemLang(), "Language for CLI output (e.g., 'en', 'zh'). Defaults to system language.")
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.maltose.yaml)")
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// AddCommand allows adding subcommands from other files.
func AddCommand(cmd *cobra.Command) {
	rootCmd.AddCommand(cmd)
}
