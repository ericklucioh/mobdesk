package i18n

import (
	"errors"
	"testing"
)

func TestResolveUsesExplicitLocaleFirst(t *testing.T) {
	locale, err := Resolve("pt_BR.UTF-8", func(key string) string {
		if key == "MOBDESK_LOCALE" {
			return "en-US"
		}
		return ""
	})
	if err != nil || locale != LocalePTBR {
		t.Fatalf("Resolve() = %q, %v; want pt-BR", locale, err)
	}
}

func TestResolveUsesEnvironmentPrecedence(t *testing.T) {
	values := map[string]string{
		"MOBDESK_LOCALE": "unsupported",
		"LC_ALL":         "pt_BR.UTF-8",
		"LC_MESSAGES":    "en-US",
	}
	locale, err := Resolve("", func(key string) string { return values[key] })
	if err != nil || locale != LocalePTBR {
		t.Fatalf("Resolve() = %q, %v; want pt-BR", locale, err)
	}
}

func TestResolveFallsBackToEnglish(t *testing.T) {
	locale, err := Resolve("", func(string) string { return "unsupported" })
	if err != nil || locale != LocaleENUS {
		t.Fatalf("Resolve() = %q, %v; want en-US", locale, err)
	}
}

func TestResolveRejectsInvalidExplicitLocale(t *testing.T) {
	_, err := Resolve("fr-FR", func(string) string { return "pt-BR" })
	if err == nil || !IsUnsupportedLocale(err) {
		t.Fatal("Resolve() accepted unsupported explicit locale")
	}
	var unsupported *UnsupportedLocaleError
	if !errors.As(err, &unsupported) || unsupported.Value != "fr-FR" {
		t.Fatalf("unexpected locale error: %v", err)
	}
}
