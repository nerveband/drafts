package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/creativeprojects/go-selfupdate"
)

// UpdateCache stores the last update check result
type UpdateCache struct {
	LastCheck      time.Time `json:"last_check"`
	LatestVersion  string    `json:"latest_version"`
	UpdateRequired bool      `json:"update_required"`
}

const (
	updateCacheFile     = "update_cache.json"
	updateCheckInterval = 24 * time.Hour
)

// getCacheDir returns the directory for storing cache files
func getCacheDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	cacheDir := filepath.Join(homeDir, ".drafts-cli")
	os.MkdirAll(cacheDir, 0755)
	return cacheDir
}

// loadUpdateCache loads the cached update check result
func loadUpdateCache() (*UpdateCache, error) {
	cacheDir := getCacheDir()
	if cacheDir == "" {
		return nil, fmt.Errorf("cannot determine cache directory")
	}

	data, err := os.ReadFile(filepath.Join(cacheDir, updateCacheFile))
	if err != nil {
		return nil, err
	}

	var cache UpdateCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, err
	}
	return &cache, nil
}

// saveUpdateCache saves the update check result to cache
func saveUpdateCache(cache *UpdateCache) error {
	cacheDir := getCacheDir()
	if cacheDir == "" {
		return fmt.Errorf("cannot determine cache directory")
	}

	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(cacheDir, updateCacheFile), data, 0644)
}

// checkForUpdates checks if a new version is available (with caching)
func checkForUpdates() (hasUpdate bool, latestVersion string, err error) {
	// Try to load from cache first
	cache, err := loadUpdateCache()
	if err == nil && time.Since(cache.LastCheck) < updateCheckInterval {
		return cache.UpdateRequired, cache.LatestVersion, nil
	}

	// Cache expired or doesn't exist, check GitHub
	source, err := selfupdate.NewGitHubSource(selfupdate.GitHubConfig{})
	if err != nil {
		return false, "", err
	}

	updater, err := selfupdate.NewUpdater(selfupdate.Config{
		Source: source,
	})
	if err != nil {
		return false, "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	latest, found, err := updater.DetectLatest(
		ctx,
		selfupdate.NewRepositorySlug(repoOwner, repoName),
	)
	if err != nil {
		return false, "", err
	}

	if !found {
		return false, "", nil
	}

	latestVersion = latest.Version()
	hasUpdate = !latest.LessOrEqual(version)

	// Save to cache
	newCache := &UpdateCache{
		LastCheck:      time.Now(),
		LatestVersion:  latestVersion,
		UpdateRequired: hasUpdate,
	}
	saveUpdateCache(newCache)

	return hasUpdate, latestVersion, nil
}

// checkAndNotifyUpdate checks for updates and prints a notification if available
// This is non-blocking and silent on errors (to not interrupt normal operation)
func checkAndNotifyUpdate() {
	// Don't show notifications in plain output mode if it would interfere
	// with structured output - but we're called after output() so it's fine

	hasUpdate, latestVersion, err := checkForUpdates()
	if err != nil || !hasUpdate {
		return
	}

	// Print notification to stderr so it doesn't interfere with JSON output
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "╭─────────────────────────────────────────────────────╮\n")
	fmt.Fprintf(os.Stderr, "│  🎁 Update available: %s → %s                 │\n", version, latestVersion)
	fmt.Fprintf(os.Stderr, "│     Run 'drafts upgrade' to update                 │\n")
	fmt.Fprintf(os.Stderr, "╰─────────────────────────────────────────────────────╯\n")
}
