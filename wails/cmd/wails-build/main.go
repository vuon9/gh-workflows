package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/vuon9/gh-workflows/wails/build"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "wails-build: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	flags := flag.NewFlagSet("wails-build", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)

	var cfg build.Config
	flags.StringVar(&cfg.WorkingDirectory, "working-directory", ".", "Caller repository working directory.")
	flags.StringVar(&cfg.PackageCommand, "package-command", "", "Command that builds a release .app bundle.")
	flags.StringVar(&cfg.AppPath, "app-path", "", "Path to the built .app bundle, relative to working-directory.")
	flags.StringVar(&cfg.AppName, "app-name", "", "Expected CFBundleName.")
	flags.StringVar(&cfg.BundleID, "bundle-id", "", "Expected CFBundleIdentifier.")
	flags.StringVar(&cfg.ArchivePath, "archive-path", "", "Output .tar.gz path for the app bundle.")
	if err := flags.Parse(args); err != nil {
		return err
	}
	return build.Run(cfg)
}
