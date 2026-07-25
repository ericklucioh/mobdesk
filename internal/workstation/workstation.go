// Package workstation manages the SSH-backed development workstation.
package workstation

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ericklucioh/mobdesk/internal/executil"
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
	Warnings       []string
}

type StopInfo struct {
	AlreadyStopped bool
	StaleState     bool
}

func New(p paths.Paths) Service {
	return Service{Paths: p, Deps: defaultDependencies()}
}

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
		return info, fmt.Errorf("Ubuntu não está disponível; execute mobdesk setup: %w", err)
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

func (s Service) run(ctx context.Context, name string, args ...string) error {
	if err := s.Deps.Run(ctx, name, args...); err != nil {
		return fmt.Errorf("comando %q falhou: %w", name, err)
	}
	return nil
}

func ensureIfconfig(ctx context.Context, out io.Writer, run func(context.Context, string, ...string) error) error {
	if _, err := executil.Resolve("ifconfig"); err == nil {
		return nil
	}
	fmt.Fprintln(out, "ifconfig não encontrado; instalando net-tools...")
	return run(ctx, "pkg", "install", "-y", "-o", "Dpkg::Options::=--force-confold", "net-tools")
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

// ProcessIsMobdeskSSHFromPID reads the runtime PID before applying the injected process check.
func (d Dependencies) ProcessIsMobdeskSSHFromPID(pidPath, configPath string) bool {
	bytes, err := d.ReadFile(pidPath)
	if err != nil {
		return false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(bytes)))
	return err == nil && d.ProcessIsMobdeskSSH(pid, configPath)
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

func defaultDependencies() Dependencies {
	return Dependencies{
		Stat: os.Stat, ReadFile: os.ReadFile, Remove: os.Remove, Run: runCommand, StartSSHD: startSSHD,
		WakeLock: wakeLock, WakeUnlock: wakeUnlock, PortOpen: portOpen, SSHResponds: sshPortResponds,
		WaitForPortClosed: waitForPortClosed, ProcessIsMobdeskSSH: ProcessIsMobdeskSSH, FindProcess: findProcess,
		AcquireLock: acquireLock, EnsureSSHConfigured: EnsureSSHConfigured, EnsureIfconfig: ensureIfconfig, Addresses: LocalIPv4Addresses, Username: currentUsername,
		MkdirAll: os.MkdirAll, WriteFile: os.WriteFile, Lstat: os.Lstat, Readlink: os.Readlink, Symlink: os.Symlink,
		Executable: os.Executable, Abs: filepath.Abs, EvalSymlinks: filepath.EvalSymlinks,
	}
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
		return fmt.Errorf("resolver sshd: %w", err)
	}
	command.Stdin, command.Stdout, command.Stderr = os.Stdin, os.Stdout, os.Stderr
	return command.Run()
}
func wakeLock() error {
	command, err := executil.Command("termux-wake-lock")
	if err != nil {
		return fmt.Errorf("termux-wake-lock não está disponível neste ambiente")
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
func findProcess(pid int) (Process, error) { return os.FindProcess(pid) }
func portOpen(ctx context.Context, port int) bool {
	connection, err := (&net.Dialer{Timeout: 250 * time.Millisecond}).DialContext(ctx, "tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		return false
	}
	_ = connection.Close()
	return true
}
func sshPortResponds(ctx context.Context, port int) bool {
	connection, err := (&net.Dialer{Timeout: 250 * time.Millisecond}).DialContext(ctx, "tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		return false
	}
	defer connection.Close()
	_ = connection.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
	banner, err := bufio.NewReader(connection).ReadString('\n')
	return err == nil && strings.HasPrefix(banner, "SSH-")
}
func waitForPortClosed(ctx context.Context, port int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !portOpen(ctx, port) {
			return true
		}
		select {
		case <-ctx.Done():
			return false
		case <-time.After(100 * time.Millisecond):
		}
	}
	return !portOpen(ctx, port)
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

var ifconfigIPv4Pattern = regexp.MustCompile(`^\s+inet\s+((?:[0-9]{1,3}\.){3}[0-9]{1,3})\b`)

func LocalIPv4Addresses() []string {
	ifconfig, err := executil.Resolve("ifconfig")
	if err != nil {
		return nil
	}
	output, err := exec.Command(ifconfig).Output()
	if err != nil {
		return nil
	}
	return ExtractIPv4Addresses(string(output))
}
func ExtractIPv4Addresses(output string) []string {
	preferred, others := []string{}, []string{}
	interfaceName := ""
	for _, line := range strings.Split(output, "\n") {
		if len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				interfaceName = strings.TrimSuffix(fields[0], ":")
			}
		}
		match := ifconfigIPv4Pattern.FindStringSubmatch(line)
		if len(match) != 2 || match[1] == "127.0.0.1" || net.ParseIP(match[1]) == nil {
			continue
		}
		if interfaceName == "wlan0" {
			preferred = appendUnique(preferred, match[1])
		} else {
			others = appendUnique(others, match[1])
		}
	}
	return append(preferred, others...)
}
func appendUnique(addresses []string, address string) []string {
	for _, existing := range addresses {
		if existing == address {
			return addresses
		}
	}
	return append(addresses, address)
}

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
		return nil, fmt.Errorf("outro comando do Mobdesk já está controlando o SSH: %w", err)
	}
	return func() { _ = syscall.Flock(int(file.Fd()), syscall.LOCK_UN); _ = file.Close() }, nil
}
