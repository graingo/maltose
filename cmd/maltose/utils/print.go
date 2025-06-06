package utils

import (
	"github.com/fatih/color"
	"github.com/graingo/maltose/cmd/maltose/i18n"
)

type TplData map[string]any

// PrintSuccess prints a localized message to the console in green.
func PrintSuccess(messageID string, templateData TplData) {
	message := i18n.T(messageID, templateData)
	color.Green(message)
}

// PrintError prints a formatted, localized error message to the console in red.
func PrintError(messageID string, templateData TplData) {
	message := i18n.T(messageID, templateData)
	color.Red("Error: " + message)
}

// PrintWarn prints a formatted, localized warning message to the console in yellow.
func PrintWarn(messageID string, templateData TplData) {
	message := i18n.T(messageID, templateData)
	color.Yellow("Warning: " + message)
}

// PrintInfo prints a standard, localized message to the console.
func PrintInfo(messageID string, templateData TplData) {
	message := i18n.T(messageID, templateData)
	color.White(message)
}
