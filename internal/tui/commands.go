package tui

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/ericklucioh/mobdesk/internal/i18n"
	"github.com/ericklucioh/mobdesk/internal/status"
	"github.com/ericklucioh/mobdesk/internal/version"
)

func runCommand(ctx context.Context, args ...string) tea.Cmd {
	return runCommandWithLocale(ctx, i18n.LocaleENUS, args...)
}

func runCommandWithLocale(ctx context.Context, locale i18n.Locale, args ...string) tea.Cmd {
	return func() tea.Msg {
		if len(args) == 0 {
			return operationMessage{err: fmt.Errorf("operation has no command")}
		}
		binary, err := os.Executable()
		if err != nil {
			return operationMessage{command: args[0], err: err}
		}
		cmd := exec.CommandContext(ctx, binary, appendLocale(args, locale)...)
		output, err := cmd.Output()
		return operationFromOutput(args[0], output, err)
	}
}

func runInstallCommand(ctx context.Context, args ...string) tea.Cmd {
	return runInstallCommandWithLocale(ctx, i18n.LocaleENUS, args...)
}

func runInstallCommandWithLocale(ctx context.Context, locale i18n.Locale, args ...string) tea.Cmd {
	return func() tea.Msg {
		stream := &operationStream{messages: make(chan tea.Msg, 1), command: args[0]}
		go stream.run(ctx, locale, args...)
		return stream.next()()
	}
}

type operationStream struct {
	messages chan tea.Msg
	command  string
}

func (s *operationStream) next() tea.Cmd {
	return func() tea.Msg {
		message, ok := <-s.messages
		if !ok {
			return operationMessage{command: s.command, err: fmt.Errorf("operation stream ended without a result")}
		}
		if progress, ok := message.(operationProgressMessage); ok {
			progress.next = s.next()
			return progress
		}
		return message
	}
}

func (s *operationStream) run(ctx context.Context, locale i18n.Locale, args ...string) {
	if len(args) == 0 {
		s.send(ctx, operationMessage{err: fmt.Errorf("operation has no command")})
		return
	}
	binary, err := os.Executable()
	if err != nil {
		s.send(ctx, operationMessage{command: args[0], err: err})
		return
	}
	command := exec.CommandContext(ctx, binary, appendLocale(args, locale)...)
	output, err := command.StdoutPipe()
	if err != nil {
		s.send(ctx, operationMessage{command: args[0], err: err})
		return
	}
	var stderr bytes.Buffer
	command.Stderr = &stderr
	if err := command.Start(); err != nil {
		s.send(ctx, operationMessage{command: args[0], err: err})
		return
	}

	var result []byte
	scanner := bufio.NewScanner(output)
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	for scanner.Scan() {
		line := append([]byte(nil), scanner.Bytes()...)
		var event struct {
			Event   string `json:"event"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal(line, &event); err == nil && event.Event == "progress" {
			if !s.send(ctx, operationProgressMessage{message: event.Message}) {
				return
			}
			continue
		}
		result = line
	}
	if scanErr := scanner.Err(); scanErr != nil {
		_ = command.Wait()
		s.send(ctx, operationMessage{command: args[0], err: fmt.Errorf("read installation progress: %w", scanErr)})
		return
	}
	commandErr := command.Wait()
	if len(stderr.Bytes()) > 0 && len(result) == 0 && commandErr != nil {
		commandErr = fmt.Errorf("%w: %s", commandErr, bytes.TrimSpace(stderr.Bytes()))
	}
	s.send(ctx, operationFromOutput(args[0], result, commandErr))
}

func (s *operationStream) send(ctx context.Context, message tea.Msg) bool {
	select {
	case s.messages <- message:
		return true
	case <-ctx.Done():
		return false
	}
}

func operationFromOutput(command string, output []byte, commandErr error) operationMessage {
	result, parseErr := decodeOperation(output)
	if parseErr == nil {
		// Structured JSON is the CLI's final response even when it returns a
		// non-zero status to signal failure to automation.
		return operationMessage{command: command, result: result}
	}
	if commandErr != nil {
		return operationMessage{command: command, err: commandErr}
	}
	return operationMessage{command: command, err: fmt.Errorf("invalid JSON response: %w", parseErr)}
}

func runStatusCommand(parent context.Context, locale i18n.Locale) tea.Cmd {
	return func() tea.Msg {
		binary, err := os.Executable()
		if err != nil {
			return statusMessage{err: err}
		}
		ctx, cancel := context.WithTimeout(parent, 15*time.Second)
		defer cancel()
		output, err := exec.CommandContext(ctx, binary, appendLocale([]string{"status", "--json"}, locale)...).Output()
		if ctx.Err() != nil {
			return statusMessage{err: fmt.Errorf("status collection exceeded 15 seconds: %w", ctx.Err())}
		}
		value := status.SystemStatus{}
		if parseErr := json.Unmarshal(output, &value); parseErr != nil {
			if err != nil {
				return statusMessage{err: fmt.Errorf("run mobdesk status: %w", err)}
			}
			return statusMessage{err: fmt.Errorf("invalid status JSON response: %w", parseErr)}
		}
		versionOutput, err := exec.CommandContext(ctx, binary, appendLocale([]string{"version", "--json"}, locale)...).Output()
		if ctx.Err() != nil {
			return statusMessage{err: fmt.Errorf("version collection exceeded 15 seconds: %w", ctx.Err())}
		}
		if err != nil {
			return statusMessage{err: fmt.Errorf("run mobdesk version: %w", err)}
		}
		info := version.Info{}
		if parseErr := json.Unmarshal(versionOutput, &info); parseErr != nil {
			return statusMessage{err: fmt.Errorf("invalid version JSON response: %w", parseErr)}
		}
		return statusMessage{value: value, info: info}
	}
}

func appendLocale(args []string, locale i18n.Locale) []string {
	result := append([]string(nil), args...)
	if locale == "" {
		locale = i18n.LocaleENUS
	}
	for _, arg := range result {
		if arg == "--locale" {
			return result
		}
	}
	return append(result, "--locale", string(locale))
}

func realShellCommand(ctx context.Context, locale i18n.Locale) tea.Cmd {
	binary, err := os.Executable()
	if err != nil {
		return func() tea.Msg { return operationMessage{command: "shell", err: err} }
	}
	return tea.ExecProcess(exec.CommandContext(ctx, binary, appendLocale([]string{"shell"}, locale)...), func(err error) tea.Msg {
		return operationMessage{command: "shell", err: err}
	})
}
