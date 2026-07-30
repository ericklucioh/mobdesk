package workstation

import (
	"context"
	"fmt"
	"os"
	"os/user"

	"github.com/ericklucioh/mobdesk/internal/executil"
)

func (s Service) run(ctx context.Context, name string, args ...string) error {
	if err := s.Deps.Run(ctx, name, args...); err != nil {
		return fmt.Errorf("command %q failed: %w", name, err)
	}
	return nil
}

func runCommand(ctx context.Context, name string, args ...string) error {
	command, err := executil.CommandContext(ctx, name, args...)
	if err != nil {
		return err
	}
	command.Stdin, command.Stdout, command.Stderr = os.Stdin, os.Stdout, os.Stderr
	return command.Run()
}

func startSSHD(ctx context.Context, config, log string) error {
	command, err := executil.CommandContext(ctx, "sshd", "-f", config, "-E", log)
	if err != nil {
		return fmt.Errorf("resolve sshd: %w", err)
	}
	command.Stdin, command.Stdout, command.Stderr = os.Stdin, os.Stdout, os.Stderr
	return command.Run()
}

func wakeLock() error {
	command, err := executil.Command("termux-wake-lock")
	if err != nil {
		return fmt.Errorf("termux-wake-lock is unavailable in this environment")
	}
	return command.Run()
}

func wakeUnlock() error {
	command, err := executil.Command("termux-wake-unlock")
	if err != nil {
		return nil
	}
	return command.Run()
}

func currentUsername() string {
	if name := os.Getenv("USER"); name != "" {
		return name
	}
	if current, err := user.Current(); err == nil {
		return current.Username
	}
	return ""
}
