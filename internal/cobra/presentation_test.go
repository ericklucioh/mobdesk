package cobra

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/ericklucioh/mobdesk/internal/i18n"
)

func TestRootHelpUsesSelectedLocale(t *testing.T) {
	tests := []struct {
		name   string
		locale string
		want   string
		avoid  string
	}{
		{name: "english", locale: "en-US", want: "Usage:", avoid: "Uso:"},
		{name: "brazilian portuguese", locale: "pt-BR", want: "Uso:", avoid: "Usage:"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := NewRootCmdWithEnv(func(string) string { return "en-US" })
			var output, errors bytes.Buffer
			root.SetOut(&output)
			root.SetErr(&errors)
			root.SetArgs([]string{"--help", "--locale", test.locale})
			if err := root.Execute(); err != nil {
				t.Fatal(err)
			}
			value := output.String()
			if !strings.Contains(value, test.want) || strings.Contains(value, test.avoid) {
				t.Fatalf("help = %q", value)
			}
			for _, command := range []string{"start", "stop", "setup", "status", "install", "uninstall", "config", "update", "version", "tui", "shell"} {
				if !strings.Contains(value, "mobdesk "+command) {
					t.Fatalf("help omitted command %q: %s", command, value)
				}
			}
			if !strings.Contains(value, "--locale") || !strings.Contains(value, "--json") {
				t.Fatalf("help omitted stable locale/json flags: %s", value)
			}
		})
	}
}

func TestTextOutputUsesSelectedLocale(t *testing.T) {
	data := map[string]any{"Current": "1", "Latest": "2"}
	english := localized([]i18n.Localizer{i18n.New(i18n.LocaleENUS)}, i18n.OutputUpdateAvailable, data)
	brazilianPortuguese := localized([]i18n.Localizer{i18n.New(i18n.LocalePTBR)}, i18n.OutputUpdateAvailable, data)
	if english != i18n.New(i18n.LocaleENUS).Text(i18n.OutputUpdateAvailable, data) || brazilianPortuguese != i18n.New(i18n.LocalePTBR).Text(i18n.OutputUpdateAvailable, data) || english == brazilianPortuguese {
		t.Fatalf("localized text = %q and %q", english, brazilianPortuguese)
	}
}

func TestJSONFailureKeepsStableFieldsAndLocale(t *testing.T) {
	output, err := captureStdout(func() error {
		return runStartJSON(context.Background(), i18n.New(i18n.LocalePTBR))
	})
	if err == nil {
		t.Fatal("start unexpectedly succeeded outside Termux")
	}
	var result operationResult
	if decodeErr := json.Unmarshal([]byte(output), &result); decodeErr != nil {
		t.Fatalf("invalid JSON: %q: %v", output, decodeErr)
	}
	if result.SchemaVersion != 1 || result.Command != "start" || result.Success || result.State != "failed" || result.Locale != "pt-BR" || result.MessageID == "" || result.ErrorCode == "" {
		t.Fatalf("unstable JSON result: %+v", result)
	}
}

func TestInvalidLocaleJSONIsValidAndLocalizedByFallback(t *testing.T) {
	root := NewRootCmdWithEnv(func(key string) string {
		if key == "LANG" {
			return "pt-BR"
		}
		return ""
	})
	root.SetArgs([]string{"status", "--json", "--locale", "fr-FR"})
	output, err := captureStdout(func() error { return root.Execute() })
	if err == nil || !strings.Contains(err.Error(), "locale") {
		t.Fatalf("invalid locale error = %v", err)
	}
	var result operationResult
	if decodeErr := json.Unmarshal([]byte(output), &result); decodeErr != nil {
		t.Fatalf("invalid JSON: %q: %v", output, decodeErr)
	}
	if result.Success || result.State != "failed" || result.Locale != "pt-BR" || result.ErrorCode != "invalid_locale" || result.MessageID != string(i18n.ErrorInvalidLocale) {
		t.Fatalf("invalid locale result: %+v", result)
	}
}

func TestRootCommandsDoNotShareLocaleState(t *testing.T) {
	english := NewRootCmdWithEnv(func(string) string { return "en-US" })
	brazilianPortuguese := NewRootCmdWithEnv(func(string) string { return "pt-BR" })
	var englishOutput, portugueseOutput bytes.Buffer
	english.SetOut(&englishOutput)
	english.SetArgs([]string{"--help"})
	brazilianPortuguese.SetOut(&portugueseOutput)
	brazilianPortuguese.SetArgs([]string{"--help"})
	if err := english.Execute(); err != nil {
		t.Fatal(err)
	}
	if err := brazilianPortuguese.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(englishOutput.String(), "Usage:") || !strings.Contains(portugueseOutput.String(), "Uso:") {
		t.Fatalf("locale state leaked: %q / %q", englishOutput.String(), portugueseOutput.String())
	}
}
