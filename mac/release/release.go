package release

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/vuon9/gh-workflows/internal/archive"
)

type Command struct {
	Name string
	Args []string
	Dir  string
	Env  []string
}

type Runner interface {
	Run(Command) error
	Output(Command) (string, error)
}

type ExecRunner struct{}

func (ExecRunner) Run(command Command) error {
	cmd := exec.Command(command.Name, command.Args...)
	cmd.Dir = command.Dir
	cmd.Env = append(os.Environ(), command.Env...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func (ExecRunner) Output(command Command) (string, error) {
	cmd := exec.Command(command.Name, command.Args...)
	cmd.Dir = command.Dir
	cmd.Env = append(os.Environ(), command.Env...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

type CertificateConfig struct {
	TempDir           string
	CertificateBase64 string
	CertificatePass   string
	KeychainPassword  string
}

type APIKeyConfig struct {
	TempDir  string
	Key      string
	KeyID    string
	EnvPath  string
}

type AppConfig struct {
	AppPath         string
	AppName         string
	TeamID          string
	SigningIdentity string
	APIKeyPath      string
	APIKeyID        string
	APIIssuerID     string
}

type DMGConfig struct {
	AppPath         string
	AppName         string
	DMGName         string
	TeamID          string
	SigningIdentity string
	APIKeyPath      string
	APIKeyID        string
	APIIssuerID     string
}

func ValidateSecrets(getenv func(string) string, names ...string) error {
	var missing []string
	for _, name := range names {
		if getenv(name) == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required secrets: %s", strings.Join(missing, ", "))
	}
	return nil
}

func InstallCertificate(runner Runner, cfg CertificateConfig) error {
	if cfg.TempDir == "" || cfg.CertificateBase64 == "" || cfg.CertificatePass == "" || cfg.KeychainPassword == "" {
		return errors.New("temp dir, certificate, certificate password, and keychain password are required")
	}
	cert, err := base64.StdEncoding.DecodeString(cfg.CertificateBase64)
	if err != nil {
		return fmt.Errorf("decoding certificate: %w", err)
	}
	certPath := filepath.Join(cfg.TempDir, "developer-id-application.p12")
	keychainPath := filepath.Join(cfg.TempDir, "developer-id-signing.keychain-db")
	if err := os.WriteFile(certPath, cert, 0o600); err != nil {
		return err
	}
	commands := []Command{
		{Name: "security", Args: []string{"create-keychain", "-p", cfg.KeychainPassword, keychainPath}},
		{Name: "security", Args: []string{"set-keychain-settings", "-lut", "21600", keychainPath}},
		{Name: "security", Args: []string{"unlock-keychain", "-p", cfg.KeychainPassword, keychainPath}},
		{Name: "security", Args: []string{"import", certPath, "-P", cfg.CertificatePass, "-A", "-t", "cert", "-f", "pkcs12", "-k", keychainPath}},
		{Name: "security", Args: []string{"set-key-partition-list", "-S", "apple-tool:,apple:,codesign:", "-s", "-k", cfg.KeychainPassword, keychainPath}},
		{Name: "security", Args: []string{"list-keychains", "-d", "user", "-s", keychainPath}},
	}
	for _, command := range commands {
		if err := runner.Run(command); err != nil {
			return err
		}
	}
	return nil
}

func WriteAPIKey(cfg APIKeyConfig) (string, error) {
	if cfg.TempDir == "" || cfg.Key == "" || cfg.KeyID == "" {
		return "", errors.New("temp dir, api key, and api key id are required")
	}
	privateKeysDir := filepath.Join(cfg.TempDir, "private_keys")
	if err := os.MkdirAll(privateKeysDir, 0o700); err != nil {
		return "", err
	}
	keyPath := filepath.Join(privateKeysDir, "AuthKey_"+cfg.KeyID+".p8")
	if err := os.WriteFile(keyPath, []byte(cfg.Key), 0o600); err != nil {
		return "", err
	}
	if cfg.EnvPath != "" {
		f, err := os.OpenFile(cfg.EnvPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
		if err != nil {
			return "", err
		}
		if _, err := fmt.Fprintf(f, "APP_STORE_CONNECT_API_KEY_PATH=%s\n", keyPath); err != nil {
			_ = f.Close()
			return "", err
		}
		if err := f.Close(); err != nil {
			return "", err
		}
	}
	return keyPath, nil
}

func ResolveSigningIdentity(runner Runner, explicit, teamID string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	if teamID == "" {
		return "", errors.New("team id is required when signing identity is not explicit")
	}
	output, err := runner.Output(Command{Name: "security", Args: []string{"find-identity", "-v", "-p", "codesigning"}})
	if err != nil {
		return "", err
	}
	needle := "(" + teamID + ")"
	for _, line := range strings.Split(output, "\n") {
		if !strings.Contains(line, "Developer ID Application:") || !strings.Contains(line, needle) {
			continue
		}
		start := strings.Index(line, "\"Developer ID Application:")
		end := strings.LastIndex(line, "\"")
		if start >= 0 && end > start {
			return line[start+1 : end], nil
		}
	}
	return "", fmt.Errorf("could not find Developer ID Application identity for team %s", teamID)
}

func SignApp(runner Runner, cfg AppConfig) error {
	identity, err := ResolveSigningIdentity(runner, cfg.SigningIdentity, cfg.TeamID)
	if err != nil {
		return err
	}
	if err := runner.Run(Command{Name: "codesign", Args: []string{"--force", "--deep", "--options", "runtime", "--timestamp", "--sign", identity, cfg.AppPath}}); err != nil {
		return err
	}
	return runner.Run(Command{Name: "codesign", Args: []string{"--verify", "--deep", "--strict", "--verbose=2", cfg.AppPath}})
}

func NotarizeApp(runner Runner, cfg AppConfig) error {
	notaryZip := filepath.Join(os.TempDir(), cfg.AppName+"-notary.zip")
	if err := runner.Run(Command{Name: "ditto", Args: []string{"-c", "-k", "--keepParent", cfg.AppPath, notaryZip}}); err != nil {
		return err
	}
	if err := notarySubmit(runner, cfg.APIKeyPath, cfg.APIKeyID, cfg.APIIssuerID, notaryZip); err != nil {
		return err
	}
	if err := runner.Run(Command{Name: "xcrun", Args: []string{"stapler", "staple", cfg.AppPath}}); err != nil {
		return err
	}
	if err := runner.Run(Command{Name: "xcrun", Args: []string{"stapler", "validate", cfg.AppPath}}); err != nil {
		return err
	}
	return runner.Run(Command{Name: "spctl", Args: []string{"--assess", "--type", "execute", "--verbose=4", cfg.AppPath}})
}

func CreateSignedDMG(runner Runner, cfg DMGConfig) error {
	if err := os.MkdirAll("release", 0o755); err != nil {
		return err
	}
	dmgPath := filepath.Join("release", cfg.DMGName)
	if err := runner.Run(Command{Name: "create-dmg", Args: []string{"--volname", cfg.AppName, "--window-pos", "200", "120", "--window-size", "800", "400", "--icon-size", "100", "--app-drop-link", "600", "185", dmgPath, cfg.AppPath}}); err != nil {
		return err
	}
	identity := cfg.SigningIdentity
	if identity == "" {
		resolved, err := ResolveSigningIdentity(runner, "", cfg.TeamID)
		if err != nil {
			return err
		}
		identity = resolved
	}
	if err := runner.Run(Command{Name: "codesign", Args: []string{"--force", "--timestamp", "--sign", identity, dmgPath}}); err != nil {
		return err
	}
	if err := runner.Run(Command{Name: "codesign", Args: []string{"--verify", "--verbose=2", dmgPath}}); err != nil {
		return err
	}
	return runner.Run(Command{Name: "hdiutil", Args: []string{"verify", dmgPath}})
}

func NotarizeDMG(runner Runner, cfg DMGConfig) error {
	dmgPath := filepath.Join("release", cfg.DMGName)
	if err := notarySubmit(runner, cfg.APIKeyPath, cfg.APIKeyID, cfg.APIIssuerID, dmgPath); err != nil {
		return err
	}
	if err := runner.Run(Command{Name: "xcrun", Args: []string{"stapler", "staple", dmgPath}}); err != nil {
		return err
	}
	if err := runner.Run(Command{Name: "xcrun", Args: []string{"stapler", "validate", dmgPath}}); err != nil {
		return err
	}
	return runner.Run(Command{Name: "spctl", Args: []string{"--assess", "--type", "open", "--context", "context:primary-signature", "--verbose=4", dmgPath}})
}

func ExtractArchive(archivePath, destination string) error {
	return archive.ExtractTarGz(archivePath, destination)
}

func ExtractAppArchive(archivePath, destination, appPath string) error {
	if appPath != "" {
		parent := filepath.Dir(filepath.Clean(appPath))
		if parent != "." {
			destination = filepath.Join(destination, parent)
		}
	}
	if err := os.MkdirAll(destination, 0o755); err != nil {
		return err
	}
	return ExtractArchive(archivePath, destination)
}

func notarySubmit(runner Runner, keyPath, keyID, issuerID, artifact string) error {
	return runner.Run(Command{Name: "xcrun", Args: []string{"notarytool", "submit", artifact, "--key", keyPath, "--key-id", keyID, "--issuer", issuerID, "--wait"}})
}
