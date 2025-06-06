package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "maltose",
	Short: "Maltose CLI application",
	Long: `Maltose CLI is a powerful tool for the Maltose framework.

It provides a collection of commands to boost your development efficiency,
including creating new projects, generating code (models, DAO), 
and generating OpenAPI documentation.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Default action when no subcommand is provided
		fmt.Println("Maltose CLI. Use 'maltose help' for more information.")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// cobra.OnInitialize(initConfig) // Example: if you have a config file

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.maltose.yaml)")
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
// func initConfig() {
// 	if cfgFile != "" {
// 		// Use config file from the flag.
// 		viper.SetConfigFile(cfgFile)
// 	} else {
// 		// Find home directory.
// 		home, err := os.UserHomeDir()
// 		cobra.CheckErr(err)

// 		// Search config in home directory with name ".maltose" (without extension).
// 		viper.AddConfigPath(home)
// 		viper.SetConfigType("yaml")
// 		viper.SetConfigName(".maltose")
// 	}

// 	viper.AutomaticEnv() // read in environment variables that match

// 	// If a config file is found, read it in.
// 	if err := viper.ReadInConfig(); err == nil {
// 		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
// 	}
// }

// AddCommand allows adding subcommands from other files.
func AddCommand(cmd *cobra.Command) {
	rootCmd.AddCommand(cmd)
}
