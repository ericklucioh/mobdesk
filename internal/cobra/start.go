package cobra

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

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
	configChanged, err := ensureSSHUbuntuCommand()
	if err != nil {
		return err
	}

	startWakeLock()
	if !portOpen(ctx, sshPort) {
		if err := startSSH(ctx); err != nil {
			return err
		}
		if !waitForPort(ctx, sshPort, 3*time.Second) {
			return fmt.Errorf("sshd não ficou disponível na porta %d", sshPort)
		}
	} else if configChanged {
		if err := reloadSSH(); err != nil {
			return err
		}
		fmt.Printf("Servidor SSH recarregado na porta %d.\n", sshPort)
	} else {
		fmt.Printf("Servidor SSH já está ativo na porta %d.\n", sshPort)
	}

	printAccessInstructions()
	fmt.Println("\nAbrindo Ubuntu...")
	return runInteractive(ctx, "proot-distro", "login", "ubuntu", "--", "bash", "-l")
}

func ensureSSHUbuntuCommand() (bool, error) {
	prefix := os.Getenv("PREFIX")
	if prefix == "" {
		prefix = "/data/data/com.termux/files/usr"
	}
	shellPath := filepath.Join(prefix, "bin", "sh")
	prootPath := filepath.Join(prefix, "bin", "proot-distro")
	wrapperPath := filepath.Join(prefix, "bin", "mobdesk-ssh-shell")
	configPath := filepath.Join(prefix, "etc", "ssh", "sshd_config")

	wrapper := fmt.Sprintf("#!%s\nexec %s login ubuntu -- bash -l\n", shellPath, prootPath)
	if err := os.WriteFile(wrapperPath, []byte(wrapper), 0o755); err != nil {
		return false, fmt.Errorf("criar shell SSH do Ubuntu: %w", err)
	}

	config, err := os.ReadFile(configPath)
	if err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("ler configuração do sshd: %w", err)
	}
	configText := string(config)
	directive := "ForceCommand " + wrapperPath
	if strings.Contains(configText, directive) {
		return false, validateSSHConfig(configPath)
	}

	configText = strings.TrimRight(configText, "\n") + "\n\n# Mobdesk: abrir sessões SSH diretamente no Ubuntu via PRoot.\n" + directive + "\n"
	if err := os.WriteFile(configPath, []byte(configText), 0o600); err != nil {
		return false, fmt.Errorf("configurar SSH para abrir o Ubuntu: %w", err)
	}
	return true, validateSSHConfig(configPath)
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

func reloadSSH() error {
	prefix := os.Getenv("PREFIX")
	if prefix == "" {
		prefix = "/data/data/com.termux/files/usr"
	}
	pidPath := filepath.Join(prefix, "var", "run", "sshd.pid")
	pidBytes, err := os.ReadFile(pidPath)
	if err != nil {
		return fmt.Errorf("recarregar sshd: ler PID em %s: %w", pidPath, err)
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(pidBytes)))
	if err != nil {
		return fmt.Errorf("recarregar sshd: PID inválido: %w", err)
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("recarregar sshd: localizar processo: %w", err)
	}
	if err := process.Signal(syscall.SIGHUP); err != nil {
		return fmt.Errorf("recarregar sshd: enviar SIGHUP: %w", err)
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

func startSSH(ctx context.Context) error {
	fmt.Printf("Iniciando servidor SSH na porta %d...\n", sshPort)
	command := exec.CommandContext(ctx, "sshd")
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("iniciar sshd: %w", err)
	}
	return nil
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

func waitForPort(ctx context.Context, port int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if portOpen(ctx, port) {
			return true
		}
		select {
		case <-ctx.Done():
			return false
		case <-time.After(100 * time.Millisecond):
		}
	}
	return portOpen(ctx, port)
}

func printAccessInstructions() {
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
	fmt.Println("\nServidor iniciado!")
	if len(addresses) == 0 {
		fmt.Printf("Acesse localmente: ssh -p %d %s@localhost\n", sshPort, name)
		return
	}
	fmt.Println("Acesse de outro computador:")
	for _, address := range addresses {
		fmt.Printf("ssh -p %d %s@%s\n", sshPort, name, address)
	}
}

func localIPv4Addresses() []string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	addresses := make([]string, 0)
	for _, networkInterface := range interfaces {
		if networkInterface.Flags&net.FlagUp == 0 || networkInterface.Flags&net.FlagLoopback != 0 {
			continue
		}
		interfaceAddresses, err := networkInterface.Addrs()
		if err != nil {
			continue
		}
		for _, interfaceAddress := range interfaceAddresses {
			var ip net.IP
			switch address := interfaceAddress.(type) {
			case *net.IPNet:
				ip = address.IP
			case *net.IPAddr:
				ip = address.IP
			}
			if ip4 := ip.To4(); ip4 != nil {
				addresses = append(addresses, ip4.String())
			}
		}
	}
	return addresses
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
