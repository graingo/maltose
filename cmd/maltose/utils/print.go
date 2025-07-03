package utils

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

type TplData map[string]any

func formatMessage(messageID string, templateData TplData) string {
	if templateData == nil {
		return messageID
	}
	msg := messageID
	for key, value := range templateData {
		placeholder := fmt.Sprintf("{{.%s}}", key)
		msg = strings.ReplaceAll(msg, placeholder, fmt.Sprintf("%v", value))
	}
	return msg
}

// PrintSuccess prints a localized message to the console in green.
func PrintSuccess(messageID string, templateData TplData) {
	message := formatMessage(messageID, templateData)
	color.Green(message)
}

// PrintError prints a formatted, localized error message to the console in red.
func PrintError(messageID string, templateData TplData) {
	message := formatMessage(messageID, templateData)
	color.Red("Error: " + message)
}

// PrintNotice prints a formatted, localized notice message to the console in cyan.
func PrintNotice(messageID string, templateData TplData) {
	message := formatMessage(messageID, templateData)
	color.Cyan(message)
}

// PrintWarn prints a formatted, localized warning message to the console in yellow.
func PrintWarn(messageID string, templateData TplData) {
	message := formatMessage(messageID, templateData)
	color.Yellow("Warning: " + message)
}

// PrintInfo prints a standard, localized message to the console.
func PrintInfo(messageID string, templateData TplData) {
	message := formatMessage(messageID, templateData)
	color.White(message)
}

// Print returns a localized message.
func Print(messageID string) string {
	return messageID
}

// Printf returns a formatted, localized message.
func Printf(messageID string, templateData TplData) string {
	return formatMessage(messageID, templateData)
}
