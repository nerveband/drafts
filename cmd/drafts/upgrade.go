package main

import (
	"context"
	"fmt"
	"runtime"

	"github.com/creativeprojects/go-selfupdate"
)

// Configure these for the repo
const repoOwner = "nerveband"
const repoName = "drafts-applescript-cli"

// version is set at build time via ldflags
var version = "3.0.2"

type UpgradeResult struct {
	Message         string `json:"message"`
	PreviousVersion string `json:"previousVersion,omitempty"`
	NewVersion      string `json:"newVersion,omitempty"`
	LatestVersion   string `json:"latestVersion,omitempty"`
}

func runUpgrade() interface{} {
	// Create GitHub source (no auth needed for public repos)
	source, err := selfupdate.NewGitHubSource(selfupdate.GitHubConfig{})
	if err != nil {
		outputError("PERMISSION_DENIED", fmt.Sprintf("failed to create update source: %v", err), "Check network access and try again")
	}

	// Create updater with checksum validation
	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Source:    source,
		Validator: &selfupdate.ChecksumValidator{UniqueFilename: "checksums.txt"},
	})
	if err != nil {
		outputError("PERMISSION_DENIED", fmt.Sprintf("failed to create updater: %v", err), "Check local filesystem permissions and try again")
	}

	// Check for latest release
	latest, found, err := updater.DetectLatest(
		context.Background(),
		selfupdate.NewRepositorySlug(repoOwner, repoName),
	)
	if err != nil {
		outputError("PERMISSION_DENIED", fmt.Sprintf("failed to check for updates: %v", err), "Check network access and the configured GitHub repository")
	}

	if !found {
		return UpgradeResult{
			Message:         "No releases found",
			PreviousVersion: version,
		}
	}

	// Compare versions
	if !isNewerVersion(latest.Version(), version) {
		return UpgradeResult{
			Message:         "Already up to date",
			PreviousVersion: version,
			LatestVersion:   latest.Version(),
		}
	}

	exe, err := selfupdate.ExecutablePath()
	if err != nil {
		outputError("PERMISSION_DENIED", fmt.Sprintf("failed to get executable path: %v", err), "Run the installed binary directly from disk and try again")
	}

	if err := updater.UpdateTo(context.Background(), latest, exe); err != nil {
		outputError("PERMISSION_DENIED", fmt.Sprintf("failed to update: %v", err), "Check filesystem permissions for the installed binary and try again")
	}

	return UpgradeResult{
		Message:         fmt.Sprintf("Successfully upgraded for %s/%s", runtime.GOOS, runtime.GOARCH),
		PreviousVersion: version,
		NewVersion:      latest.Version(),
	}
}

func runVersion() interface{} {
	return map[string]interface{}{
		"name":    "drafts",
		"version": version,
		"os":      runtime.GOOS,
		"arch":    runtime.GOARCH,
	}
}
