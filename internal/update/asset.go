package update

import (
	"fmt"
	"io"
	"net/http"
	"runtime"

	"github.com/pariwork/googledrive-ai/internal/build"
)

// Asset is one release asset as returned by the GitHub Releases API. Only the
// fields self-update needs are decoded.
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// SelectAsset returns the download URL of the release asset matching the running
// platform. The wanted name is build.CurrentAssetName (googledrive-mcp-<goos>-<goarch>,
// with a .exe suffix on Windows), the exact same scheme the release pipeline and
// install script use, so the three stay in agreement. Matching is by exact name.
// tag names the release for the error message when no asset matches.
func SelectAsset(assets []Asset, tag string) (string, error) {
	want := build.CurrentAssetName()
	for _, a := range assets {
		if a.Name == want {
			return a.BrowserDownloadURL, nil
		}
	}
	return "", fmt.Errorf("no binary found for %s/%s in release %s", runtime.GOOS, runtime.GOARCH, tag)
}

// DownloadAsset returns a download callback (shaped for Apply) that GETs url with
// Accept: application/octet-stream and yields the response body. The body is left
// open for the caller to stream; Apply drains it into the staging file.
func DownloadAsset(client *http.Client, url string) func() (io.Reader, error) {
	return func() (io.Reader, error) {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("building download request: %w", err)
		}
		req.Header.Set("Accept", "application/octet-stream")

		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("downloading %s: %w", url, err)
		}
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
			resp.Body.Close()
			return nil, fmt.Errorf("downloading %s: status %d: %s", url, resp.StatusCode, body)
		}
		return resp.Body, nil
	}
}
