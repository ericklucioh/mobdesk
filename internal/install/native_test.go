package install

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ericklucioh/mobdesk/internal/paths"
)

type nativeRunner struct {
	commands []string
	versions int
}

type javaRunner struct {
	commands  []string
	installed bool
}

type mavenRunner struct {
	packages map[string]bool
}

type canceledRunner struct{}

type pipxCanceledRunner struct{}

func (canceledRunner) Run(ctx context.Context, _ string, _ ...string) CommandResult {
	return CommandResult{Err: ctx.Err()}
}

func (pipxCanceledRunner) Run(ctx context.Context, name string, args ...string) CommandResult {
	if name == "python" && len(args) > 0 && args[0] == "--version" {
		return CommandResult{Stdout: []byte("Python 3\n")}
	}
	return CommandResult{Err: ctx.Err()}
}

func (r *javaRunner) Run(_ context.Context, name string, args ...string) CommandResult {
	r.commands = append(r.commands, name+" "+strings.Join(args, " "))
	if name == "pkg" {
		r.installed = true
		return CommandResult{}
	}
	if !r.installed {
		return CommandResult{Err: errors.New("java missing")}
	}
	if name == "java" && len(args) > 0 && args[0] == "-XshowSettings:properties" {
		return CommandResult{Stderr: []byte("Property settings:\n    java.home = /data/data/com.termux/files/usr/lib/jvm/java-21-openjdk\n")}
	}
	return CommandResult{Stdout: []byte(name + " 21.0.12\n")}
}

func (r *mavenRunner) Run(_ context.Context, name string, args ...string) CommandResult {
	if r.packages == nil {
		r.packages = make(map[string]bool)
	}
	if name == "pkg" {
		for _, arg := range args {
			if arg == "openjdk-21" || arg == "maven" {
				r.packages[arg] = true
			}
		}
		return CommandResult{}
	}
	if name == "java" && len(args) > 0 && args[0] == "-XshowSettings:properties" {
		if !r.packages["openjdk-21"] {
			return CommandResult{Err: errors.New("java missing")}
		}
		return CommandResult{Stderr: []byte("java.home = /data/data/com.termux/files/usr/lib/jvm/java-21-openjdk\n")}
	}
	if name == "java" || name == "javac" || name == "jar" {
		if !r.packages["openjdk-21"] {
			return CommandResult{Err: errors.New("java missing")}
		}
		return CommandResult{Stdout: []byte(name + " 21.0.12\n")}
	}
	if name == "mvn" {
		if !r.packages["maven"] {
			return CommandResult{Err: errors.New("maven missing")}
		}
		return CommandResult{Stdout: []byte("Apache Maven 3.9.16\n")}
	}
	return CommandResult{}
}

func (r *nativeRunner) Run(_ context.Context, name string, args ...string) CommandResult {
	r.commands = append(r.commands, name+" "+strings.Join(args, " "))
	if name == "git" {
		r.versions++
		if r.versions == 1 {
			return CommandResult{Err: errors.New("git missing")}
		}
		return CommandResult{Stdout: []byte("git version 1.0\n")}
	}
	return CommandResult{}
}

func TestInstallUsesNativePkg(t *testing.T) {
	runner := &nativeRunner{}
	p := paths.New(t.TempDir(), "")
	result, err := Install(context.Background(), "git", Options{
		Paths:       p,
		Runner:      runner,
		Now:         time.Now,
		StorageFree: func(string) (int64, error) { return StorageWarningBytes + 1, nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Installed || !result.Changed {
		t.Fatalf("unexpected installation result: %+v", result)
	}
	if !containsCommand(runner.commands, "pkg install -y git") {
		t.Fatalf("native pkg command was not used: %v", runner.commands)
	}
}

func TestCatalogUsesOnlyNativeStrategies(t *testing.T) {
	want := map[string]bool{
		"git": true, "neovim": true, "tmux": true, "go": true, "python": true,
		"java": true, "maven": true, "kotlin": true, "gradle": true, "node": true, "c": true, "cpp": true, "lua": true, "gh": true,
		"zellij": true, "lazygit": true, "tree": true, "htop": true, "ncdu": true, "inxi": true, "yazi": true, "micro": true,
		"tuifi": true,
	}
	for _, profile := range Tools() {
		if !want[profile.Name] {
			t.Fatalf("unexpected catalog profile %q", profile.Name)
		}
		delete(want, profile.Name)
		if profile.InstallKind != "pkg" && profile.InstallKind != "pipx" {
			t.Fatalf("profile %q has unsupported native strategy: %+v", profile.Name, profile)
		}
		if profile.Package == "" || profile.Executable == "" {
			t.Fatalf("profile %q is incomplete: %+v", profile.Name, profile)
		}
	}
	if len(want) > 0 {
		t.Fatalf("missing catalog profiles: %v", want)
	}
}

func TestNativePkgProfilesUseOfficialPackages(t *testing.T) {
	for name, want := range map[string]string{
		"gradle": "gradle", "inxi": "inxi", "kotlin": "kotlin", "lazygit": "lazygit", "yazi": "yazi", "zellij": "zellij",
	} {
		profile, ok := Resolve(name)
		if !ok || profile.InstallKind != "pkg" || profile.Package != want {
			t.Fatalf("native profile %q = %+v, %t", name, profile, ok)
		}
	}
	for _, name := range []string{"kotlin", "gradle"} {
		profile, _ := Resolve(name)
		if !sameStrings(profile.Requires, []string{"java"}) {
			t.Fatalf("%s requirements = %v", name, profile.Requires)
		}
	}
}

func TestTuifiUsesPrivatePipxStrategy(t *testing.T) {
	profile, ok := Resolve("tuifi")
	if !ok || profile.InstallKind != "pipx" || !profile.UserBin || !sameStrings(profile.Requires, []string{"python"}) {
		t.Fatalf("unexpected TUIFI profile: %+v", profile)
	}
}

func TestPublishManagedLinkRejectsUserFile(t *testing.T) {
	directory := t.TempDir()
	target := filepath.Join(directory, "target")
	link := filepath.Join(directory, "tool")
	if err := os.WriteFile(target, []byte("tool"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(link, []byte("user file"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := publishManagedLink(link, target); err == nil {
		t.Fatal("managed link overwrote a user file")
	}
}

func TestRemoveManagedLinkRejectsReplacement(t *testing.T) {
	directory := t.TempDir()
	target := filepath.Join(directory, "target")
	link := filepath.Join(directory, "tool")
	if err := os.WriteFile(target, []byte("tool"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(link); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(link, []byte("user file"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := removeManagedLink(link, target); err == nil {
		t.Fatal("managed link removal deleted a replacement file")
	}
}

func TestFileSHA256ChangesWithManagedFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tool")
	if err := os.WriteFile(path, []byte("first"), 0o700); err != nil {
		t.Fatal(err)
	}
	first, err := fileSHA256(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("second"), 0o700); err != nil {
		t.Fatal(err)
	}
	second, err := fileSHA256(path)
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("managed file hash did not change")
	}
}

func TestUninstallTuifiPreservesModifiedExecutable(t *testing.T) {
	p := paths.New(t.TempDir(), "/data/data/com.termux/files/usr")
	profile, _ := Resolve("tuifi")
	link, target, directory := managedExecutablePaths(p, profile)
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("original"), 0o700); err != nil {
		t.Fatal(err)
	}
	digest, err := fileSHA256(target)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(link), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(target, []byte("modified"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(p.InstallationsDir(), 0o700); err != nil {
		t.Fatal(err)
	}
	record := InstallationRecord{
		Name:                profile.Name,
		Package:             profile.Package,
		Strategy:            profile.InstallKind,
		State:               "installed",
		Source:              "mobdesk",
		InstalledFiles:      []string{link},
		InstalledDirs:       []string{directory},
		InstalledFileHashes: map[string]string{target: digest},
	}
	if err := saveRecord(p.InstallationsDir(), record); err != nil {
		t.Fatal(err)
	}
	runner := &nativeRunner{}
	result, err := Uninstall(context.Background(), "tuifi", Options{
		Paths: p, Runner: runner, Now: time.Now,
		StorageFree: func(string) (int64, error) { return StorageWarningBytes + 1, nil },
	})
	if err == nil || !sameStrings(result.Conflicts, []string{target}) {
		t.Fatalf("modified TUIFI uninstall = %+v, %v", result, err)
	}
	if len(runner.commands) != 0 {
		t.Fatalf("modified executable ran pipx uninstall: %v", runner.commands)
	}
}

func TestInstallTuifiRecordsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	p := paths.New(t.TempDir(), "/data/data/com.termux/files/usr")
	result, err := Install(ctx, "tuifi", Options{
		Paths:       p,
		Runner:      pipxCanceledRunner{},
		Now:         time.Now,
		StorageFree: func(string) (int64, error) { return StorageWarningBytes + 1, nil },
	})
	if err == nil || result.State != "failed" {
		t.Fatalf("canceled TUIFI installation = %+v, %v", result, err)
	}
	record, loadErr := loadInstallationRecord(p, "tuifi")
	if loadErr != nil || record.State != "failed" {
		t.Fatalf("canceled TUIFI installation state = %+v, %v", record, loadErr)
	}
}

func TestJavaProfileRequiresJDKCommands(t *testing.T) {
	profile, ok := Resolve("java")
	if !ok {
		t.Fatal("java profile is missing")
	}
	want := []ExecutableSpec{{Name: "java", VersionArg: []string{"--version"}}, {Name: "javac", VersionArg: []string{"--version"}}, {Name: "jar", VersionArg: []string{"--version"}}}
	if profile.Package != "openjdk-21" || !sameExecutableSpecs(profileExecutables(profile), want) {
		t.Fatalf("unexpected Java profile: %+v", profile)
	}
}

func TestMavenProfileRequiresJava(t *testing.T) {
	profile, ok := Resolve("maven")
	if !ok {
		t.Fatal("maven profile is missing")
	}
	if profile.Package != "maven" || !sameStrings(profile.Requires, []string{"java"}) {
		t.Fatalf("unexpected Maven profile: %+v", profile)
	}
	runner := &mavenRunner{}
	result, err := Install(context.Background(), "maven", Options{
		Paths:       paths.New(t.TempDir(), "/data/data/com.termux/files/usr"),
		Runner:      runner,
		Now:         time.Now,
		StorageFree: func(string) (int64, error) { return StorageWarningBytes + 1, nil },
	})
	if err != nil || !result.Installed || !runner.packages["openjdk-21"] || !runner.packages["maven"] {
		t.Fatalf("Maven installation = %+v, %v, packages=%v", result, err, runner.packages)
	}
}

func TestInstallJavaRecordsTermuxJavaHome(t *testing.T) {
	runner := &javaRunner{}
	prefix := "/data/data/com.termux/files/usr"
	p := paths.New(t.TempDir(), prefix)
	result, err := Install(context.Background(), "java", Options{
		Paths:       p,
		Runner:      runner,
		Now:         time.Now,
		StorageFree: func(string) (int64, error) { return StorageWarningBytes + 1, nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Installed || !result.Changed || result.JavaHome != prefix+"/lib/jvm/java-21-openjdk" {
		t.Fatalf("unexpected Java installation result: %+v", result)
	}
	if !containsCommand(runner.commands, "pkg install -y openjdk-21") {
		t.Fatalf("OpenJDK package was not installed: %v", runner.commands)
	}
	record, err := loadInstallationRecord(p, "java")
	if err != nil || record.JavaHome != result.JavaHome {
		t.Fatalf("Java home was not persisted: %+v, %v", record, err)
	}
}

func TestParseJavaHomeRejectsOutsideTermuxPrefix(t *testing.T) {
	prefix := "/data/data/com.termux/files/usr"
	if _, err := parseJavaHome("java.home = /tmp/jdk", prefix); err == nil {
		t.Fatal("outside Java home was accepted")
	}
	if _, err := parseJavaHome("java.home = /data/data/com.termux/files/usr/lib/jvm/java-21-openjdk", prefix); err != nil {
		t.Fatalf("native Java home was rejected: %v", err)
	}
}

func TestInstallJavaRecordsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	p := paths.New(t.TempDir(), "/data/data/com.termux/files/usr")
	result, err := Install(ctx, "java", Options{
		Paths:       p,
		Runner:      canceledRunner{},
		Now:         time.Now,
		StorageFree: func(string) (int64, error) { return StorageWarningBytes + 1, nil },
	})
	if err == nil || result.State != "failed" {
		t.Fatalf("canceled Java installation = %+v, %v", result, err)
	}
	record, loadErr := loadInstallationRecord(p, "java")
	if loadErr != nil || record.State != "failed" {
		t.Fatalf("canceled Java installation state = %+v, %v", record, loadErr)
	}
}

func TestInstallJavaHonorsStorageBlock(t *testing.T) {
	p := paths.New(t.TempDir(), "/data/data/com.termux/files/usr")
	result, err := Install(context.Background(), "java", Options{
		Paths:       p,
		Now:         time.Now,
		StorageFree: func(string) (int64, error) { return StorageBlockBytes - 1, nil },
	})
	if err == nil || result.State != "blocked" || !result.StorageBlocked {
		t.Fatalf("blocked Java installation = %+v, %v", result, err)
	}
}

func TestUninstallJavaRemovesManagedPackage(t *testing.T) {
	p := paths.New(t.TempDir(), "/data/data/com.termux/files/usr")
	if err := os.MkdirAll(p.InstallationsDir(), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := saveRecord(p.InstallationsDir(), InstallationRecord{Name: "java", Package: "openjdk-21", Packages: []string{"openjdk-21"}, Strategy: "pkg", State: "installed"}); err != nil {
		t.Fatal(err)
	}
	runner := &nativeRunner{}
	result, err := Uninstall(context.Background(), "java", Options{Paths: p, Runner: runner, Now: time.Now})
	if err != nil || result.State != "uninstalled" || !containsCommand(runner.commands, "pkg uninstall -y openjdk-21") {
		t.Fatalf("Java uninstall = %+v, %v, commands=%v", result, err, runner.commands)
	}
}

func TestUninstallJavaRefusesRequiredProfile(t *testing.T) {
	p := paths.New(t.TempDir(), "/data/data/com.termux/files/usr")
	if err := os.MkdirAll(p.InstallationsDir(), 0o700); err != nil {
		t.Fatal(err)
	}
	for _, record := range []InstallationRecord{
		{Name: "java", Package: "openjdk-21", Packages: []string{"openjdk-21"}, Strategy: "pkg", State: "installed"},
		{Name: "maven", Package: "maven", Packages: []string{"maven"}, Dependencies: []string{"java"}, Strategy: "pkg", State: "installed"},
	} {
		if err := saveRecord(p.InstallationsDir(), record); err != nil {
			t.Fatal(err)
		}
	}
	result, err := Uninstall(context.Background(), "java", Options{Paths: p, Runner: &nativeRunner{}, Now: time.Now})
	if err == nil || !sameStrings(result.Conflicts, []string{"maven"}) {
		t.Fatalf("Java uninstall did not report Maven dependency: %+v, %v", result, err)
	}
}

func TestNodeProfileRequiresNPM(t *testing.T) {
	profile, ok := Resolve("node")
	if !ok {
		t.Fatal("node profile is missing")
	}
	want := []ExecutableSpec{{Name: "node", VersionArg: []string{"--version"}}, {Name: "npm", VersionArg: []string{"--version"}}}
	if strings.TrimSpace(profile.Package) != "nodejs" || !sameExecutableSpecs(profileExecutables(profile), want) {
		t.Fatalf("unexpected node profile: %+v", profile)
	}
}

func TestSharedNativePackageReleasesProfileBeforeRemovingLastOwner(t *testing.T) {
	p := paths.New(t.TempDir(), "")
	if err := os.MkdirAll(p.InstallationsDir(), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := saveRecord(p.InstallationsDir(), InstallationRecord{Name: "c", Packages: []string{"clang"}, Strategy: "pkg", State: "installed"}); err != nil {
		t.Fatal(err)
	}
	if err := saveRecord(p.InstallationsDir(), InstallationRecord{Name: "cpp", Packages: []string{"clang"}, Strategy: "pkg", State: "installed"}); err != nil {
		t.Fatal(err)
	}
	runner := &nativeRunner{}
	first, err := Uninstall(context.Background(), "c", Options{Paths: p, Runner: runner, Now: time.Now})
	if err != nil {
		t.Fatal(err)
	}
	if !first.Changed || first.State != "uninstalled" || !sameStrings(first.PreservedPackages, []string{"clang"}) || containsCommand(runner.commands, "pkg uninstall -y clang") {
		t.Fatalf("unexpected shared-package removal: %+v, commands=%v", first, runner.commands)
	}
	c, err := loadInstallationRecord(p, "c")
	if err != nil || c.State != "uninstalled" {
		t.Fatalf("c ownership was not released: %+v, %v", c, err)
	}

	second, err := Uninstall(context.Background(), "cpp", Options{Paths: p, Runner: runner, Now: time.Now})
	if err != nil {
		t.Fatal(err)
	}
	if !second.Changed || len(second.PreservedPackages) != 0 || !containsCommand(runner.commands, "pkg uninstall -y clang") {
		t.Fatalf("last owner did not remove clang: %+v, commands=%v", second, runner.commands)
	}
}

func containsCommand(commands []string, want string) bool {
	for _, command := range commands {
		if command == want {
			return true
		}
	}
	return false
}

func sameExecutableSpecs(got, want []ExecutableSpec) bool {
	if len(got) != len(want) {
		return false
	}
	for index := range got {
		if got[index].Name != want[index].Name || !sameStrings(got[index].VersionArg, want[index].VersionArg) {
			return false
		}
	}
	return true
}

func sameStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for index := range got {
		if got[index] != want[index] {
			return false
		}
	}
	return true
}
