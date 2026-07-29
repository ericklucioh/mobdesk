// Package workstation manages the SSH-backed development workstation.
package workstation

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/ericklucioh/mobdesk/internal/i18n"
	"github.com/ericklucioh/mobdesk/internal/paths"
)

const SSHPort = 8022

type Process interface {
	Signal(os.Signal) error
}

type Dependencies struct {
	Stat                func(string) (os.FileInfo, error)
	ReadFile            func(string) ([]byte, error)
	Remove              func(string) error
	Run                 func(context.Context, string, ...string) error
	StartSSHD           func(context.Context, string, string) error
	WakeLock            func() error
	WakeUnlock          func() error
	PortOpen            func(context.Context, int) bool
	SSHResponds         func(context.Context, int) bool
	WaitForPortClosed   func(context.Context, int, time.Duration) bool
	ProcessIsMobdeskSSH func(int, string) bool
	FindProcess         func(int) (Process, error)
	AcquireLock         func(string) (func(), error)
	EnsureSSHConfigured func(paths.Paths) error
	EnsureIfconfig      func(context.Context, io.Writer, func(context.Context, string, ...string) error) error
	Addresses           func() []string
	Username            func() string
	MkdirAll            func(string, os.FileMode) error
	Chmod               func(string, os.FileMode) error
	WriteFile           func(string, []byte, os.FileMode) error
	Lstat               func(string) (os.FileInfo, error)
	Readlink            func(string) (string, error)
	Symlink             func(string, string) error
	Executable          func() (string, error)
	Abs                 func(string) (string, error)
	EvalSymlinks        func(string) (string, error)
}

type Service struct {
	Paths paths.Paths
	Deps  Dependencies
}

type StartInfo struct {
	AlreadyRunning bool
	Addresses      []string
	Username       string
	Warnings       []error
}

type StopInfo struct {
	AlreadyStopped bool
	StaleState     bool
}

func New(p paths.Paths) Service {
	return Service{Paths: p, Deps: defaultDependencies()}
}

func workstationError(detail string, cause error) error {
	data := map[string]any{"Detail": detail}
	return i18n.NewError(i18n.ServiceWorkstationError, "workstation_operation_failed", data, cause)
}

func workstationWarning(detail string, cause error) error {
	return i18n.NewError(i18n.ServiceWorkstationWarning, "workstation_warning", map[string]any{"Detail": detail}, cause)
}

func defaultDependencies() Dependencies {
	return Dependencies{
		Stat: os.Stat, ReadFile: os.ReadFile, Remove: os.Remove, Run: runCommand, StartSSHD: startSSHD,
		WakeLock: wakeLock, WakeUnlock: wakeUnlock, PortOpen: portOpen, SSHResponds: sshPortResponds,
		WaitForPortClosed: waitForPortClosed, ProcessIsMobdeskSSH: ProcessIsMobdeskSSH, FindProcess: findProcess,
		AcquireLock: acquireLock, EnsureSSHConfigured: EnsureSSHConfigured, EnsureIfconfig: ensureIfconfig, Addresses: LocalIPv4Addresses, Username: currentUsername,
		MkdirAll: os.MkdirAll, Chmod: os.Chmod, WriteFile: os.WriteFile, Lstat: os.Lstat, Readlink: os.Readlink, Symlink: os.Symlink,
		Executable: os.Executable, Abs: filepath.Abs, EvalSymlinks: filepath.EvalSymlinks,
	}
}
