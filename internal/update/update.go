// Package update checks GitHub releases for newer CLI versions.
package update

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

const (
	defaultBaseURL = "https://api.github.com"
	repo           = "usetero/cli"
	checkTimeout   = 5 * time.Second
)

// Result holds version info when an update is available.
type Result struct {
	Current string
	Latest  string
}

// githubRelease is the subset of the GitHub release response we need.
type githubRelease struct {
	TagName string `json:"tag_name"`
}

// Check queries GitHub for the latest release and compares it to currentVersion.
// Returns a non-nil Result if a newer version is available, nil if up-to-date.
// baseURL overrides the GitHub API base URL (for testing); pass "" for default.
func Check(ctx context.Context, currentVersion string, baseURL string) (*Result, error) {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	ctx, cancel := context.WithTimeout(ctx, checkTimeout)
	defer cancel()

	url := fmt.Sprintf("%s/repos/%s/releases/latest", baseURL, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github api: %s", resp.Status)
	}

	var release githubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}

	latest := normalize(release.TagName)
	current := normalize(currentVersion)

	if !semver.IsValid(latest) || !semver.IsValid(current) {
		return nil, fmt.Errorf("invalid version: current=%q latest=%q", current, latest)
	}

	if semver.Compare(latest, current) > 0 {
		return &Result{Current: current, Latest: latest}, nil
	}

	return nil, nil
}

// normalize ensures a version string has a "v" prefix for semver compatibility.
func normalize(v string) string {
	v = strings.TrimSpace(v)
	if !strings.HasPrefix(v, "v") {
		v = "v" + v
	}
	return v
}
