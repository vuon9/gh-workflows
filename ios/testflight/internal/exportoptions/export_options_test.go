package exportoptions

import (
	"strings"
	"testing"
)

func TestGeneratePlistUsesAppStoreConnectDefaults(t *testing.T) {
	xml, err := GeneratePlist(Options{
		Method:       "app-store-connect",
		TeamID:       "ABCDE12345",
		SigningStyle: "automatic",
	})
	if err != nil {
		t.Fatalf("GeneratePlist returned error: %v", err)
	}

	content := string(xml)
	for _, want := range []string{
		"<key>method</key>",
		"<string>app-store-connect</string>",
		"<key>teamID</key>",
		"<string>ABCDE12345</string>",
		"<key>signingStyle</key>",
		"<string>automatic</string>",
		"<key>uploadSymbols</key>",
		"<true/>",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("plist missing %q:\n%s", want, content)
		}
	}
}

func TestGeneratePlistSupportsManualProvisioningProfiles(t *testing.T) {
	xml, err := GeneratePlist(Options{
		Method:       "app-store-connect",
		TeamID:       "ABCDE12345",
		SigningStyle: "manual",
		ProvisioningProfiles: map[string]string{
			"com.example.MyApp": "MyApp App Store",
		},
	})
	if err != nil {
		t.Fatalf("GeneratePlist returned error: %v", err)
	}

	content := string(xml)
	for _, want := range []string{
		"<key>provisioningProfiles</key>",
		"<key>com.example.MyApp</key>",
		"<string>MyApp App Store</string>",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("plist missing %q:\n%s", want, content)
		}
	}
}

func TestGeneratePlistRejectsMissingTeamID(t *testing.T) {
	_, err := GeneratePlist(Options{Method: "app-store-connect", SigningStyle: "automatic"})
	if err == nil {
		t.Fatal("GeneratePlist returned nil error for missing team id")
	}
}
