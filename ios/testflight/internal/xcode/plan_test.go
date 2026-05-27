package xcode

import "testing"

func TestPlanBuildsArchiveAndExportCommands(t *testing.T) {
	plan, err := NewPlan(Config{
		ProjectPath:       "MyApp.xcodeproj",
		Scheme:            "MyApp",
		Configuration:     "Release",
		Destination:       "generic/platform=iOS",
		ArchivePath:       "build/TestFlight/MyApp.xcarchive",
		ExportPath:        "build/TestFlight/export",
		ExportOptionsPath: "build/TestFlight/ExportOptions.plist",
		Clean:             true,
	})
	if err != nil {
		t.Fatalf("NewPlan returned error: %v", err)
	}

	archive := plan.ArchiveCommand()
	wantArchive := []string{
		"xcodebuild", "archive",
		"-project", "MyApp.xcodeproj",
		"-scheme", "MyApp",
		"-configuration", "Release",
		"-destination", "generic/platform=iOS",
		"-archivePath", "build/TestFlight/MyApp.xcarchive",
		"clean",
	}
	assertArgs(t, archive.Args, wantArchive)

	export := plan.ExportCommand()
	wantExport := []string{
		"xcodebuild", "-exportArchive",
		"-archivePath", "build/TestFlight/MyApp.xcarchive",
		"-exportPath", "build/TestFlight/export",
		"-exportOptionsPlist", "build/TestFlight/ExportOptions.plist",
	}
	assertArgs(t, export.Args, wantExport)
}

func TestPlanBuildsWorkspaceArchiveCommand(t *testing.T) {
	plan, err := NewPlan(Config{
		WorkspacePath:     "App.xcworkspace",
		Scheme:            "App",
		Configuration:     "Release",
		Destination:       "generic/platform=iOS",
		ArchivePath:       "build/TestFlight/App.xcarchive",
		ExportPath:        "build/TestFlight/export",
		ExportOptionsPath: "build/TestFlight/ExportOptions.plist",
	})
	if err != nil {
		t.Fatalf("NewPlan returned error: %v", err)
	}

	archive := plan.ArchiveCommand()
	want := []string{
		"xcodebuild", "archive",
		"-workspace", "App.xcworkspace",
		"-scheme", "App",
		"-configuration", "Release",
		"-destination", "generic/platform=iOS",
		"-archivePath", "build/TestFlight/App.xcarchive",
	}
	assertArgs(t, archive.Args, want)
}

func TestPlanAddsManualSigningSettingsToArchiveCommand(t *testing.T) {
	plan, err := NewPlan(Config{
		ProjectPath:         "MyApp.xcodeproj",
		Scheme:              "MyApp",
		TeamID:              "ABCDE12345",
		SigningStyle:        "manual",
		ProvisioningProfile: "My App Store Profile",
	})
	if err != nil {
		t.Fatalf("NewPlan returned error: %v", err)
	}

	archive := plan.ArchiveCommand()
	for _, want := range []string{
		"CODE_SIGN_STYLE=Manual",
		"DEVELOPMENT_TEAM=ABCDE12345",
		"CODE_SIGN_IDENTITY=Apple Distribution",
		"PROVISIONING_PROFILE_SPECIFIER=My App Store Profile",
	} {
		if !containsArg(archive.Args, want) {
			t.Fatalf("archive command missing %q:\n%v", want, archive.Args)
		}
	}
}

func TestPlanRejectsMissingProjectAndWorkspace(t *testing.T) {
	_, err := NewPlan(Config{Scheme: "App"})
	if err == nil {
		t.Fatal("NewPlan returned nil error for missing project/workspace")
	}
}

func containsArg(args []string, want string) bool {
	for _, arg := range args {
		if arg == want {
			return true
		}
	}
	return false
}

func assertArgs(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("args length mismatch:\ngot  %v\nwant %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("arg %d mismatch:\ngot  %v\nwant %v", i, got, want)
		}
	}
}
