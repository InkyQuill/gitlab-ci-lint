package update

import (
	"context"
	"fmt"
	"runtime"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/InkyQuill/gitlab-ci-lint/pkg/version"
)

const (
	repoSlug = "InkyQuill/gitlab-ci-lint"
)

// CheckForUpdates checks if a newer version is available on GitHub Releases
// Returns (latestVersion, isAvailable, error)
func CheckForUpdates(ctx context.Context) (string, bool, error) {
	repo := selfupdate.ParseSlug(repoSlug)
	latest, found, err := selfupdate.DetectLatest(ctx, repo)
	if err != nil {
		return "", false, fmt.Errorf("error checking for updates: %w", err)
	}
	if !found {
		return "", false, fmt.Errorf("no release found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	latestVersion := latest.Version()
	if latest.LessOrEqual(version.Version) {
		return latestVersion, false, nil
	}

	return latestVersion, true, nil
}

// Upgrade downloads and installs the latest version
func Upgrade(ctx context.Context, verbose bool) error {
	repo := selfupdate.ParseSlug(repoSlug)
	latest, found, err := selfupdate.DetectLatest(ctx, repo)
	if err != nil {
		return fmt.Errorf("error detecting latest version: %w", err)
	}
	if !found {
		return fmt.Errorf("no release found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	if verbose {
		fmt.Printf("Current version: %s\n", version.Version)
		fmt.Printf("Latest version: %s\n", latest.Version())
	}

	if latest.LessOrEqual(version.Version) {
		fmt.Println("Already up to date!")
		return nil
	}

	// Get executable path
	exe, err := selfupdate.ExecutablePath()
	if err != nil {
		return fmt.Errorf("could not locate executable path: %w", err)
	}

	if verbose {
		fmt.Printf("Updating %s...\n", exe)
	}

	// Download and update with SHA256 verification
	if err := selfupdate.UpdateTo(ctx, latest.AssetURL, latest.AssetName, exe); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	fmt.Printf("Successfully updated to %s\n", latest.Version())
	return nil
}
