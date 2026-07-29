package i18n

import (
	"bytes"
	"text/template"
)

// MessageID identifies a translatable user-facing message.
type MessageID string

const (
	ErrorInvalidLocale     MessageID = "error.invalid_locale"
	LocaleEnglishName      MessageID = "locale.english_name"
	LocalePortugueseBRName MessageID = "locale.portuguese_br_name"
	ErrorMissingMessage    MessageID = "error.missing_message"
)

var requiredMessageIDs = []MessageID{
	ErrorInvalidLocale,
	LocaleEnglishName,
	LocalePortugueseBRName,
	ErrorMissingMessage,
}

// Localizer renders messages from one immutable locale catalog.
type Localizer struct {
	Locale   Locale
	messages map[MessageID]string
}

// New returns a localizer for locale. Catalogs are embedded in the binary and
// validated during package initialization.
func New(locale Locale) Localizer {
	if locale != LocalePTBR {
		locale = LocaleENUS
	}
	return Localizer{Locale: locale, messages: catalogs[locale]}
}

// Text renders a message with optional template data. Message templates use
// Go's text/template syntax so translations can reorder named values safely.
func (l Localizer) Text(id MessageID, data any) string {
	message, ok := l.messages[id]
	if !ok {
		message = catalogs[LocaleENUS][ErrorMissingMessage]
		data = map[string]any{"ID": id}
	}
	if data == nil {
		return message
	}
	templateValue, err := template.New(string(id)).Option("missingkey=error").Parse(message)
	if err != nil {
		return message
	}
	var output bytes.Buffer
	if err := templateValue.Execute(&output, data); err != nil {
		return message
	}
	return output.String()
}

// RequiredMessageIDs returns a copy of the IDs that every catalog must contain.
func RequiredMessageIDs() []MessageID {
	return append([]MessageID(nil), requiredMessageIDs...)
}
