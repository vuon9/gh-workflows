package exportoptions

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"sort"
)

type Options struct {
	Method                 string
	Destination            string
	TeamID                 string
	SigningStyle           string
	ProvisioningProfiles   map[string]string
	UploadSymbols          bool
	StripSwiftSymbols      bool
	ManageVersionAndBuild  bool
	HasUploadSymbols       bool
	HasStripSwiftSymbols   bool
	HasManageVersionNumber bool
}

func GeneratePlist(options Options) ([]byte, error) {
	if options.Method == "" {
		options.Method = "app-store-connect"
	}
	if options.Destination == "" {
		options.Destination = "upload"
	}
	if options.SigningStyle == "" {
		options.SigningStyle = "automatic"
	}
	if options.TeamID == "" {
		return nil, errors.New("team id is required")
	}
	if options.SigningStyle == "manual" && len(options.ProvisioningProfiles) == 0 {
		return nil, errors.New("manual signing requires at least one provisioning profile")
	}
	if !options.HasUploadSymbols {
		options.UploadSymbols = true
	}
	if !options.HasStripSwiftSymbols {
		options.StripSwiftSymbols = true
	}

	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	buf.WriteString(`<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">` + "\n")
	buf.WriteString(`<plist version="1.0">` + "\n")
	buf.WriteString("<dict>\n")
	writeString(&buf, "destination", options.Destination)
	writeString(&buf, "method", options.Method)
	writeString(&buf, "signingStyle", options.SigningStyle)
	writeString(&buf, "teamID", options.TeamID)
	writeBool(&buf, "uploadSymbols", options.UploadSymbols)
	writeBool(&buf, "stripSwiftSymbols", options.StripSwiftSymbols)
	writeBool(&buf, "manageAppVersionAndBuildNumber", options.ManageVersionAndBuild)
	if len(options.ProvisioningProfiles) > 0 {
		buf.WriteString("  <key>provisioningProfiles</key>\n")
		buf.WriteString("  <dict>\n")
		keys := make([]string, 0, len(options.ProvisioningProfiles))
		for bundleID := range options.ProvisioningProfiles {
			keys = append(keys, bundleID)
		}
		sort.Strings(keys)
		for _, bundleID := range keys {
			writeStringIndented(&buf, bundleID, options.ProvisioningProfiles[bundleID], "    ")
		}
		buf.WriteString("  </dict>\n")
	}
	buf.WriteString("</dict>\n")
	buf.WriteString("</plist>\n")
	return buf.Bytes(), nil
}

func writeString(buf *bytes.Buffer, key, value string) {
	writeStringIndented(buf, key, value, "  ")
}

func writeStringIndented(buf *bytes.Buffer, key, value, indent string) {
	buf.WriteString(fmt.Sprintf("%s<key>%s</key>\n", indent, xmlEscape(key)))
	buf.WriteString(fmt.Sprintf("%s<string>%s</string>\n", indent, xmlEscape(value)))
}

func writeBool(buf *bytes.Buffer, key string, value bool) {
	buf.WriteString(fmt.Sprintf("  <key>%s</key>\n", xmlEscape(key)))
	if value {
		buf.WriteString("  <true/>\n")
		return
	}
	buf.WriteString("  <false/>\n")
}

func xmlEscape(value string) string {
	var buf bytes.Buffer
	_ = xml.EscapeText(&buf, []byte(value))
	return buf.String()
}
