package cli

import "github.com/fatih/color"

// PrintSuccess prints a message to the console in green.
func PrintSuccess(format string, a ...interface{}) {
	color.Green(format, a...)
}

// PrintError prints a formatted error message to the console in red.
func PrintError(format string, a ...interface{}) {
	color.Red("Error: "+format, a...)
}

// PrintWarn prints a formatted warning message to the console in yellow.
func PrintWarn(format string, a ...interface{}) {
	color.Yellow("Warning: "+format, a...)
}

// PrintInfo prints a standard message to the console.
func PrintInfo(format string, a ...interface{}) {
	color.White(format, a...)
}
