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

// PrintNotice prints a formatted, localized notice message to the console in cyan.
func PrintNotice(messageID string, templateData TplData) {
	message := i18n.T(messageID, templateData)
	color.Cyan(message)
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

// Print returns a localized message.
func Print(messageID string) string {
	return i18n.T(messageID, nil)
}

// Printf returns a formatted, localized message.
func Printf(messageID string, templateData TplData) string {
	return i18n.T(messageID, templateData)
}
