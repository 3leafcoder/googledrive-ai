package update

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

// roundTripFunc adapts a function into an http.RoundTripper so tests intercept
// the hardcoded GitHub API URL without real network access.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
	}
}

// FetchLatest queries the release API with the GitHub JSON Accept header and no
// Authorization header, and decodes the tag and assets.
func TestFetchLatestUnauthenticated(t *testing.T) {
	var gotAuth, gotAccept, gotURL string
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotAuth = r.Header.Get("Authorization")
		gotAccept = r.Header.Get("Accept")
		gotURL = r.URL.String()
		return jsonResponse(200, `{"tag_name":"v1.2.3","assets":[{"name":"x","browser_download_url":"u"}]}`), nil
	})}

	rel, err := FetchLatest(client)
	if err != nil {
		t.Fatalf("FetchLatest: %v", err)
	}
	if gotAuth != "" {
		t.Errorf("Authorization header sent (%q); must be unauthenticated", gotAuth)
	}
	if gotAccept != "application/vnd.github+json" {
		t.Errorf("Accept = %q, want application/vnd.github+json", gotAccept)
	}
	if gotURL != latestReleaseURL {
		t.Errorf("URL = %q, want %q", gotURL, latestReleaseURL)
	}
	if rel.TagName != "v1.2.3" || len(rel.Assets) != 1 {
		t.Errorf("decoded release = %+v", rel)
	}
}

// A non-200 API response surfaces the status code and body (e.g. the 403 rate
// limit) so the cause is legible.
func TestFetchLatestNon200SurfacesStatusAndBody(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return jsonResponse(403, "API rate limit exceeded"), nil
	})}

	_, err := FetchLatest(client)
	if err == nil {
		t.Fatal("expected error on non-200")
	}
	if !strings.Contains(err.Error(), "403") || !strings.Contains(err.Error(), "rate limit") {
		t.Fatalf("error %q should include status and body", err.Error())
	}
}

// Run prints "Already on latest version" and does NOT download when the latest
// release is not strictly newer than the running version.
func TestRunAlreadyOnLatest(t *testing.T) {
	calls := 0
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return jsonResponse(200, `{"tag_name":"v1.2.3","assets":[]}`), nil
	})}

	var out bytes.Buffer
	if err := Run(client, "1.2.3", &out); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if got := out.String(); got != "Already on latest version (1.2.3)\n" {
		t.Fatalf("output = %q", got)
	}
	if calls != 1 {
		t.Fatalf("made %d HTTP calls, want 1 (API only, no download)", calls)
	}
}

// An empty tag_name yields a "no releases found" error, wrapped as the
// check-for-updates failure.
func TestRunEmptyTagIsNoReleasesFound(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return jsonResponse(200, `{"tag_name":"","assets":[]}`), nil
	})}

	err := Run(client, "1.0.0", io.Discard)
	if err == nil || !strings.Contains(err.Error(), "no releases found") {
		t.Fatalf("error = %v, want 'no releases found'", err)
	}
	if !strings.Contains(err.Error(), "failed to check for updates") {
		t.Fatalf("error %q should be wrapped as 'failed to check for updates'", err)
	}
}

// A network/transport error is wrapped as "failed to check for updates".
func TestRunNetworkErrorWrapped(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return nil, fmt.Errorf("dial tcp: connection refused")
	})}

	err := Run(client, "1.0.0", io.Discard)
	if err == nil || !strings.Contains(err.Error(), "failed to check for updates") {
		t.Fatalf("error = %v, want wrapped 'failed to check for updates'", err)
	}
}

// RunCheck previews an available update: it prints the current and latest
// versions plus the install hint, returns nil (exit 0), and makes exactly one
// HTTP call (the API query) — no asset download.
func TestRunCheckReportsAvailableUpdate(t *testing.T) {
	calls := 0
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return jsonResponse(200, `{"tag_name":"v2.0.0","assets":[{"name":"googledrive-mcp-darwin-arm64","browser_download_url":"u"}]}`), nil
	})}

	var out bytes.Buffer
	if err := RunCheck(client, "1.0.0", &out); err != nil {
		t.Fatalf("RunCheck: %v", err)
	}
	got := out.String()
	for _, want := range []string{"1.0.0", "2.0.0", "googledrive-mcp update"} {
		if !strings.Contains(got, want) {
			t.Errorf("output %q missing %q", got, want)
		}
	}
	if calls != 1 {
		t.Fatalf("made %d HTTP calls, want 1 (API only, no download)", calls)
	}
}

// RunCheck prints "Already on latest version" and returns nil when nothing newer
// exists — no download, exit 0.
func TestRunCheckAlreadyOnLatest(t *testing.T) {
	calls := 0
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return jsonResponse(200, `{"tag_name":"v1.2.3","assets":[{"name":"googledrive-mcp-darwin-arm64","browser_download_url":"u"}]}`), nil
	})}

	var out bytes.Buffer
	if err := RunCheck(client, "1.2.3", &out); err != nil {
		t.Fatalf("RunCheck: %v", err)
	}
	if got := out.String(); got != "Already on latest version (1.2.3)\n" {
		t.Fatalf("output = %q", got)
	}
	if calls != 1 {
		t.Fatalf("made %d HTTP calls, want 1 (API only, no download)", calls)
	}
}

// A "dev" build refuses to update — both Run and RunCheck return the refusal
// error and make ZERO HTTP calls, so the refusal happens before any network
// request.
func TestDevBuildRefusesBeforeNetwork(t *testing.T) {
	const wantErr = "cannot update a development build; install a released version first"
	for _, tc := range []struct {
		name string
		run  func(*http.Client, string, io.Writer) error
	}{
		{"Run", Run},
		{"RunCheck", RunCheck},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				return jsonResponse(200, `{"tag_name":"v9.9.9","assets":[]}`), nil
			})}

			err := tc.run(client, "dev", io.Discard)
			if err == nil || err.Error() != wantErr {
				t.Fatalf("error = %v, want %q", err, wantErr)
			}
			if calls != 0 {
				t.Fatalf("made %d HTTP calls, want 0 (refuse before any network request)", calls)
			}
		})
	}
}

// When the release IS newer, Run proceeds past the version check into asset
// selection (proven here by the no-matching-asset error), rather than stopping
// at "already on latest". The download+replace itself is covered by the asset
// and replace tests.
func TestRunNewerProceedsToAssetSelection(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return jsonResponse(200, `{"tag_name":"v9.9.9","assets":[{"name":"googledrive-mcp-plan9-mips","browser_download_url":"u"}]}`), nil
	})}

	err := Run(client, "1.0.0", io.Discard)
	if err == nil || !strings.Contains(err.Error(), "no binary found") {
		t.Fatalf("error = %v, want asset-selection failure for a newer release", err)
	}
}
