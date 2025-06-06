package cli

import (
	"fmt"
	"os"

	"github.com/Xuanwo/go-locale"
	"github.com/graingo/maltose"
	"github.com/graingo/maltose/cmd/maltose/i18n"
	"github.com/spf13/cobra"
)

var (
	versionFlag bool
)

var rootCmd = &cobra.Command{
	Use:   "maltose",
	Short: i18n.T("root_cmd_short", nil),
	Long:  i18n.T("root_cmd_long", nil),
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
	rootCmd.PersistentFlags().BoolVarP(&versionFlag, "version", "v", false, "Print the version number of Maltose")
	rootCmd.PersistentFlags().StringVar(&i18n.Lang, "lang", getSystemLang(), "Language for CLI output (e.g., 'en', 'zh'). Defaults to system language.")
}
