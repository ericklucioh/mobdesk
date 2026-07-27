package workstation

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// ProcessIsMobdeskSSHFromPID reads the runtime PID before applying the injected process check.
func (d Dependencies) ProcessIsMobdeskSSHFromPID(pidPath, configPath string) bool {
	bytes, err := d.ReadFile(pidPath)
	if err != nil {
		return false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(bytes)))
	return err == nil && d.ProcessIsMobdeskSSH(pid, configPath)
}

func findProcess(pid int) (Process, error) { return os.FindProcess(pid) }

func ProcessIsMobdeskSSH(pid int, configPath string) bool {
	commandLine, err := processCommandLine(pid)
	if err != nil || !strings.Contains(commandLine, configPath) {
		return false
	}
	executable, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", pid))
	return err == nil && filepath.Base(executable) == "sshd"
}

func processCommandLine(pid int) (string, error) {
	bytes, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	return strings.ReplaceAll(string(bytes), "\x00", " "), err
}

func acquireLock(path string) (func(), error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("criar lock do SSH: %w", err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, fmt.Errorf("abrir lock do SSH: %w", err)
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("outro comando do Mobdesk já está em execução: %w", err)
	}
	return func() { _ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN); _ = file.Close() }, nil
}
