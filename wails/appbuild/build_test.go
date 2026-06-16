package appbuild

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunBuildCommandValidatesAppAndArchivesIt(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	archivePath := filepath.Join(dir, "wails-app.tar.gz")
	cfg := Config{
		WorkingDirectory: dir,
		PackageCommand:  `mkdir -p bin/Example.app/Contents && printf '%s' '<?xml version="1.0" encoding="UTF-8"?><plist version="1.0"><dict><key>CFBundleName</key><string>Example</string><key>CFBundleIdentifier</key><string>com.example.app</string></dict></plist>' > bin/Example.app/Contents/Info.plist`,
		AppPath:         "bin/Example.app",
		AppName:         "Example",
		BundleID:        "com.example.app",
		ArchivePath:     archivePath,
	}

	if err := Run(cfg); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(archivePath); err != nil {
		t.Fatal(err)
	}
}

func TestRunRejectsMismatchedBundleID(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfg := Config{
		WorkingDirectory: dir,
		PackageCommand:  `mkdir -p bin/Example.app/Contents && printf '%s' '<?xml version="1.0" encoding="UTF-8"?><plist version="1.0"><dict><key>CFBundleName</key><string>Example</string><key>CFBundleIdentifier</key><string>com.example.other</string></dict></plist>' > bin/Example.app/Contents/Info.plist`,
		AppPath:         "bin/Example.app",
		AppName:         "Example",
		BundleID:        "com.example.app",
		ArchivePath:     filepath.Join(dir, "wails-app.tar.gz"),
	}

	err := Run(cfg)
	if err == nil {
		t.Fatal("Run returned nil, want bundle id mismatch error")
	}
	if got := err.Error(); got != `expected CFBundleIdentifier "com.example.app", got "com.example.other"` {
		t.Fatalf("error = %q", got)
	}
}
