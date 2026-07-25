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

func runCommand(args ...string) tea.Cmd {
	return func() tea.Msg {
		binary, err := os.Executable()
		if err != nil {
			return operationMessage{command: args[0], err: err}
		}
		cmd := exec.Command(binary, args...)
		output, err := cmd.Output()
		result, parseErr := decodeOperation(output)
		if parseErr != nil && err == nil {
			return operationMessage{command: args[0], err: fmt.Errorf("resposta JSON inválida: %w", parseErr)}
		}
		if parseErr != nil && err != nil {
			// O comando pode ter falhado antes de produzir JSON; preserva o erro
			// de processo, que é mais útil para a tela de resultado.
			_ = json.Valid(output)
		}
		return operationMessage{command: args[0], result: result, err: err}
	}
}

func runStatusCommand() tea.Cmd {
	return func() tea.Msg {
		binary, err := os.Executable()
		if err != nil {
			return statusMessage{err: err}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
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
		return statusMessage{value: value, info: version.Current()}
	}
}

func realShellCommand() tea.Cmd {
	binary, err := os.Executable()
	if err != nil {
		return func() tea.Msg { return operationMessage{command: "shell", err: err} }
	}
	return tea.ExecProcess(exec.Command(binary, "shell"), func(err error) tea.Msg {
		return operationMessage{command: "shell", err: err}
	})
}
