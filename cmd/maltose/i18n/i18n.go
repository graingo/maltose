package i18n

import (
	"embed"
	"encoding/json"

	"github.com/Xuanwo/go-locale"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed en.json zh.json
var fs embed.FS

var bundle *i18n.Bundle
var Localizer *i18n.Localizer

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
	// lang := getSystemLang()

	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// Load files from the embedded filesystem
	bundle.LoadMessageFileFS(fs, "en.json")
	bundle.LoadMessageFileFS(fs, "zh.json")

	Localizer = i18n.NewLocalizer(bundle, "en")
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
