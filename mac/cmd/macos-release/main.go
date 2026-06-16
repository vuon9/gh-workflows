package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/vuon9/gh-workflows/mac/release"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "macos-release: %v\n", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return errors.New("subcommand is required")
	}
	runner := release.ExecRunner{}
	switch args[0] {
	case "validate-secrets":
		return release.ValidateSecrets(os.Getenv,
			"APPLE_DEVELOPER_ID_APPLICATION_CERTIFICATE_P12_BASE64",
			"APPLE_DEVELOPER_ID_APPLICATION_CERTIFICATE_PASSWORD",
			"APP_STORE_CONNECT_API_KEY_P8",
			"APP_STORE_CONNECT_API_KEY_ID",
			"APP_STORE_CONNECT_API_ISSUER_ID",
		)
	case "install-certificate":
		return release.InstallCertificate(runner, release.CertificateConfig{
			TempDir:           env("RUNNER_TEMP"),
			CertificateBase64: env("APPLE_DEVELOPER_ID_APPLICATION_CERTIFICATE_P12_BASE64"),
			CertificatePass:   env("APPLE_DEVELOPER_ID_APPLICATION_CERTIFICATE_PASSWORD"),
			KeychainPassword:  env("SIGNING_KEYCHAIN_PASSWORD"),
		})
	case "write-api-key":
		_, err := release.WriteAPIKey(release.APIKeyConfig{
			TempDir: env("RUNNER_TEMP"),
			Key:     env("APP_STORE_CONNECT_API_KEY_P8"),
			KeyID:   env("APP_STORE_CONNECT_API_KEY_ID"),
			EnvPath: env("GITHUB_ENV"),
		})
		return err
	case "extract-app":
		flags := flag.NewFlagSet("extract-app", flag.ContinueOnError)
		archivePath := flags.String("archive-path", "", "App bundle .tar.gz archive path.")
		destination := flags.String("destination", ".", "Extraction destination.")
		if err := flags.Parse(args[1:]); err != nil {
			return err
		}
		return release.ExtractArchive(*archivePath, *destination)
	case "sign-app":
		cfg, err := parseAppConfig(args[1:])
		if err != nil {
			return err
		}
		return release.SignApp(runner, cfg)
	case "notarize-app":
		cfg, err := parseAppConfig(args[1:])
		if err != nil {
			return err
		}
		return release.NotarizeApp(runner, cfg)
	case "create-dmg":
		cfg, err := parseDMGConfig(args[1:])
		if err != nil {
			return err
		}
		return release.CreateSignedDMG(runner, cfg)
	case "notarize-dmg":
		cfg, err := parseDMGConfig(args[1:])
		if err != nil {
			return err
		}
		return release.NotarizeDMG(runner, cfg)
	default:
		return fmt.Errorf("unknown subcommand %q", args[0])
	}
}

func parseAppConfig(args []string) (release.AppConfig, error) {
	flags := flag.NewFlagSet("app", flag.ContinueOnError)
	var cfg release.AppConfig
	flags.StringVar(&cfg.AppPath, "app-path", "", "Path to .app bundle.")
	flags.StringVar(&cfg.AppName, "app-name", "", "App display name.")
	flags.StringVar(&cfg.TeamID, "team-id", "", "Apple Developer team ID.")
	flags.StringVar(&cfg.SigningIdentity, "signing-identity", env("MACOS_CODESIGN_IDENTITY"), "Developer ID Application signing identity.")
	flags.StringVar(&cfg.APIKeyPath, "api-key-path", env("APP_STORE_CONNECT_API_KEY_PATH"), "App Store Connect API key path.")
	flags.StringVar(&cfg.APIKeyID, "api-key-id", env("APP_STORE_CONNECT_API_KEY_ID"), "App Store Connect API key ID.")
	flags.StringVar(&cfg.APIIssuerID, "api-issuer-id", env("APP_STORE_CONNECT_API_ISSUER_ID"), "App Store Connect API issuer ID.")
	return cfg, flags.Parse(args)
}

func parseDMGConfig(args []string) (release.DMGConfig, error) {
	flags := flag.NewFlagSet("dmg", flag.ContinueOnError)
	var cfg release.DMGConfig
	flags.StringVar(&cfg.AppPath, "app-path", "", "Path to .app bundle.")
	flags.StringVar(&cfg.AppName, "app-name", "", "App display name.")
	flags.StringVar(&cfg.DMGName, "dmg-name", "", "DMG file name.")
	flags.StringVar(&cfg.TeamID, "team-id", "", "Apple Developer team ID.")
	flags.StringVar(&cfg.SigningIdentity, "signing-identity", env("MACOS_CODESIGN_IDENTITY"), "Developer ID Application signing identity.")
	flags.StringVar(&cfg.APIKeyPath, "api-key-path", env("APP_STORE_CONNECT_API_KEY_PATH"), "App Store Connect API key path.")
	flags.StringVar(&cfg.APIKeyID, "api-key-id", env("APP_STORE_CONNECT_API_KEY_ID"), "App Store Connect API key ID.")
	flags.StringVar(&cfg.APIIssuerID, "api-issuer-id", env("APP_STORE_CONNECT_API_ISSUER_ID"), "App Store Connect API issuer ID.")
	return cfg, flags.Parse(args)
}

func env(name string) string {
	return os.Getenv(name)
}
