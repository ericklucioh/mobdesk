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

// SSHPort is the dedicated port used by the Mobdesk-managed SSH server.
const SSHPort = 8022

// Process is the subset of an operating-system process used during shutdown.
type Process interface {
	Signal(os.Signal) error
}

// Dependencies provides the operating-system boundaries used by Service.
// Tests replace only the functions needed for their scenario.
type Dependencies struct {
	// Filesystem.
	Stat         func(string) (os.FileInfo, error)
	ReadFile     func(string) ([]byte, error)
	Remove       func(string) error
	MkdirAll     func(string, os.FileMode) error
	Chmod        func(string, os.FileMode) error
	WriteFile    func(string, []byte, os.FileMode) error
	Lstat        func(string) (os.FileInfo, error)
	Readlink     func(string) (string, error)
	Symlink      func(string, string) error
	Executable   func() (string, error)
	Abs          func(string) (string, error)
	EvalSymlinks func(string) (string, error)

	// Process and SSH.
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

	// Network and Android runtime.
	Addresses       func() []string
	Username        func() string
	AndroidTimezone func(context.Context) string
}

// Service manages the Termux workstation state and Mobdesk-owned SSH server.
type Service struct {
	Paths paths.Paths
	Deps  Dependencies
}

// StartInfo summarizes the result of starting the workstation.
type StartInfo struct {
	AlreadyRunning bool
	Addresses      []string
	Username       string
	Warnings       []error
}

// StopInfo summarizes the result of stopping the workstation.
type StopInfo struct {
	AlreadyStopped bool
	StaleState     bool
}

// New creates a workstation service rooted at the provided Termux paths.
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
		Stat:         os.Stat,
		ReadFile:     os.ReadFile,
		Remove:       os.Remove,
		MkdirAll:     os.MkdirAll,
		Chmod:        os.Chmod,
		WriteFile:    os.WriteFile,
		Lstat:        os.Lstat,
		Readlink:     os.Readlink,
		Symlink:      os.Symlink,
		Executable:   os.Executable,
		Abs:          filepath.Abs,
		EvalSymlinks: filepath.EvalSymlinks,

		Run:                 runCommand,
		StartSSHD:           startSSHD,
		WakeLock:            wakeLock,
		WakeUnlock:          wakeUnlock,
		PortOpen:            portOpen,
		SSHResponds:         sshPortResponds,
		WaitForPortClosed:   waitForPortClosed,
		ProcessIsMobdeskSSH: ProcessIsMobdeskSSH,
		FindProcess:         findProcess,
		AcquireLock:         acquireLock,
		EnsureSSHConfigured: EnsureSSHConfigured,
		EnsureIfconfig:      ensureIfconfig,

		Addresses:       LocalIPv4Addresses,
		Username:        currentUsername,
		AndroidTimezone: androidTimezone,
	}
}
