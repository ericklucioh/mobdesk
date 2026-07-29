package i18n

import "testing"

func TestCatalogsAreComplete(t *testing.T) {
	if err := ValidateCatalogs(); err != nil {
		t.Fatal(err)
	}
}

func TestLocalizerRendersBothLocales(t *testing.T) {
	english := New(LocaleENUS).Text(LocaleEnglishName, nil)
	portuguese := New(LocalePTBR).Text(LocaleEnglishName, nil)
	if english != "English" || portuguese != "Inglês" {
		t.Fatalf("unexpected locale names: %q and %q", english, portuguese)
	}
}

func TestLocalizerRendersTemplateData(t *testing.T) {
	message := New(LocalePTBR).Text(ErrorMissingMessage, map[string]any{"ID": "test.id"})
	if message != "Tradução ausente: test.id" {
		t.Fatalf("unexpected rendered message: %q", message)
	}
}
