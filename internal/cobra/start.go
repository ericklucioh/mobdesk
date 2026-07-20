package cobra

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"
)

const sshPort = 8022

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "iniciar o ambiente e o servidor SSH",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStart(cmd.Context())
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "parar o servidor SSH",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runStop(cmd.Context())
	},
}

func runStart(ctx context.Context) error {
	if _, err := os.Stat(os.ExpandEnv("$HOME/.local/share/mobdesk/setup.done")); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("setup ainda não foi concluído; execute: mobdesk setup")
		}
		return fmt.Errorf("verificar estado do setup: %w", err)
	}
	if _, err := os.Stat(os.ExpandEnv("$HOME/.local/share/mobdesk/password.done")); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("senha SSH ainda não foi configurada; execute: mobdesk setup")
		}
		return fmt.Errorf("verificar configuração da senha SSH: %w", err)
	}

	if err := runCommand(ctx, "proot-distro", "login", "ubuntu", "--", "true"); err != nil {
		return fmt.Errorf("Ubuntu não está disponível; execute mobdesk setup: %w", err)
	}
	if err := ensureMobdeskSSH(); err != nil {
		return err
	}
	if err := ensureIfconfig(ctx); err != nil {
		fmt.Printf("Aviso: não foi possível preparar a detecção do IP local: %v\n", err)
	}

	startWakeLock()
	if err := ensureSSHRunning(ctx); err != nil {
		unlockWakeLock()
		return err
	}

	printAccessInstructions()
	fmt.Println("\nAbrindo Ubuntu...")
	return runInteractive(ctx, "proot-distro", "login", "ubuntu", "--", "bash", "-l")
}

func runStop(ctx context.Context) error {
	lock, err := acquireSSHLock()
	if err != nil {
		return err
	}
	defer lock.release()

	paths := mobdeskSSHPaths()
	pidPath := paths.pid

	pidBytes, err := os.ReadFile(pidPath)
	if os.IsNotExist(err) {
		if !portOpen(ctx, sshPort) {
			unlockWakeLock()
			fmt.Println("Servidor SSH já está parado.")
			return nil
		}
		return fmt.Errorf("a porta %d está ocupada, mas o PID do Mobdesk não foi encontrado em %s", sshPort, pidPath)
	}
	if err != nil {
		return fmt.Errorf("ler PID do sshd: %w", err)
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
	if err != nil {
		return fmt.Errorf("PID do sshd inválido: %w", err)
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("localizar processo do sshd: %w", err)
	}
	if !processIsMobdeskSSH(pid, paths.config) {
		if !portOpen(ctx, sshPort) {
			_ = os.Remove(pidPath)
			unlockWakeLock()
			fmt.Println("Servidor SSH já estava parado; estado obsoleto removido.")
			return nil
		}
		return fmt.Errorf("o PID %d não pertence ao servidor SSH do Mobdesk", pid)
	}
	if err := process.Signal(syscall.SIGTERM); err != nil {
		if !portOpen(ctx, sshPort) {
			_ = os.Remove(pidPath)
			unlockWakeLock()
			fmt.Println("Servidor SSH já estava parado.")
			return nil
		}
		return fmt.Errorf("parar sshd: %w", err)
	}

	if !waitForPortClosed(ctx, sshPort, 3*time.Second) {
		return fmt.Errorf("sshd recebeu o sinal de parada, mas a porta %d ainda está ativa", sshPort)
	}
	_ = os.Remove(pidPath)
	unlockWakeLock()
	fmt.Println("Servidor SSH parado.")
	return nil
}

func validateSSHConfig(configPath string) error {
	command := exec.Command("sshd", "-t", "-f", configPath)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("configuração do sshd inválida: %w", err)
	}
	return nil
}

func startWakeLock() {
	if _, err := exec.LookPath("termux-wake-lock"); err != nil {
		fmt.Println("Aviso: termux-wake-lock não está disponível neste ambiente.")
		return
	}
	if err := exec.Command("termux-wake-lock").Run(); err != nil {
		fmt.Printf("Aviso: não foi possível ativar o wake-lock: %v\n", err)
	}
}

func unlockWakeLock() {
	if _, err := exec.LookPath("termux-wake-unlock"); err != nil {
		return
	}
	if err := exec.Command("termux-wake-unlock").Run(); err != nil {
		fmt.Printf("Aviso: não foi possível liberar o wake-lock: %v\n", err)
	}
}

func startSSH(ctx context.Context) error {
	fmt.Printf("Iniciando servidor SSH na porta %d...\n", sshPort)
	paths := mobdeskSSHPaths()
	if !mobdeskSSHProcess() {
		_ = os.Remove(paths.pid)
	}
	command := exec.CommandContext(ctx, "sshd", "-f", paths.config, "-E", paths.log)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("iniciar sshd: %w", err)
	}
	return nil
}

func ensureSSHRunning(ctx context.Context) error {
	lock, err := acquireSSHLock()
	if err != nil {
		return err
	}
	defer lock.release()

	if mobdeskSSHProcess() {
		if !sshPortResponds(ctx, sshPort) {
			return fmt.Errorf("o processo SSH do Mobdesk existe, mas a porta %d não responde como SSH", sshPort)
		}
		fmt.Printf("Servidor SSH já está ativo na porta %d.\n", sshPort)
		return nil
	}
	if portOpen(ctx, sshPort) {
		return fmt.Errorf("a porta %d está ocupada por outro processo", sshPort)
	}
	if err := startSSH(ctx); err != nil {
		return err
	}
	if !waitForSSH(ctx, sshPort, 3*time.Second) {
		return fmt.Errorf("sshd não ficou disponível na porta %d", sshPort)
	}
	return nil
}

func mobdeskSSHProcess() bool {
	paths := mobdeskSSHPaths()
	bytes, err := os.ReadFile(paths.pid)
	if err != nil {
		return false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(bytes)))
	return err == nil && processIsMobdeskSSH(pid, paths.config)
}

func portOpen(ctx context.Context, port int) bool {
	dialer := net.Dialer{Timeout: 250 * time.Millisecond}
	connection, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		return false
	}
	_ = connection.Close()
	return true
}

func waitForSSH(ctx context.Context, port int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if sshPortResponds(ctx, port) {
			return true
		}
		select {
		case <-ctx.Done():
			return false
		case <-time.After(100 * time.Millisecond):
		}
	}
	return sshPortResponds(ctx, port)
}

func sshPortResponds(ctx context.Context, port int) bool {
	dialer := net.Dialer{Timeout: 250 * time.Millisecond}
	connection, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
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

func printAccessInstructions() {
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7DD3FC"))
	commandStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#A7F3D0"))

	name := os.Getenv("USER")
	if name == "" {
		if current, err := user.Current(); err == nil {
			name = current.Username
		}
	}
	if name == "" {
		name = "usuario"
	}

	addresses := localIPv4Addresses()
	fmt.Println("\nSERVIDOR INICIADO!")
	if len(addresses) == 0 {
		fmt.Println(headerStyle.Render("ACESSO LOCAL VIA SSH"))
		command := fmt.Sprintf("ssh -p %d %s@localhost", sshPort, name)
		fmt.Println(commandStyle.Render(command))
		return
	}
	fmt.Println(headerStyle.Render("ACESSO REMOTO VIA SSH"))
	for _, address := range addresses {
		command := fmt.Sprintf("ssh -p %d %s@%s", sshPort, name, address)
		fmt.Println(commandStyle.Render(command))
	}
}

func localIPv4Addresses() []string {
	ifconfig, err := exec.LookPath("ifconfig")
	if err != nil {
		return nil
	}
	output, err := exec.Command(ifconfig).Output()
	if err != nil {
		return nil
	}
	return extractIPv4Addresses(string(output))
}

var ifconfigIPv4Pattern = regexp.MustCompile(`^\s+inet\s+((?:[0-9]{1,3}\.){3}[0-9]{1,3})\b`)

func extractIPv4Addresses(output string) []string {
	preferred := make([]string, 0)
	others := make([]string, 0)
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

func ensureIfconfig(ctx context.Context) error {
	if _, err := exec.LookPath("ifconfig"); err == nil {
		return nil
	}
	fmt.Println("ifconfig não encontrado; instalando net-tools...")
	return runCommand(ctx, "pkg", "install", "-y", "-o", "Dpkg::Options::=--force-confold", "net-tools")
}

func appendUnique(addresses []string, address string) []string {
	for _, existing := range addresses {
		if existing == address {
			return addresses
		}
	}
	return append(addresses, address)
}

func runInteractive(ctx context.Context, name string, args ...string) error {
	fmt.Printf("\n$ %s %s\n", name, strings.Join(args, " "))
	command := exec.CommandContext(ctx, name, args...)
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("comando interativo %q falhou: %w", name, err)
	}
	return nil
}
