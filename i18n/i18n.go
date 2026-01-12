package i18n

import (
	"embed"
	"encoding/json"

	"github.com/jakubsacha/signature-collector/logging"
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
	logging.WithField("language", lang).Info("Initialized i18n")
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
		logging.WithFields(map[string]interface{}{
			"message_id": messageID,
			"error":      err.Error(),
		}).Warn("Localization error")
		return messageID
	}
	return msg
}
