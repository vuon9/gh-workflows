package archive

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateAndExtractTarGzPreservesBundleFiles(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	appPath := filepath.Join(dir, "Example.app")
	contentsPath := filepath.Join(appPath, "Contents", "MacOS")
	if err := os.MkdirAll(contentsPath, 0o755); err != nil {
		t.Fatal(err)
	}
	executable := filepath.Join(contentsPath, "Example")
	if err := os.WriteFile(executable, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(executable, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("MacOS/Example", filepath.Join(appPath, "Contents", "Current")); err != nil {
		t.Fatal(err)
	}

	archivePath := filepath.Join(dir, "app.tar.gz")
	if err := CreateTarGz(archivePath, appPath); err != nil {
		t.Fatal(err)
	}

	outDir := filepath.Join(dir, "out")
	if err := ExtractTarGz(archivePath, outDir); err != nil {
		t.Fatal(err)
	}

	extractedExecutable := filepath.Join(outDir, "Example.app", "Contents", "MacOS", "Example")
	info, err := os.Stat(extractedExecutable)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o755 {
		t.Fatalf("executable mode = %v, want 0755", got)
	}
	linkTarget, err := os.Readlink(filepath.Join(outDir, "Example.app", "Contents", "Current"))
	if err != nil {
		t.Fatal(err)
	}
	if linkTarget != "MacOS/Example" {
		t.Fatalf("link target = %q, want MacOS/Example", linkTarget)
	}
}
