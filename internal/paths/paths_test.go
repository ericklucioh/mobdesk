package paths

import "testing"

func TestTermuxLayoutUsesNativeWorkspace(t *testing.T) {
	p := New("/home/mobdesk", "/termux")
	if got, want := p.Workspace(), "/home/mobdesk/workspace"; got != want {
		t.Fatalf("workspace = %q, want %q", got, want)
	}
	if got, want := p.ShellConfig(), "/home/mobdesk/.config/mobdesk/shell.bash"; got != want {
		t.Fatalf("shell config = %q, want %q", got, want)
	}
}
