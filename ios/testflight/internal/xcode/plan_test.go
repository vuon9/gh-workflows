package xcode

import "testing"

func TestPlanBuildsArchiveAndExportCommands(t *testing.T) {
	plan, err := NewPlan(Config{
		ProjectPath:       "Spotto.xcodeproj",
		Scheme:            "Spotto",
		Configuration:     "Release",
		Destination:       "generic/platform=iOS",
		ArchivePath:       "build/TestFlight/Spotto.xcarchive",
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
		"-project", "Spotto.xcodeproj",
		"-scheme", "Spotto",
		"-configuration", "Release",
		"-destination", "generic/platform=iOS",
		"-archivePath", "build/TestFlight/Spotto.xcarchive",
		"clean",
	}
	assertArgs(t, archive.Args, wantArchive)

	export := plan.ExportCommand()
	wantExport := []string{
		"xcodebuild", "-exportArchive",
		"-archivePath", "build/TestFlight/Spotto.xcarchive",
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

func TestPlanRejectsMissingProjectAndWorkspace(t *testing.T) {
	_, err := NewPlan(Config{Scheme: "App"})
	if err == nil {
		t.Fatal("NewPlan returned nil error for missing project/workspace")
	}
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
