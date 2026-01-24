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
var version = "0.2.0"

func runUpgrade() interface{} {
	fmt.Printf("Current version: %s\n", version)
	fmt.Printf("Checking for updates...\n")

	// Create GitHub source (no auth needed for public repos)
	source, err := selfupdate.NewGitHubSource(selfupdate.GitHubConfig{})
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("failed to create update source: %v", err),
		}
	}

	// Create updater with checksum validation
	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Source:    source,
		Validator: &selfupdate.ChecksumValidator{UniqueFilename: "checksums.txt"},
	})
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("failed to create updater: %v", err),
		}
	}

	// Check for latest release
	latest, found, err := updater.DetectLatest(
		context.Background(),
		selfupdate.NewRepositorySlug(repoOwner, repoName),
	)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("failed to check for updates: %v", err),
		}
	}

	if !found {
		fmt.Println("No releases found")
		return map[string]interface{}{
			"success": true,
			"message": "No releases found",
			"version": version,
		}
	}

	// Compare versions
	if latest.LessOrEqual(version) {
		fmt.Printf("Already up to date (latest: %s)\n", latest.Version())
		return map[string]interface{}{
			"success":        true,
			"message":        "Already up to date",
			"version":        version,
			"latest_version": latest.Version(),
		}
	}

	// Download and install
	fmt.Printf("New version available: %s\n", latest.Version())
	fmt.Printf("Downloading for %s/%s...\n", runtime.GOOS, runtime.GOARCH)

	exe, err := selfupdate.ExecutablePath()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("failed to get executable path: %v", err),
		}
	}

	if err := updater.UpdateTo(context.Background(), latest, exe); err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("failed to update: %v", err),
		}
	}

	fmt.Printf("Successfully upgraded to %s\n", latest.Version())
	return map[string]interface{}{
		"success":          true,
		"message":          "Successfully upgraded",
		"previous_version": version,
		"new_version":      latest.Version(),
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
