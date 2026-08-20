package workstation

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ericklucioh/mobdesk/internal/i18n"
)

func (s Service) Start(ctx context.Context) (info StartInfo, err error) {
	defer func() {
		if err != nil && i18n.ErrorCode(err) == "" {
			err = workstationError("start workstation", err)
		}
	}()
	info = StartInfo{Username: s.Deps.Username()}
	if info.Username == "" {
		info.Username = "user"
	}
	if _, err := s.Deps.Stat(s.Paths.SetupDone()); err != nil {
		if os.IsNotExist(err) {
			return info, workstationError("setup is incomplete; run mobdesk setup", nil)
		}
		return info, workstationError("check setup state", err)
	}
	if _, err := s.Deps.Stat(s.Paths.PasswordDone()); err != nil {
		if os.IsNotExist(err) {
			return info, workstationError("SSH password is not configured; run mobdesk setup", nil)
		}
		return info, workstationError("check SSH password configuration", err)
	}
	if err := s.Deps.EnsureSSHConfigured(s.Paths); err != nil {
		return info, err
	}
	if err := s.Deps.EnsureIfconfig(ctx, io.Discard, s.run); err != nil {
		info.Warnings = append(info.Warnings, workstationWarning("prepare local IP detection", err))
	}
	info.Addresses = s.Deps.Addresses()
	if err := s.Deps.WakeLock(); err != nil {
		info.Warnings = append(info.Warnings, workstationWarning("enable wake-lock", err))
	}
	alreadyRunning, err := s.ensureSSHRunning(ctx)
	if err != nil {
		_ = s.Deps.WakeUnlock()
		return info, err
	}
	info.AlreadyRunning = alreadyRunning
	return info, nil
}

func (s Service) Stop(ctx context.Context) (info StopInfo, err error) {
	defer func() {
		if err != nil && i18n.ErrorCode(err) == "" {
			err = workstationError("stop workstation", err)
		}
	}()
	info = StopInfo{}
	release, err := s.Deps.AcquireLock(s.Paths.SSHLock())
	if err != nil {
		return info, err
	}
	defer release()

	pidPath := s.Paths.SSHPID()
	pidBytes, err := s.Deps.ReadFile(pidPath)
	if os.IsNotExist(err) {
		if !s.Deps.PortOpen(ctx, SSHPort) {
			_ = s.Deps.WakeUnlock()
			info.AlreadyStopped = true
			return info, nil
		}
		return info, workstationError(fmt.Sprintf("port %d is occupied and Mobdesk PID is missing at %s", SSHPort, pidPath), nil)
	}
	if err != nil {
		return info, workstationError("read sshd PID", err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
	if err != nil {
		return info, workstationError("invalid sshd PID", err)
	}
	process, err := s.Deps.FindProcess(pid)
	if err != nil {
		return info, workstationError("find sshd process", err)
	}
	if !s.Deps.ProcessIsMobdeskSSH(pid, s.Paths.SSHConfig()) {
		if !s.Deps.PortOpen(ctx, SSHPort) {
			_ = s.Deps.Remove(pidPath)
			_ = s.Deps.WakeUnlock()
			info.AlreadyStopped, info.StaleState = true, true
			return info, nil
		}
		return info, i18n.NewError(i18n.ServiceWorkstationPID, "workstation_pid_mismatch", map[string]any{"PID": pid}, nil)
	}
	if err := process.Signal(syscall.SIGTERM); err != nil {
		if !s.Deps.PortOpen(ctx, SSHPort) {
			_ = s.Deps.Remove(pidPath)
			_ = s.Deps.WakeUnlock()
			info.AlreadyStopped = true
			return info, nil
		}
		return info, workstationError("stop sshd", err)
	}
	if !s.Deps.WaitForPortClosed(ctx, SSHPort, 3*time.Second) {
		return info, workstationError(fmt.Sprintf("sshd stopped but port %d is still active", SSHPort), nil)
	}
	_ = s.Deps.Remove(pidPath)
	_ = s.Deps.WakeUnlock()
	return info, nil
}

func (s Service) ensureSSHRunning(ctx context.Context) (bool, error) {
	release, err := s.Deps.AcquireLock(s.Paths.SSHLock())
	if err != nil {
		return false, err
	}
	defer release()
	if s.Deps.ProcessIsMobdeskSSHFromPID(s.Paths.SSHPID(), s.Paths.SSHConfig()) {
		if !s.Deps.SSHResponds(ctx, SSHPort) {
			return false, workstationError(fmt.Sprintf("Mobdesk SSH process exists but port %d does not respond as SSH", SSHPort), nil)
		}
		return true, nil
	}
	if s.Deps.PortOpen(ctx, SSHPort) {
		return false, workstationError(fmt.Sprintf("port %d is occupied by another process", SSHPort), nil)
	}
	_ = s.Deps.Remove(s.Paths.SSHPID())
	if err := s.Deps.StartSSHD(ctx, s.Paths.SSHConfig(), s.Paths.SSHLog()); err != nil {
		return false, workstationError("start sshd", err)
	}
	if !waitForSSH(ctx, s.Deps.SSHResponds, SSHPort, 3*time.Second) {
		return false, workstationError(fmt.Sprintf("sshd did not become available on port %d", SSHPort), nil)
	}
	return false, nil
}

func waitForSSH(ctx context.Context, responds func(context.Context, int) bool, port int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if responds(ctx, port) {
			return true
		}
		select {
		case <-ctx.Done():
			return false
		case <-time.After(100 * time.Millisecond):
		}
	}
	return responds(ctx, port)
}
