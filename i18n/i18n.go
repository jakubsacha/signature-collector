package i18n

import (
	"embed"
	"encoding/json"
	"log"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed locales/*.json
var localeFS embed.FS

var bundle *i18n.Bundle
var localizer *i18n.Localizer
var currentLang string

func Init(lang string) error {
	currentLang = lang
	// Create bundle with the requested language as default
	bundle = i18n.NewBundle(language.MustParse(lang))
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	// Load all locale files
	entries, err := localeFS.ReadDir("locales")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			_, err = bundle.LoadMessageFileFS(localeFS, "locales/"+entry.Name())
			if err != nil {
				return err
			}
		}
	}

	// Set the localizer with fallback languages
	localizer = i18n.NewLocalizer(bundle, lang, "en")
	log.Printf("Initialized i18n with language: %s", lang)
	return nil
}

func GetLanguage() string {
	return currentLang
}

func T(messageID string, templateData map[string]interface{}) string {
	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		MessageID:    messageID,
		TemplateData: templateData,
	})
	if err != nil {
		log.Printf("Localization error for message ID '%s': %v", messageID, err)
		return messageID
	}
	return msg
}
