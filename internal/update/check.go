package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pariwork/googledrive-ai/internal/version"
)

// latestReleaseURL is the public GitHub Releases API endpoint for the latest
// googledrive-mcp release. It is queried unauthenticated — no token, login, or other
// credential — the same zero-auth posture as the install script.
//
// A future move to a private repo must change this download path *and* the
// self-update-from-github-releases assertion together, because auth would then
// be required.
const latestReleaseURL = "https://api.github.com/repos/3leafcoder/googledrive-ai/releases/latest"

// guardDevBuild refuses to self-update an un-versioned local build. A binary
// built with a plain `go build` reports its version as the "dev" sentinel (see
// version.Current); self-updating it would silently overwrite a developer's
// working binary with a release. Both Run and RunCheck call this first, so the
// refusal happens before any network request is made.
func guardDevBuild(currentVersion string) error {
	if currentVersion == "dev" {
		return fmt.Errorf("cannot update a development build; install a released version first")
	}
	return nil
}

// Release is the subset of the GitHub Releases API response self-update needs.
type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

// FetchLatest queries the public GitHub Releases API for the latest release with
// Accept: application/vnd.github+json. No Authorization header is set and no
// token is read from the environment. Network errors and non-200 responses are
// returned with the status code and response body so the cause (including the
// per-IP 60/hour rate-limit 403) stays legible. client is injectable so tests
// run without real network access.
func FetchLatest(client *http.Client) (*Release, error) {
	req, err := http.NewRequest(http.MethodGet, latestReleaseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("querying %s: %w", latestReleaseURL, err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API %s returned status %d: %s", latestReleaseURL, resp.StatusCode, body)
	}

	var rel Release
	if err := json.Unmarshal(body, &rel); err != nil {
		return nil, fmt.Errorf("decoding release response: %w", err)
	}
	return &rel, nil
}

// Run performs the full self-update. It fetches the latest release, reads the
// version from tag_name (leading "v" stripped), and compares it to
// currentVersion. When the latest release is not strictly newer it prints
// "Already on latest version (<version>)" and returns without downloading
// anything. Otherwise it selects the matching platform asset, downloads it, and
// replaces the running binary in place, printing the before → after version.
//
// Any failure of the fetch/version-check phase is wrapped as "failed to check
// for updates" so the CLI surfaces a single, legible cause. client is injectable
// so tests avoid real network access.
func Run(client *http.Client, currentVersion string, out io.Writer) error {
	if err := guardDevBuild(currentVersion); err != nil {
		return err
	}
	rel, err := FetchLatest(client)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	latest := strings.TrimPrefix(rel.TagName, "v")
	if latest == "" {
		return fmt.Errorf("failed to check for updates: no releases found")
	}

	if !version.IsNewer(latest, currentVersion) {
		fmt.Fprintf(out, "Already on latest version (%s)\n", latest)
		return nil
	}

	url, err := SelectAsset(rel.Assets, rel.TagName)
	if err != nil {
		return err
	}
	return Apply(DownloadAsset(client, url), currentVersion, latest, out)
}

// RunCheck previews whether an update is available without touching disk. It
// performs only the Releases API query and version comparison — it never selects,
// downloads, or applies an asset. When a newer release exists it prints the
// current and latest versions plus the hint to run `googledrive-mcp update`; otherwise
// it prints "Already on latest version (<version>)". Both outcomes return nil so
// the CLI exits 0 — an available update is not an error. Only a fetch/parse
// failure returns an error (wrapped as "failed to check for updates"). client is
// injectable so tests avoid real network access.
func RunCheck(client *http.Client, currentVersion string, out io.Writer) error {
	if err := guardDevBuild(currentVersion); err != nil {
		return err
	}
	rel, err := FetchLatest(client)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	latest := strings.TrimPrefix(rel.TagName, "v")
	if latest == "" {
		return fmt.Errorf("failed to check for updates: no releases found")
	}

	if !version.IsNewer(latest, currentVersion) {
		fmt.Fprintf(out, "Already on latest version (%s)\n", latest)
		return nil
	}

	fmt.Fprintf(out, "Update available: %s -> %s\n", currentVersion, latest)
	fmt.Fprintln(out, "Run `googledrive-mcp update` to install.")
	return nil
}
