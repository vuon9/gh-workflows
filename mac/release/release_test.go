package release

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestResolveSigningIdentityUsesExplicitIdentity(t *testing.T) {
	t.Parallel()

	runner := &recordingRunner{}
	identity, err := ResolveSigningIdentity(runner, "Developer ID Application: Example (TEAMID1234)", "TEAMID1234")
	if err != nil {
		t.Fatal(err)
	}
	if identity != "Developer ID Application: Example (TEAMID1234)" {
		t.Fatalf("identity = %q", identity)
	}
	if len(runner.commands) != 0 {
		t.Fatalf("runner commands = %v, want none", runner.commands)
	}
}

func TestResolveSigningIdentityFindsTeamIdentity(t *testing.T) {
	t.Parallel()

	runner := &recordingRunner{output: `  1) ABC "Developer ID Application: Other (OTHER12345)"
  2) DEF "Developer ID Application: Example (TEAMID1234)"
     2 valid identities found`}
	identity, err := ResolveSigningIdentity(runner, "", "TEAMID1234")
	if err != nil {
		t.Fatal(err)
	}
	if identity != "Developer ID Application: Example (TEAMID1234)" {
		t.Fatalf("identity = %q", identity)
	}
}

func TestInstallCertificateWritesCertAndRunsKeychainCommands(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	runner := &recordingRunner{}
	cfg := CertificateConfig{
		TempDir:          dir,
		CertificateBase64: "Y2VydA==",
		CertificatePass:   "cert-pass",
		KeychainPassword:  "keychain-pass",
	}
	if err := InstallCertificate(runner, cfg); err != nil {
		t.Fatal(err)
	}

	certPath := filepath.Join(dir, "developer-id-application.p12")
	if got, err := os.ReadFile(certPath); err != nil || string(got) != "cert" {
		t.Fatalf("cert file = %q, %v; want cert", got, err)
	}
	wantNames := []string{"security", "security", "security", "security", "security", "security"}
	if got := runner.commandNames(); !reflect.DeepEqual(got, wantNames) {
		t.Fatalf("command names = %v, want %v", got, wantNames)
	}
}

func TestCreateSignedDMGPlansCreateDmgAndVerificationCommands(t *testing.T) {
	t.Parallel()

	runner := &recordingRunner{}
	cfg := DMGConfig{
		AppPath:         "bin/Example.app",
		AppName:         "Example",
		DMGName:         "Example.dmg",
		SigningIdentity: "Developer ID Application: Example (TEAMID1234)",
	}
	if err := CreateSignedDMG(runner, cfg); err != nil {
		t.Fatal(err)
	}

	wantNames := []string{"create-dmg", "codesign", "codesign", "hdiutil"}
	if got := runner.commandNames(); !reflect.DeepEqual(got, wantNames) {
		t.Fatalf("command names = %v, want %v", got, wantNames)
	}
	if runner.commands[0].Dir != "" {
		t.Fatalf("create-dmg dir = %q, want empty", runner.commands[0].Dir)
	}
}

type recordingRunner struct {
	output   string
	commands []Command
}

func (r *recordingRunner) Run(command Command) error {
	r.commands = append(r.commands, command)
	return nil
}

func (r *recordingRunner) Output(command Command) (string, error) {
	r.commands = append(r.commands, command)
	return r.output, nil
}

func (r *recordingRunner) commandNames() []string {
	names := make([]string, len(r.commands))
	for i, command := range r.commands {
		names[i] = command.Name
	}
	return names
}
