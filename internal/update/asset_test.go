package update

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"

	"github.com/3leafcoder/googledrive-ai/internal/build"
)

// SelectAsset picks the asset whose name exactly matches the running platform
// and ignores near-misses (other arches, a prefix match), returning its
// browser_download_url.
func TestSelectAssetExactMatch(t *testing.T) {
	want := build.CurrentAssetName()
	assets := []Asset{
		{Name: "googledrive-mcp-plan9-mips", BrowserDownloadURL: "https://example.com/wrong"},
		{Name: want + "-extra", BrowserDownloadURL: "https://example.com/prefix"},
		{Name: want, BrowserDownloadURL: "https://example.com/right"},
	}

	url, err := SelectAsset(assets, "v1.2.3")
	if err != nil {
		t.Fatalf("SelectAsset: %v", err)
	}
	if url != "https://example.com/right" {
		t.Fatalf("got %q, want the exact-match URL", url)
	}
}

// When no asset matches the platform, the error names the goos/goarch and the
// release tag.
func TestSelectAssetNoMatchNamesPlatformAndTag(t *testing.T) {
	_, err := SelectAsset([]Asset{{Name: "googledrive-mcp-plan9-mips"}}, "v9.9.9")
	if err == nil {
		t.Fatal("expected an error when no asset matches")
	}
	wantMsg := fmt.Sprintf("no binary found for %s/%s in release v9.9.9", runtime.GOOS, runtime.GOARCH)
	if err.Error() != wantMsg {
		t.Fatalf("got %q, want %q", err.Error(), wantMsg)
	}
}

// DownloadAsset GETs the URL with Accept: application/octet-stream and streams
// back the response body.
func TestDownloadAssetSendsOctetStreamAccept(t *testing.T) {
	var gotAccept string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAccept = r.Header.Get("Accept")
		fmt.Fprint(w, "binary-bytes")
	}))
	defer srv.Close()

	r, err := DownloadAsset(srv.Client(), srv.URL)()
	if err != nil {
		t.Fatalf("download: %v", err)
	}
	body, _ := io.ReadAll(r)
	if gotAccept != "application/octet-stream" {
		t.Fatalf("Accept header = %q, want application/octet-stream", gotAccept)
	}
	if string(body) != "binary-bytes" {
		t.Fatalf("body = %q, want binary-bytes", body)
	}
}

// A non-200 download surfaces the status and body as an error.
func TestDownloadAssetNon200IsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := DownloadAsset(srv.Client(), srv.URL)()
	if err == nil {
		t.Fatal("expected an error on non-200 response")
	}
	if !strings.Contains(err.Error(), "404") || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("error %q should include the status and body", err.Error())
	}
}
