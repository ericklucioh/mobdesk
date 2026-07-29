package tui

import (
	"testing"

	"github.com/ericklucioh/mobdesk/internal/i18n"
)

func TestAppendLocaleForwardsSelectedLocale(t *testing.T) {
	args := appendLocale([]string{"status", "--json"}, i18n.LocalePTBR)
	if len(args) != 4 || args[2] != "--locale" || args[3] != "pt-BR" {
		t.Fatalf("args = %v", args)
	}
}

func TestAppendLocaleDoesNotDuplicateExplicitLocale(t *testing.T) {
	args := appendLocale([]string{"version", "--locale", "en-US"}, i18n.LocalePTBR)
	if len(args) != 3 || args[2] != "en-US" {
		t.Fatalf("args = %v", args)
	}
}
