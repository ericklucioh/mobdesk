package i18n

import (
	"errors"
	"os"
	"strings"
)

// Locale identifies one of the locales supported by Mobdesk.
type Locale string

const (
	LocaleENUS Locale = "en-US"
	LocalePTBR Locale = "pt-BR"
)

var environmentKeys = []string{
	"MOBDESK_LOCALE",
	"LC_ALL",
	"LC_MESSAGES",
	"LANG",
}

// UnsupportedLocaleError identifies an explicit locale value that Mobdesk
// cannot resolve. Presentation layers localize it with ErrorInvalidLocale.
type UnsupportedLocaleError struct {
	Value string
}

func (e *UnsupportedLocaleError) Error() string { return string(ErrorInvalidLocale) }

func IsUnsupportedLocale(err error) bool {
	var target *UnsupportedLocaleError
	return errors.As(err, &target)
}

// Resolve selects a locale using the explicit value and POSIX environment
// precedence defined by the localization contract.
func Resolve(explicit string, environ func(string) string) (Locale, error) {
	if environ == nil {
		environ = os.Getenv
	}
	if strings.TrimSpace(explicit) != "" {
		locale, ok := parseLocale(explicit)
		if !ok {
			return "", &UnsupportedLocaleError{Value: explicit}
		}
		return locale, nil
	}
	for _, key := range environmentKeys {
		if value := strings.TrimSpace(environ(key)); value != "" {
			if locale, ok := parseLocale(value); ok {
				return locale, nil
			}
			if key == "MOBDESK_LOCALE" {
				continue
			}
		}
	}
	return LocaleENUS, nil
}

func parseLocale(value string) (Locale, bool) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	normalized = strings.Split(normalized, ".")[0]
	normalized = strings.Split(normalized, "@")[0]
	normalized = strings.ReplaceAll(normalized, "_", "-")
	switch normalized {
	case "c", "posix", "en", "en-us":
		return LocaleENUS, true
	case "pt", "pt-br":
		return LocalePTBR, true
	default:
		return "", false
	}
}
