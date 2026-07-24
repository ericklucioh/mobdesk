package executil

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Resolve avoids os/exec.LookPath on Android. Go uses faccessat2 while
// resolving a bare command name, but Termux's seccomp policy can reject that
// syscall even when the executable itself is available.
func Resolve(name string) (string, error) {
	if strings.ContainsRune(name, filepath.Separator) || runtime.GOOS != "android" {
		if strings.ContainsRune(name, filepath.Separator) {
			return name, nil
		}
		return exec.LookPath(name)
	}

	prefix := os.Getenv("PREFIX")
	if prefix == "" {
		prefix = "/data/data/com.termux/files/usr"
	}
	return termuxPath(name, prefix)
}

func termuxPath(name, prefix string) (string, error) {
	path := filepath.Join(prefix, "bin", name)
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("resolver comando %q em %s: %w", name, path, err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("comando %q aponta para um diretório: %s", name, path)
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
