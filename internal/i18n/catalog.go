package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
)

//go:embed locale/*.json
var catalogFiles embed.FS

var catalogs = loadCatalogs()

func loadCatalogs() map[Locale]map[MessageID]string {
	result := make(map[Locale]map[MessageID]string, 2)
	for _, locale := range []Locale{LocaleENUS, LocalePTBR} {
		content, err := catalogFiles.ReadFile("locale/" + string(locale) + ".json")
		if err != nil {
			panic(fmt.Sprintf("load %s catalog: %v", locale, err))
		}
		var raw map[string]string
		if err := json.Unmarshal(content, &raw); err != nil {
			panic(fmt.Sprintf("parse %s catalog: %v", locale, err))
		}
		messages := make(map[MessageID]string, len(raw))
		for id, message := range raw {
			messages[MessageID(id)] = message
		}
		result[locale] = messages
	}
	if err := validateCatalogs(result); err != nil {
		panic(err)
	}
	return result
}

// ValidateCatalogs checks that all supported locales contain the same required
// message IDs and no empty translations.
func ValidateCatalogs() error {
	return validateCatalogs(catalogs)
}

func validateCatalogs(values map[Locale]map[MessageID]string) error {
	for _, locale := range []Locale{LocaleENUS, LocalePTBR} {
		messages, ok := values[locale]
		if !ok {
			return fmt.Errorf("missing catalog for locale %s", locale)
		}
		for _, id := range requiredMessageIDs {
			if messages[id] == "" {
				return fmt.Errorf("missing translation %s for locale %s", id, locale)
			}
		}
	}
	return nil
}
