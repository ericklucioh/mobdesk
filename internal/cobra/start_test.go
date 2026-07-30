package cobra

import (
	"testing"

	"github.com/ericklucioh/mobdesk/internal/workstation"
)

func TestExtractIPv4AddressesPrefersWLAN(t *testing.T) {
	output := `Warning: cannot open /proc/net/dev (Permission denied). Limited output.
lo: flags=73<UP,LOOPBACK,RUNNING>  mtu 65536
        inet 127.0.0.1  netmask 255.0.0.0

wlan0: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500
        inet 192.168.3.228  netmask 255.255.255.0  broadcast 192.168.3.255
`

	want := []string{"192.168.3.228"}
	got := workstation.ExtractIPv4Addresses(output)
	assertAddresses(t, got, want)
}

func TestExtractIPv4AddressesIgnoresLoopback(t *testing.T) {
	output := `lo: flags=73<UP,LOOPBACK,RUNNING>  mtu 65536
        inet 127.0.0.1  netmask 255.0.0.0
`

	got := workstation.ExtractIPv4Addresses(output)
	if len(got) != 0 {
		t.Fatalf("expected no addresses, got %v", got)
	}
}

func TestExtractIPv4AddressesUsesOtherInterfacesAsFallback(t *testing.T) {
	output := `rmnet0: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500
        inet 10.23.45.6  netmask 255.255.255.0
`

	want := []string{"10.23.45.6"}
	got := workstation.ExtractIPv4Addresses(output)
	assertAddresses(t, got, want)
}

func TestExtractIPv4AddressesHandlesEmptyOutput(t *testing.T) {
	if got := workstation.ExtractIPv4Addresses(""); len(got) != 0 {
		t.Fatalf("expected no addresses, got %v", got)
	}
}

func assertAddresses(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("addresses differ: got %v, want %v", got, want)
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("address differs at position %d: got %q, want %q", index, got[index], want[index])
		}
	}
}
