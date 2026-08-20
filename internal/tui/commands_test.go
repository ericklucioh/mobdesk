package tui

import (
	"errors"
	"strings"
	"testing"
)

func TestOperationFromOutputPreservesStructuredFailure(t *testing.T) {
	message := operationFromOutput("setup", []byte(`{"success":false,"state":"failed","message":"setup pendente"}`), errors.New("exit status 1"))

	if message.err != nil || message.result.Success || message.result.Message != "setup pendente" {
		t.Fatalf("structured failure was lost: %+v", message)
	}
}

func TestOperationFromOutputReturnsProcessErrorWithoutJSON(t *testing.T) {
	commandErr := errors.New("exit status 1")
	message := operationFromOutput("install", nil, commandErr)

	if !errors.Is(message.err, commandErr) {
		t.Fatalf("process error was not preserved: %+v", message)
	}
}

func TestOperationMessageTextUsesPrimaryVersion(t *testing.T) {
	text := operationMessageText(operationMessage{
		command: "install",
		result:  operationResult{Language: "node", Version: "node: v24.0.0\nnpm: 11.0.0"},
	})
	if strings.Contains(text, "\n") || !strings.Contains(text, "node: v24.0.0") || strings.Contains(text, "npm:") {
		t.Fatalf("unexpected install message: %q", text)
	}
}
