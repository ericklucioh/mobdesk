package tui

import (
	"errors"
	"strings"
	"testing"

	"github.com/ericklucioh/mobdesk/internal/status"
	"github.com/ericklucioh/mobdesk/internal/version"
)

func TestOperationFromOutputPreservesStructuredFailure(t *testing.T) {
	message := operationFromOutput("setup", []byte(`{"schema_version":1,"command":"setup","success":false,"state":"failed","message":"setup pendente"}`), errors.New("exit status 1"))

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

func TestOperationFromOutputRejectsIncompleteContract(t *testing.T) {
	message := operationFromOutput("install", []byte(`{"success":true,"state":"installed"}`), nil)
	if message.err == nil || !strings.Contains(message.err.Error(), "invalid JSON response") {
		t.Fatalf("incomplete contract was accepted: %+v", message)
	}
}

func TestResponseValidatorsRejectIncompatibleSchemas(t *testing.T) {
	if validStatusResponse(status.SystemStatus{SchemaVersion: 1, Command: "status", State: "healthy"}) {
		t.Fatal("status v1 was accepted")
	}
	if validStatusResponse(status.SystemStatus{SchemaVersion: status.SchemaVersion, Command: "other", State: "healthy"}) {
		t.Fatal("status command mismatch was accepted")
	}
	if validVersionResponse(version.Info{SchemaVersion: version.SchemaVersion, Command: "version"}) {
		t.Fatal("version without state was accepted")
	}
}

func TestHasJSONFieldsRejectsMissingEnvelopeMembers(t *testing.T) {
	if hasJSONFields([]byte(`{"schema_version":1,"command":"status","success":true,"state":"healthy"}`), "schema_version", "command", "success", "state", "message") {
		t.Fatal("incomplete envelope was accepted")
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
