package install

import "testing"

func TestResolvePiProfile(t *testing.T) {
	profile, ok := Resolve("pi")
	if !ok {
		t.Fatal("Pi profile was not found")
	}
	if profile.Package != "@earendil-works/pi-coding-agent@0.84.4" {
		t.Fatalf("unexpected Pi package: %q", profile.Package)
	}
	if profile.Executable != "pi" || profile.InstallKind != "npm" || !profile.UserBin {
		t.Fatalf("unexpected Pi install profile: %+v", profile)
	}
	if len(profile.Requires) != 1 || profile.Requires[0] != "node" {
		t.Fatalf("unexpected Pi dependencies: %v", profile.Requires)
	}

	alias, ok := Resolve("pi-coding-agent")
	if !ok || alias.Name != "pi" {
		t.Fatalf("Pi alias did not resolve: %+v", alias)
	}
}
