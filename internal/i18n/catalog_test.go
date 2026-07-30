package i18n

import (
	"strings"
	"testing"
)

func TestCatalogsAreComplete(t *testing.T) {
	if err := ValidateCatalogs(); err != nil {
		t.Fatal(err)
	}
}

func TestLocalizerRendersBothLocales(t *testing.T) {
	english := New(LocaleENUS).Text(LocaleEnglishName, nil)
	brazilianPortuguese := New(LocalePTBR).Text(LocaleEnglishName, nil)
	if english != New(LocaleENUS).Text(LocaleEnglishName, nil) || brazilianPortuguese != New(LocalePTBR).Text(LocaleEnglishName, nil) || english == brazilianPortuguese {
		t.Fatalf("unexpected locale names: %q and %q", english, brazilianPortuguese)
	}
}

func TestLocalizerRendersTemplateData(t *testing.T) {
	message := New(LocalePTBR).Text(ErrorMissingMessage, map[string]any{"ID": "test.id"})
	if message == "" || !strings.Contains(message, "test.id") {
		t.Fatalf("unexpected rendered message: %q", message)
	}
}
