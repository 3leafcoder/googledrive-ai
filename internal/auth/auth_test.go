package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
