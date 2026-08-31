package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestConfigDir(t *testing.T) {
	dir, err := ConfigDir()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasSuffix(dir, filepath.Join(".config", "googledrive-ai")) {
		t.Errorf("ConfigDir() = %q, want suffix %q", dir, filepath.Join(".config", "googledrive-ai"))
	}
}

func TestScopes(t *testing.T) {
	want := map[string]bool{
		"https://www.googleapis.com/auth/drive":        true,
		"https://www.googleapis.com/auth/spreadsheets": true,
		"https://www.googleapis.com/auth/documents":    true,
	}
	if len(Scopes) != len(want) {
		t.Fatalf("got %d scopes, want %d", len(Scopes), len(want))
	}
	for _, s := range Scopes {
		if !want[s] {
			t.Errorf("unexpected scope %q", s)
		}
	}
}

func TestLoadCredentials_MissingFile(t *testing.T) {
	// Override config dir to a temp directory with no credentials.json.
	orig := configDir
	// Can't override const, so we test via LoadCredentials directly.
	// The real credentials.json likely doesn't exist in CI, so this tests the error path.
	_, err := LoadCredentials()
	if err == nil {
		t.Skip("credentials.json exists; skipping missing-file test")
	}
	_ = orig
	if !strings.Contains(err.Error(), "Setup steps") {
		t.Errorf("expected setup instructions in error, got: %v", err)
	}
}

func TestTokenRoundTrip(t *testing.T) {
	// Use a temp dir for token storage.
	tmpDir := t.TempDir()
	tp := filepath.Join(tmpDir, "token.json")

	tok := &oauth2.Token{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "Bearer",
	}

	// Write token.
	data, err := json.MarshalIndent(tok, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tp, data, 0600); err != nil {
		t.Fatal(err)
	}

	// Read it back.
	readData, err := os.ReadFile(tp)
	if err != nil {
		t.Fatal(err)
	}
	var loaded oauth2.Token
	if err := json.Unmarshal(readData, &loaded); err != nil {
		t.Fatal(err)
	}
	if loaded.AccessToken != tok.AccessToken {
		t.Errorf("AccessToken = %q, want %q", loaded.AccessToken, tok.AccessToken)
	}
	if loaded.RefreshToken != tok.RefreshToken {
		t.Errorf("RefreshToken = %q, want %q", loaded.RefreshToken, tok.RefreshToken)
	}
}

// fakeTokenServer returns an httptest.Server that acts as an OAuth2 token endpoint.
func fakeTokenServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token":  "fake-access-token",
			"refresh_token": "fake-refresh-token",
			"token_type":    "Bearer",
			"expires_in":    3600,
		})
	}))
}

func testOAuthConfig(tokenURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/auth",
			TokenURL: tokenURL,
		},
		Scopes: []string{"https://www.googleapis.com/auth/drive"},
	}
}

func TestCallbackPage_DisplaysCode(t *testing.T) {
	codeCh := make(chan string, 1)
	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		select {
		case codeCh <- code:
		default:
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, callbackPage, code)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/callback?code=test-auth-code-123")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if !strings.Contains(string(body), "test-auth-code-123") {
		t.Error("callback HTML does not contain the authorization code")
	}
	if !strings.Contains(string(body), "copy the code below") {
		t.Error("callback HTML missing copy instructions")
	}
}

func TestBrowserAuthFlow_StdinPaste(t *testing.T) {
	ts := fakeTokenServer(t)
	defer ts.Close()

	// Provide code via stdin reader.
	old := stdinReader
	stdinReader = strings.NewReader("fake-auth-code\n")
	defer func() { stdinReader = old }()

	config := testOAuthConfig(ts.URL)
	tok, err := browserAuthFlow(context.Background(), config)
	if err != nil {
		t.Fatalf("browserAuthFlow failed: %v", err)
	}
	if tok.AccessToken != "fake-access-token" {
		t.Errorf("AccessToken = %q, want %q", tok.AccessToken, "fake-access-token")
	}
}

func TestBrowserAuthFlow_CallbackWins(t *testing.T) {
	ts := fakeTokenServer(t)
	defer ts.Close()

	// Stdin never provides input (blocks forever) — callback should win.
	old := stdinReader
	stdinReader = strings.NewReader("") // EOF immediately, no code
	defer func() { stdinReader = old }()

	config := testOAuthConfig(ts.URL)

	// We need to hit the callback ourselves since no real browser opens.
	// browserAuthFlow picks a random port; we'll call it after it starts.
	// Use a goroutine to run browserAuthFlow, and find the port from redirectURL.

	type result struct {
		tok *oauth2.Token
		err error
	}
	ch := make(chan result, 1)

	// We need to capture the redirect URL to know the callback port.
	// browserAuthFlow sets config.RedirectURL, so we can read it after it starts.
	// But the function blocks. Instead, we start it in a goroutine, wait briefly,
	// then hit the callback.
	go func() {
		tok, err := browserAuthFlow(context.Background(), config)
		ch <- result{tok, err}
	}()

	// Wait for the redirect URL to be set (browserAuthFlow sets it before blocking).
	var callbackURL string
	for range 50 {
		time.Sleep(10 * time.Millisecond)
		if config.RedirectURL != "" {
			callbackURL = config.RedirectURL
			break
		}
	}
	if callbackURL == "" {
		t.Fatal("redirect URL never set")
	}

	resp, err := http.Get(callbackURL + "?code=callback-code-456")
	if err != nil {
		t.Fatalf("hitting callback: %v", err)
	}
	resp.Body.Close()

	r := <-ch
	if r.err != nil {
		t.Fatalf("browserAuthFlow failed: %v", r.err)
	}
	if r.tok.AccessToken != "fake-access-token" {
		t.Errorf("AccessToken = %q, want %q", r.tok.AccessToken, "fake-access-token")
	}
}

func TestBrowserAuthFlow_ContextCancel(t *testing.T) {
	ts := fakeTokenServer(t)
	defer ts.Close()

	// Stdin blocks (no input) and no callback — context cancel should exit cleanly.
	old := stdinReader
	r, w := io.Pipe()
	stdinReader = r
	defer func() {
		stdinReader = old
		w.Close()
	}()

	config := testOAuthConfig(ts.URL)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		_, err := browserAuthFlow(ctx, config)
		done <- err
	}()

	cancel()
	err := <-done
	if err == nil || !strings.Contains(err.Error(), "context canceled") {
		t.Errorf("expected context canceled error, got: %v", err)
	}
}
