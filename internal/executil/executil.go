package executil

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ericklucioh/mobdesk/internal/i18n"
)

// Resolve avoids os/exec.LookPath on Android. Go uses faccessat2 while
// resolving a bare command name, but Termux's seccomp policy can reject that
// syscall even when the executable itself is available.
func Resolve(name string) (string, error) {
	if strings.ContainsRune(name, filepath.Separator) {
		return name, nil
	}

	prefix := os.Getenv("PREFIX")
	// Release binaries are built with GOOS=linux for the Termux ARM64
	// userspace, so runtime.GOOS is linux even while running on Android.
	// PREFIX is the reliable Termux marker in that situation.
	if prefix == "" && runtime.GOOS != "android" {
		return exec.LookPath(name)
	}
	if prefix == "" {
		prefix = "/data/data/com.termux/files/usr"
	}
	return termuxPath(name, prefix)
}

func termuxPath(name, prefix string) (string, error) {
	path := filepath.Join(prefix, "bin", name)
	info, err := os.Stat(path)
	if err != nil {
		return "", i18n.NewError(i18n.ServiceExecError, "exec_resolve", map[string]any{"Name": name, "Detail": path}, err)
	}
	if info.IsDir() {
		return "", i18n.NewError(i18n.ServiceExecError, "exec_directory", map[string]any{"Name": name, "Detail": path}, nil)
	}
	return path, nil
}

func CommandContext(ctx context.Context, name string, args ...string) (*exec.Cmd, error) {
	path, err := Resolve(name)
	if err != nil {
		return nil, err
	}
	return exec.CommandContext(ctx, path, args...), nil
}

func Command(name string, args ...string) (*exec.Cmd, error) {
	path, err := Resolve(name)
	if err != nil {
		return nil, err
	}
	return exec.Command(path, args...), nil
}
