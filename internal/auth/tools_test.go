package auth

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"golang.org/x/oauth2"
)

// fakeTokenSource returns a fixed token.
type fakeTokenSource struct {
	tok *oauth2.Token
	err error
}

func (f *fakeTokenSource) Token() (*oauth2.Token, error) {
	return f.tok, f.err
}

// setupTestConfigDir creates a temp config dir and overrides ConfigDir for the test.
func setupTestConfigDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	// Override HOME so ConfigDir() returns our temp path.
	t.Setenv("HOME", dir)
	return filepath.Join(dir, ".config", configDir)
}

// writeToken writes a token file for a profile in the test config dir.
func writeToken(t *testing.T, configDir, profile string, tok *oauth2.Token) {
	t.Helper()
	var dir string
	if profile == "" {
		dir = configDir
	} else {
		dir = filepath.Join(configDir, "profiles", profile)
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(tok)
	if err := os.WriteFile(filepath.Join(dir, "token.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
}

// writeProfileInfo writes profile.json for a profile.
func writeProfileInfo(t *testing.T, configDir, profile, email string) {
	t.Helper()
	var dir string
	if profile == "" {
		dir = configDir
	} else {
		dir = filepath.Join(configDir, "profiles", profile)
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(profileInfo{Email: email})
	if err := os.WriteFile(filepath.Join(dir, "profile.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestAuthStatus(t *testing.T) {
	configDir := setupTestConfigDir(t)
	writeProfileInfo(t, configDir, "", "user@example.com")

	expiry := time.Now().Add(time.Hour)
	ts := &fakeTokenSource{tok: &oauth2.Token{
		AccessToken: "test-access",
		Expiry:      expiry,
	}}
	pm := NewProfileManager(&oauth2.Config{}, "", ts)

	handler := authStatusHandler(pm)
	result, err := handler(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected tool error: %s", result.Content[0].(mcp.TextContent).Text)
	}

	var status ProfileStatus
	text := result.Content[0].(mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &status); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	if status.Profile != "default" {
		t.Errorf("profile = %q, want %q", status.Profile, "default")
	}
	if status.Email != "user@example.com" {
		t.Errorf("email = %q, want %q", status.Email, "user@example.com")
	}
	if len(status.Scopes) != 3 {
		t.Errorf("scopes count = %d, want 3", len(status.Scopes))
	}
	if status.Expiry != expiry.Unix() {
		t.Errorf("expiry = %d, want %d", status.Expiry, expiry.Unix())
	}
}

func TestAuthListProfiles_NoProfiles(t *testing.T) {
	setupTestConfigDir(t)

	pm := NewProfileManager(&oauth2.Config{}, "", &fakeTokenSource{
		tok: &oauth2.Token{AccessToken: "x"},
	})

	handler := authListProfilesHandler(pm)
	result, err := handler(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var profiles []ProfileDetail
	text := result.Content[0].(mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &profiles); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	if len(profiles) != 1 {
		t.Fatalf("expected 1 profile, got %d", len(profiles))
	}
	if profiles[0].Name != "default" {
		t.Errorf("name = %q, want %q", profiles[0].Name, "default")
	}
	if profiles[0].IsValid {
		t.Error("expected default profile to be invalid on fresh install")
	}
}

func TestAuthListProfiles_WithProfiles(t *testing.T) {
	configDir := setupTestConfigDir(t)

	// Default profile with valid token.
	writeToken(t, configDir, "", &oauth2.Token{
		AccessToken:  "access-default",
		RefreshToken: "refresh-default",
		Expiry:       time.Now().Add(time.Hour),
	})
	writeProfileInfo(t, configDir, "", "default@example.com")

	// Named profile with expired token (no refresh token = invalid).
	writeToken(t, configDir, "work", &oauth2.Token{
		AccessToken: "access-work",
		Expiry:      time.Now().Add(-time.Hour),
	})
	writeProfileInfo(t, configDir, "work", "work@example.com")

	pm := NewProfileManager(&oauth2.Config{
		Endpoint: oauth2.Endpoint{TokenURL: "http://invalid.test/token"},
	}, "", &fakeTokenSource{
		tok: &oauth2.Token{AccessToken: "x"},
	})

	handler := authListProfilesHandler(pm)
	result, err := handler(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var profiles []ProfileDetail
	text := result.Content[0].(mcp.TextContent).Text
	if err := json.Unmarshal([]byte(text), &profiles); err != nil {
		t.Fatalf("failed to parse result: %v", err)
	}

	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(profiles))
	}

	// Default profile should be valid (not expired).
	defaultP := profiles[0]
	if defaultP.Name != "default" {
		t.Errorf("first profile name = %q, want %q", defaultP.Name, "default")
	}
	if defaultP.Email != "default@example.com" {
		t.Errorf("default email = %q, want %q", defaultP.Email, "default@example.com")
	}
	if !defaultP.IsValid {
		t.Error("expected default profile to be valid")
	}

	// Work profile should be marked invalid (expired, no refresh).
	workP := profiles[1]
	if workP.Name != "work" {
		t.Errorf("second profile name = %q, want %q", workP.Name, "work")
	}
	if workP.Email != "work@example.com" {
		t.Errorf("work email = %q, want %q", workP.Email, "work@example.com")
	}
	if workP.IsValid {
		t.Error("expected work profile to be invalid (expired, no refresh token)")
	}
}

func TestAuthSwitchProfile_Success(t *testing.T) {
	configDir := setupTestConfigDir(t)

	// Write a valid token for "work" profile.
	writeToken(t, configDir, "work", &oauth2.Token{
		AccessToken:  "access-work",
		RefreshToken: "refresh-work",
		Expiry:       time.Now().Add(time.Hour),
	})
	writeProfileInfo(t, configDir, "work", "work@example.com")

	pm := NewProfileManager(&oauth2.Config{}, "", &fakeTokenSource{
		tok: &oauth2.Token{AccessToken: "initial"},
	})

	if pm.ActiveProfile() != "default" {
		t.Fatalf("initial profile = %q, want %q", pm.ActiveProfile(), "default")
	}

	handler := authSwitchProfileHandler(pm)
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"profile": "work"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Fatalf("unexpected tool error: %s", result.Content[0].(mcp.TextContent).Text)
	}

	if pm.ActiveProfile() != "work" {
		t.Errorf("active profile = %q, want %q", pm.ActiveProfile(), "work")
	}

	// Verify the token source was swapped.
	tok, err := pm.Token()
	if err != nil {
		t.Fatalf("failed to get token after switch: %v", err)
	}
	if tok.AccessToken != "access-work" {
		t.Errorf("access token = %q, want %q", tok.AccessToken, "access-work")
	}
}

func TestAuthSwitchProfile_NonexistentProfile(t *testing.T) {
	setupTestConfigDir(t)

	pm := NewProfileManager(&oauth2.Config{}, "", &fakeTokenSource{
		tok: &oauth2.Token{AccessToken: "initial"},
	})

	handler := authSwitchProfileHandler(pm)
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"profile": "nonexistent"}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected tool error for nonexistent profile")
	}

	// Profile should remain unchanged.
	if pm.ActiveProfile() != "default" {
		t.Errorf("profile should remain %q after failed switch, got %q", "default", pm.ActiveProfile())
	}
}

func TestAuthSwitchProfile_MissingParam(t *testing.T) {
	pm := NewProfileManager(&oauth2.Config{}, "", &fakeTokenSource{
		tok: &oauth2.Token{AccessToken: "x"},
	})

	handler := authSwitchProfileHandler(pm)
	result, err := handler(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Fatal("expected tool error for missing profile param")
	}
}
