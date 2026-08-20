package cobra

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/creack/pty/v2"
	"github.com/ericklucioh/mobdesk/internal/executil"
	"github.com/ericklucioh/mobdesk/internal/i18n"
	"github.com/ericklucioh/mobdesk/internal/paths"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var shellCmd = newShellCmd(nil)

func newShellCmd(state *commandState) *cobra.Command {
	return &cobra.Command{Use: "shell", RunE: func(cmd *cobra.Command, _ []string) error {
		return runShell(cmd.Context(), paths.Current(), commandLocalizer(state, cmd))
	}}
}

func runShell(ctx context.Context, p paths.Paths, localizers ...i18n.Localizer) error {
	if err := ensureSetupCompleted(p, localizers...); err != nil {
		return err
	}
	if info, err := os.Stat(p.Workspace()); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("workspace is missing; run mobdesk setup")
		}
		return fmt.Errorf("check workspace: %w", err)
	} else if !info.IsDir() {
		return fmt.Errorf("workspace is not a directory: %s", p.Workspace())
	}
	return runInteractiveInDir(ctx, p.Workspace(), "bash", "-i")
}

func ensureSetupCompleted(p paths.Paths, localizers ...i18n.Localizer) error {
	path := p.SetupDone()
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s", localized(localizers, i18n.ErrorSetupIncomplete, nil))
		}
		return fmt.Errorf("check setup state: %w", err)
	}
	return nil
}

func runInteractive(ctx context.Context, name string, args ...string) error {
	return runInteractiveInDir(ctx, "", name, args...)
}

func runInteractiveInDir(ctx context.Context, dir, name string, args ...string) error {
	fmt.Printf("\n$ %s %s\n", name, strings.Join(args, " "))
	command, err := executil.CommandContext(ctx, name, args...)
	if err != nil {
		return err
	}
	command.Dir = dir
	ptmx, err := pty.Start(command)
	if err != nil {
		return fmt.Errorf("start PTY for %q: %w", name, err)
	}
	defer func() { _ = ptmx.Close() }()
	_ = pty.InheritSize(os.Stdin, ptmx)

	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		state, rawErr := term.MakeRaw(fd)
		if rawErr == nil {
			defer func() { _ = term.Restore(fd, state) }()
		}
	}

	inputDone := make(chan struct{})
	go func() {
		_, _ = io.Copy(ptmx, os.Stdin)
		close(inputDone)
	}()
	_, copyErr := io.Copy(os.Stdout, ptmx)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if waitErr := command.Wait(); waitErr != nil {
		return fmt.Errorf("interactive command %q failed: %w", name, waitErr)
	}
	if copyErr != nil {
		return fmt.Errorf("read interactive output %q: %w", name, copyErr)
	}
	select {
	case <-inputDone:
	default:
	}
	return nil
}
