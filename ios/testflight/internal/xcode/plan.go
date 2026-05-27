package xcode

import (
	"errors"
	"fmt"
)

type Config struct {
	ProjectPath         string
	WorkspacePath       string
	Scheme              string
	Configuration       string
	Destination         string
	TestDestination     string
	ArchivePath         string
	ExportPath          string
	ExportOptionsPath   string
	APIKeyPath          string
	APIKeyID            string
	APIIssuerID         string
	TeamID              string
	SigningStyle        string
	ProvisioningProfile string
	Clean               bool
	AllowProvisioning   bool
}

type Command struct {
	Name string
	Args []string
}

type Plan struct {
	config Config
}

func NewPlan(config Config) (Plan, error) {
	if config.ProjectPath == "" && config.WorkspacePath == "" {
		return Plan{}, errors.New("project path or workspace path is required")
	}
	if config.ProjectPath != "" && config.WorkspacePath != "" {
		return Plan{}, errors.New("project path and workspace path are mutually exclusive")
	}
	if config.Scheme == "" {
		return Plan{}, errors.New("scheme is required")
	}
	if config.Configuration == "" {
		config.Configuration = "Release"
	}
	if config.Destination == "" {
		config.Destination = "generic/platform=iOS"
	}
	if config.ArchivePath == "" {
		config.ArchivePath = fmt.Sprintf("build/TestFlight/%s.xcarchive", config.Scheme)
	}
	if config.ExportPath == "" {
		config.ExportPath = "build/TestFlight/export"
	}
	if config.ExportOptionsPath == "" {
		config.ExportOptionsPath = "build/TestFlight/ExportOptions.plist"
	}
	return Plan{config: config}, nil
}

func (p Plan) TestCommand() (Command, bool) {
	if p.config.TestDestination == "" {
		return Command{}, false
	}
	args := []string{"xcodebuild", "test"}
	args = appendProjectOrWorkspace(args, p.config)
	args = append(args, "-scheme", p.config.Scheme, "-destination", p.config.TestDestination)
	return Command{Name: args[0], Args: args}, true
}

func (p Plan) ArchiveCommand() Command {
	args := []string{"xcodebuild", "archive"}
	args = appendProjectOrWorkspace(args, p.config)
	args = append(args,
		"-scheme", p.config.Scheme,
		"-configuration", p.config.Configuration,
		"-destination", p.config.Destination,
		"-archivePath", p.config.ArchivePath,
	)
	args = appendArchiveSigning(args, p.config)
	args = appendAuthentication(args, p.config)
	if p.config.Clean {
		args = append(args, "clean")
	}
	return Command{Name: args[0], Args: args}
}

func (p Plan) ExportCommand() Command {
	args := []string{
		"xcodebuild", "-exportArchive",
		"-archivePath", p.config.ArchivePath,
		"-exportPath", p.config.ExportPath,
		"-exportOptionsPlist", p.config.ExportOptionsPath,
	}
	args = appendAuthentication(args, p.config)
	return Command{Name: args[0], Args: args}
}

func appendProjectOrWorkspace(args []string, config Config) []string {
	if config.WorkspacePath != "" {
		return append(args, "-workspace", config.WorkspacePath)
	}
	return append(args, "-project", config.ProjectPath)
}

func appendAuthentication(args []string, config Config) []string {
	if config.AllowProvisioning {
		args = append(args, "-allowProvisioningUpdates")
	}
	if config.APIKeyPath != "" {
		args = append(args, "-authenticationKeyPath", config.APIKeyPath)
	}
	if config.APIKeyID != "" {
		args = append(args, "-authenticationKeyID", config.APIKeyID)
	}
	if config.APIIssuerID != "" {
		args = append(args, "-authenticationKeyIssuerID", config.APIIssuerID)
	}
	return args
}

func appendArchiveSigning(args []string, config Config) []string {
	if config.SigningStyle != "manual" {
		return args
	}
	args = append(args, "CODE_SIGN_STYLE=Manual")
	if config.TeamID != "" {
		args = append(args, "DEVELOPMENT_TEAM="+config.TeamID)
	}
	args = append(args, "CODE_SIGN_IDENTITY=Apple Distribution")
	if config.ProvisioningProfile != "" {
		args = append(args, "PROVISIONING_PROFILE_SPECIFIER="+config.ProvisioningProfile)
	}
	return args
}
