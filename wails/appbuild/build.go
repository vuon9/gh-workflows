package appbuild

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/vuon9/gh-workflows/internal/archive"
)

type Config struct {
	WorkingDirectory string
	PackageCommand   string
	AppPath          string
	AppName          string
	BundleID         string
	ArchivePath      string
}

func Run(cfg Config) error {
	if cfg.WorkingDirectory == "" {
		cfg.WorkingDirectory = "."
	}
	if cfg.PackageCommand == "" {
		return errors.New("package command is required")
	}
	if cfg.AppPath == "" || cfg.AppName == "" || cfg.BundleID == "" || cfg.ArchivePath == "" {
		return errors.New("app path, app name, bundle id, and archive path are required")
	}

	cmd := exec.Command("bash", "-lc", cfg.PackageCommand)
	cmd.Dir = cfg.WorkingDirectory
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("running package command: %w", err)
	}

	fullAppPath := filepath.Join(cfg.WorkingDirectory, cfg.AppPath)
	if err := ValidateApp(fullAppPath, cfg.AppName, cfg.BundleID); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(cfg.ArchivePath), 0o755); err != nil {
		return err
	}
	return archive.CreateTarGz(cfg.ArchivePath, fullAppPath)
}

func ValidateApp(appPath, appName, bundleID string) error {
	info, err := os.Stat(appPath)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", appPath)
	}
	plistPath := filepath.Join(appPath, "Contents", "Info.plist")
	values, err := readInfoPlist(plistPath)
	if err != nil {
		return err
	}
	if got := values["CFBundleName"]; got != appName {
		return fmt.Errorf("expected CFBundleName %q, got %q", appName, got)
	}
	if got := values["CFBundleIdentifier"]; got != bundleID {
		return fmt.Errorf("expected CFBundleIdentifier %q, got %q", bundleID, got)
	}
	return nil
}

func readInfoPlist(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	decoder := xml.NewDecoder(bytes.NewReader(data))
	values := make(map[string]string)
	var lastKey string
	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			return values, nil
		}
		if err != nil {
			return nil, err
		}
		start, ok := token.(xml.StartElement)
		if !ok {
			continue
		}
		switch start.Name.Local {
		case "key":
			var key string
			if err := decoder.DecodeElement(&key, &start); err != nil {
				return nil, err
			}
			lastKey = key
		case "string":
			var value string
			if err := decoder.DecodeElement(&value, &start); err != nil {
				return nil, err
			}
			if lastKey != "" {
				values[lastKey] = value
				lastKey = ""
			}
		}
	}
}
