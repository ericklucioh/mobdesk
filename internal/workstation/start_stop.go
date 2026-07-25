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
)

func (s Service) Start(ctx context.Context) (StartInfo, error) {
	info := StartInfo{Username: s.Deps.Username()}
	if info.Username == "" {
		info.Username = "usuario"
	}
	if _, err := s.Deps.Stat(s.Paths.SetupDone()); err != nil {
		if os.IsNotExist(err) {
			return info, fmt.Errorf("setup ainda não foi concluído; execute: mobdesk setup")
		}
		return info, fmt.Errorf("verificar estado do setup: %w", err)
	}
	if _, err := s.Deps.Stat(s.Paths.PasswordDone()); err != nil {
		if os.IsNotExist(err) {
			return info, fmt.Errorf("senha SSH ainda não foi configurada; execute: mobdesk setup")
		}
		return info, fmt.Errorf("verificar configuração da senha SSH: %w", err)
	}
	if err := s.run(ctx, "proot-distro", "login", "ubuntu", "--", "true"); err != nil {
		return info, fmt.Errorf("ubuntu não está disponível; execute mobdesk setup: %w", err)
	}
	if err := s.Deps.EnsureSSHConfigured(s.Paths); err != nil {
		return info, err
	}
	if err := s.Deps.EnsureIfconfig(ctx, io.Discard, s.run); err != nil {
		info.Warnings = append(info.Warnings, fmt.Sprintf("não foi possível preparar a detecção do IP local: %v", err))
	}
	info.Addresses = s.Deps.Addresses()
	if err := s.Deps.WakeLock(); err != nil {
		info.Warnings = append(info.Warnings, fmt.Sprintf("não foi possível ativar o wake-lock: %v", err))
	}
	alreadyRunning, err := s.ensureSSHRunning(ctx)
	if err != nil {
		_ = s.Deps.WakeUnlock()
		return info, err
	}
	info.AlreadyRunning = alreadyRunning
	return info, nil
}

func (s Service) Stop(ctx context.Context) (StopInfo, error) {
	info := StopInfo{}
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
		return info, fmt.Errorf("a porta %d está ocupada, mas o PID do Mobdesk não foi encontrado em %s", SSHPort, pidPath)
	}
	if err != nil {
		return info, fmt.Errorf("ler PID do sshd: %w", err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
	if err != nil {
		return info, fmt.Errorf("PID do sshd inválido: %w", err)
	}
	process, err := s.Deps.FindProcess(pid)
	if err != nil {
		return info, fmt.Errorf("localizar processo do sshd: %w", err)
	}
	if !s.Deps.ProcessIsMobdeskSSH(pid, s.Paths.SSHConfig()) {
		if !s.Deps.PortOpen(ctx, SSHPort) {
			_ = s.Deps.Remove(pidPath)
			_ = s.Deps.WakeUnlock()
			info.AlreadyStopped, info.StaleState = true, true
			return info, nil
		}
		return info, fmt.Errorf("o PID %d não pertence ao servidor SSH do Mobdesk", pid)
	}
	if err := process.Signal(syscall.SIGTERM); err != nil {
		if !s.Deps.PortOpen(ctx, SSHPort) {
			_ = s.Deps.Remove(pidPath)
			_ = s.Deps.WakeUnlock()
			info.AlreadyStopped = true
			return info, nil
		}
		return info, fmt.Errorf("parar sshd: %w", err)
	}
	if !s.Deps.WaitForPortClosed(ctx, SSHPort, 3*time.Second) {
		return info, fmt.Errorf("sshd recebeu o sinal de parada, mas a porta %d ainda está ativa", SSHPort)
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
			return false, fmt.Errorf("o processo SSH do Mobdesk existe, mas a porta %d não responde como SSH", SSHPort)
		}
		return true, nil
	}
	if s.Deps.PortOpen(ctx, SSHPort) {
		return false, fmt.Errorf("a porta %d está ocupada por outro processo", SSHPort)
	}
	_ = s.Deps.Remove(s.Paths.SSHPID())
	if err := s.Deps.StartSSHD(ctx, s.Paths.SSHConfig(), s.Paths.SSHLog()); err != nil {
		return false, fmt.Errorf("iniciar sshd: %w", err)
	}
	if !waitForSSH(ctx, s.Deps.SSHResponds, SSHPort, 3*time.Second) {
		return false, fmt.Errorf("sshd não ficou disponível na porta %d", SSHPort)
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
