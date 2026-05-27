package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunWritesExportDestinationToPlist(t *testing.T) {
	dir := t.TempDir()
	restore := chdir(t, dir)
	defer restore()

	err := run([]string{
		"--project", "MyApp.xcodeproj",
		"--scheme", "MyApp",
		"--team-id", "ABCDE12345",
		"--skip-tests",
		"--dry-run",
		"--export-destination", "export",
	})
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join("build", "TestFlight", "ExportOptions.plist"))
	if err != nil {
		t.Fatalf("reading ExportOptions.plist: %v", err)
	}
	if !strings.Contains(string(content), "<string>export</string>") {
		t.Fatalf("ExportOptions.plist does not contain export destination:\n%s", content)
	}
}

func TestRunCanUploadExistingArchiveWithoutRearchiving(t *testing.T) {
	dir := t.TempDir()
	restore := chdir(t, dir)
	defer restore()

	var stdout bytes.Buffer
	restoreStdout := captureStdout(t, &stdout)

	err := run([]string{
		"--project", "MyApp.xcodeproj",
		"--scheme", "MyApp",
		"--team-id", "ABCDE12345",
		"--skip-tests",
		"--skip-archive",
		"--archive-path", "build/TestFlight/MyApp.xcarchive",
		"--dry-run",
	})
	if err != nil {
		restoreStdout()
		t.Fatalf("run returned error: %v", err)
	}
	restoreStdout()

	output := stdout.String()
	if strings.Contains(output, "\"archive\"") {
		t.Fatalf("dry run unexpectedly included archive command:\n%s", output)
	}
	if !strings.Contains(output, "\"-exportArchive\"") {
		t.Fatalf("dry run did not include export command:\n%s", output)
	}
}

func TestReusableWorkflowSupportsManualSigningSecrets(t *testing.T) {
	workflow := readRepoFile(t, ".github", "workflows", "ios-testflight.yml")
	for _, want := range []string{
		"APPLE_DISTRIBUTION_CERTIFICATE_P12_BASE64:",
		"APPLE_DISTRIBUTION_CERTIFICATE_PASSWORD:",
		"APPLE_PROVISIONING_PROFILE_BASE64:",
		"Install Apple signing assets",
		"security create-keychain",
		"security import",
		"Library/Developer/Xcode/UserData/Provisioning Profiles",
	} {
		if !strings.Contains(workflow, want) {
			t.Fatalf("workflow missing %q", want)
		}
	}
}

func chdir(t *testing.T, dir string) func() {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir(%q): %v", dir, err)
	}
	return func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatalf("restore Chdir(%q): %v", wd, err)
		}
	}
}

func readRepoFile(t *testing.T, parts ...string) string {
	t.Helper()
	path := filepath.Join(append([]string{"..", "..", "..", ".."}, parts...)...)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q): %v", path, err)
	}
	return string(content)
}

func captureStdout(t *testing.T, out *bytes.Buffer) func() {
	t.Helper()
	old := os.Stdout
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = write
	done := make(chan struct{})
	go func() {
		_, _ = out.ReadFrom(read)
		close(done)
	}()
	return func() {
		_ = write.Close()
		<-done
		os.Stdout = old
		_ = read.Close()
	}
}
