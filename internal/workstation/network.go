package workstation

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/ericklucioh/mobdesk/internal/executil"
)

func ensureIfconfig(ctx context.Context, out io.Writer, run func(context.Context, string, ...string) error) error {
	if _, err := executil.Resolve("ifconfig"); err == nil {
		return nil
	}
	fmt.Fprintln(out, "ifconfig não encontrado; instalando net-tools...")
	return run(ctx, "pkg", "install", "-y", "-o", "Dpkg::Options::=--force-confold", "net-tools")
}

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
