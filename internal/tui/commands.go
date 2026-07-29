package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/ericklucioh/mobdesk/internal/status"
	"github.com/ericklucioh/mobdesk/internal/version"
)

func runCommand(ctx context.Context, args ...string) tea.Cmd {
	return func() tea.Msg {
		if len(args) == 0 {
			return operationMessage{err: fmt.Errorf("operação sem comando")}
		}
		binary, err := os.Executable()
		if err != nil {
			return operationMessage{command: args[0], err: err}
		}
		cmd := exec.CommandContext(ctx, binary, args...)
		output, err := cmd.Output()
		return operationFromOutput(args[0], output, err)
	}
}

func operationFromOutput(command string, output []byte, commandErr error) operationMessage {
	result, parseErr := decodeOperation(output)
	if parseErr == nil {
		// JSON estruturado é a resposta final da CLI, mesmo quando ela retorna
		// status não zero para sinalizar falha à automação.
		return operationMessage{command: command, result: result}
	}
	if commandErr != nil {
		return operationMessage{command: command, err: commandErr}
	}
	return operationMessage{command: command, err: fmt.Errorf("resposta JSON inválida: %w", parseErr)}
}

func runStatusCommand(parent context.Context) tea.Cmd {
	return func() tea.Msg {
		binary, err := os.Executable()
		if err != nil {
			return statusMessage{err: err}
		}
		ctx, cancel := context.WithTimeout(parent, 15*time.Second)
		defer cancel()
		output, err := exec.CommandContext(ctx, binary, "status", "--json").Output()
		if ctx.Err() != nil {
			return statusMessage{err: fmt.Errorf("coleta de status excedeu 15 segundos: %w", ctx.Err())}
		}
		value := status.SystemStatus{}
		if parseErr := json.Unmarshal(output, &value); parseErr != nil {
			if err != nil {
				return statusMessage{err: fmt.Errorf("executar mobdesk status: %w", err)}
			}
			return statusMessage{err: fmt.Errorf("resposta JSON inválida do status: %w", parseErr)}
		}
		versionOutput, err := exec.CommandContext(ctx, binary, "version", "--json").Output()
		if ctx.Err() != nil {
			return statusMessage{err: fmt.Errorf("coleta da versão excedeu 15 segundos: %w", ctx.Err())}
		}
		if err != nil {
			return statusMessage{err: fmt.Errorf("executar mobdesk version: %w", err)}
		}
		info := version.Info{}
		if parseErr := json.Unmarshal(versionOutput, &info); parseErr != nil {
			return statusMessage{err: fmt.Errorf("resposta JSON inválida da versão: %w", parseErr)}
		}
		return statusMessage{value: value, info: info}
	}
}

func realShellCommand(ctx context.Context) tea.Cmd {
	binary, err := os.Executable()
	if err != nil {
		return func() tea.Msg { return operationMessage{command: "shell", err: err} }
	}
	return tea.ExecProcess(exec.CommandContext(ctx, binary, "shell"), func(err error) tea.Msg {
		return operationMessage{command: "shell", err: err}
	})
}
