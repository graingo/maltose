package i18n

import (
	"embed"
	"encoding/json"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed en.json zh.json
var fs embed.FS

var bundle *i18n.Bundle
var Localizer *i18n.Localizer
var Lang string

func Init() {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// Load files from the embedded filesystem
	bundle.LoadMessageFileFS(fs, "en.json")
	bundle.LoadMessageFileFS(fs, "zh.json")

	// Lang is set from a command-line flag in cli/root.go
	Localizer = i18n.NewLocalizer(bundle, Lang)
}

func T(messageID string, templateData map[string]interface{}, pluralCount ...interface{}) string {
	config := &i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: templateData,
	}
	if len(pluralCount) > 0 {
		config.PluralCount = pluralCount[0]
	}

	localized, err := Localizer.Localize(config)
	if err != nil {
		// Fallback to messageID if translation not found
		return messageID
	}
	return localized
}
