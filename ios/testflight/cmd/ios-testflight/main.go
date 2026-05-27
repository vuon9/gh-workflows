package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/vuon9/gh-workflows/ios/testflight/internal/exportoptions"
	"github.com/vuon9/gh-workflows/ios/testflight/internal/xcode"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "ios-testflight: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	flags := flag.NewFlagSet("ios-testflight", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)

	var cfg xcode.Config
	var teamID string
	var method string
	var destination string
	var signingStyle string
	var bundleID string
	var profileName string
	var skipTests bool
	var skipArchive bool
	var skipCertCheck bool
	var dryRun bool

	flags.StringVar(&cfg.ProjectPath, "project", "", "Path to .xcodeproj. Mutually exclusive with --workspace.")
	flags.StringVar(&cfg.WorkspacePath, "workspace", "", "Path to .xcworkspace. Mutually exclusive with --project.")
	flags.StringVar(&cfg.Scheme, "scheme", "", "Xcode scheme to test, archive, and export.")
	flags.StringVar(&cfg.Configuration, "configuration", "Release", "Xcode build configuration.")
	flags.StringVar(&cfg.Destination, "archive-destination", "generic/platform=iOS", "xcodebuild archive destination.")
	flags.StringVar(&cfg.TestDestination, "test-destination", "platform=iOS Simulator,name=iPhone 17", "xcodebuild test destination.")
	flags.StringVar(&cfg.ArchivePath, "archive-path", "", "Archive output path. Defaults to build/TestFlight/<scheme>.xcarchive.")
	flags.StringVar(&cfg.ExportPath, "export-path", "build/TestFlight/export", "Export output path.")
	flags.StringVar(&cfg.ExportOptionsPath, "export-options-path", "build/TestFlight/ExportOptions.plist", "ExportOptions.plist path.")
	flags.StringVar(&cfg.APIKeyPath, "api-key-path", os.Getenv("APP_STORE_CONNECT_API_KEY_PATH"), "App Store Connect API .p8 key path.")
	flags.StringVar(&cfg.APIKeyID, "api-key-id", os.Getenv("APP_STORE_CONNECT_API_KEY_ID"), "App Store Connect API key ID.")
	flags.StringVar(&cfg.APIIssuerID, "api-issuer-id", os.Getenv("APP_STORE_CONNECT_API_ISSUER_ID"), "App Store Connect API issuer ID.")
	flags.StringVar(&teamID, "team-id", os.Getenv("APPLE_TEAM_ID"), "Apple Developer team ID.")
	flags.StringVar(&method, "export-method", "app-store-connect", "Export method for ExportOptions.plist.")
	flags.StringVar(&destination, "export-destination", "upload", "ExportOptions.plist destination: upload or export.")
	flags.StringVar(&signingStyle, "signing-style", "automatic", "Export signing style: automatic or manual.")
	flags.StringVar(&bundleID, "bundle-id", "", "Bundle ID for manual provisioning profile mapping.")
	flags.StringVar(&profileName, "provisioning-profile", "", "Provisioning profile name for manual signing.")
	flags.BoolVar(&cfg.Clean, "clean", false, "Clean archive action before building.")
	flags.BoolVar(&cfg.AllowProvisioning, "allow-provisioning-updates", true, "Pass -allowProvisioningUpdates to xcodebuild archive/export.")
	flags.BoolVar(&skipTests, "skip-tests", false, "Skip simulator tests before archive.")
	flags.BoolVar(&skipArchive, "skip-archive", false, "Skip archive and export an existing --archive-path.")
	flags.BoolVar(&skipCertCheck, "skip-cert-check", false, "Skip local Apple Distribution identity check.")
	flags.BoolVar(&dryRun, "dry-run", false, "Print commands without executing xcodebuild.")

	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %s", strings.Join(flags.Args(), " "))
	}
	if teamID == "" {
		return errors.New("--team-id or APPLE_TEAM_ID is required")
	}
	if !dryRun && (cfg.APIKeyPath == "" || cfg.APIKeyID == "" || cfg.APIIssuerID == "") {
		return errors.New("App Store Connect API credentials are required for non-dry-run uploads")
	}
	if skipTests {
		cfg.TestDestination = ""
	}
	cfg.TeamID = teamID
	cfg.SigningStyle = signingStyle
	cfg.ProvisioningProfile = profileName

	plan, err := xcode.NewPlan(cfg)
	if err != nil {
		return err
	}

	profiles := map[string]string(nil)
	if signingStyle == "manual" {
		if bundleID == "" || profileName == "" {
			return errors.New("manual signing requires --bundle-id and --provisioning-profile")
		}
		profiles = map[string]string{bundleID: profileName}
	}
	exportPlist, err := exportoptions.GeneratePlist(exportoptions.Options{
		Method:               method,
		Destination:          destination,
		TeamID:               teamID,
		SigningStyle:         signingStyle,
		ProvisioningProfiles: profiles,
	})
	if err != nil {
		return err
	}

	if !skipCertCheck && !dryRun {
		if err := requireDistributionIdentity(); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(filepath.Dir(cfg.ExportOptionsPath), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(cfg.ExportOptionsPath, exportPlist, 0o644); err != nil {
		return err
	}

	if testCommand, ok := plan.TestCommand(); ok {
		if err := runCommand(testCommand, dryRun); err != nil {
			return err
		}
	}
	if !skipArchive {
		if err := runCommand(plan.ArchiveCommand(), dryRun); err != nil {
			return err
		}
	}
	if err := runCommand(plan.ExportCommand(), dryRun); err != nil {
		return err
	}
	return nil
}

func requireDistributionIdentity() error {
	cmd := exec.Command("security", "find-identity", "-p", "codesigning", "-v")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("checking code signing identities: %w", err)
	}
	text := string(output)
	if strings.Contains(text, "Apple Distribution:") || strings.Contains(text, "iPhone Distribution:") {
		return nil
	}
	return errors.New("missing Apple Distribution signing identity in Keychain")
}

func runCommand(command xcode.Command, dryRun bool) error {
	if dryRun {
		fmt.Println(shellQuote(command.Args))
		return nil
	}
	cmd := exec.Command(command.Name, command.Args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func shellQuote(args []string) string {
	quoted := make([]string, len(args))
	for i, arg := range args {
		quoted[i] = strconv.Quote(arg)
	}
	return strings.Join(quoted, " ")
}
