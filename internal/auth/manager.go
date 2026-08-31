package auth

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/oauth2"
)

// ProfileManager manages the active auth profile and supports runtime switching.
// It implements oauth2.TokenSource, delegating to the active profile's token source.
type ProfileManager struct {
	mu     sync.RWMutex
	config *oauth2.Config
	active string // "" means default profile
	ts     oauth2.TokenSource
}

// NewProfileManager creates a ProfileManager with the given config and initial token source.
func NewProfileManager(config *oauth2.Config, profile string, ts oauth2.TokenSource) *ProfileManager {
	return &ProfileManager{
		config: config,
		active: profile,
		ts:     ts,
	}
}

// Token returns a token from the active profile's token source.
func (m *ProfileManager) Token() (*oauth2.Token, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ts.Token()
}

// ActiveProfile returns the active profile name. Returns "default" for the unnamed profile.
func (m *ProfileManager) ActiveProfile() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.active == "" {
		return "default"
	}
	return m.active
}

// Status returns auth status for the active profile.
func (m *ProfileManager) Status(ctx context.Context) (*ProfileStatus, error) {
	m.mu.RLock()
	ts := m.ts
	active := m.active
	m.mu.RUnlock()

	name := active
	if name == "" {
		name = "default"
	}

	tok, err := ts.Token()
	if err != nil {
		return nil, fmt.Errorf("retrieving token: %w", err)
	}

	email, _ := loadProfileEmail(active)

	return &ProfileStatus{
		Profile: name,
		Email:   email,
		Scopes:  Scopes,
		Expiry:  tok.Expiry.Unix(),
	}, nil
}

// ProfileStatus holds status information for the active profile.
type ProfileStatus struct {
	Profile string   `json:"profile"`
	Email   string   `json:"email"`
	Scopes  []string `json:"scopes"`
	Expiry  int64    `json:"expiry"`
}

// ProfileDetail holds listing information for a profile.
type ProfileDetail struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	IsValid bool   `json:"is_valid"`
}

// List returns all profiles with validity status.
func (m *ProfileManager) List(ctx context.Context) ([]ProfileDetail, error) {
	m.mu.RLock()
	config := m.config
	m.mu.RUnlock()

	entries, err := ListProfiles()
	if err != nil {
		return nil, err
	}

	// If no profiles exist at all, return the default profile as not-yet-authenticated.
	if len(entries) == 0 {
		return []ProfileDetail{{Name: "default", Email: "", IsValid: false}}, nil
	}

	details := make([]ProfileDetail, len(entries))
	for i, e := range entries {
		raw := e.Name
		if raw == "default" {
			raw = ""
		}
		valid := isTokenValid(ctx, config, raw)
		details[i] = ProfileDetail{
			Name:    e.Name,
			Email:   e.Email,
			IsValid: valid,
		}
	}
	return details, nil
}

// SwitchProfile swaps the active token source to the given profile.
func (m *ProfileManager) SwitchProfile(ctx context.Context, profile string) error {
	raw := profile
	if profile == "default" {
		raw = ""
	}

	// Load the cached token to verify it exists.
	tok, err := loadCachedTokenForProfile(raw)
	if err != nil {
		return fmt.Errorf("profile %q does not exist or has no saved token", profile)
	}

	m.mu.RLock()
	config := m.config
	m.mu.RUnlock()

	ts := config.TokenSource(ctx, tok)
	// Verify the token is usable (can be refreshed if expired).
	if _, err := ts.Token(); err != nil {
		return fmt.Errorf("profile %q has no valid token: %w", profile, err)
	}

	m.mu.Lock()
	m.active = raw
	m.ts = &cachingTokenSource{ts: ts, profile: raw}
	m.mu.Unlock()

	return nil
}

// isTokenValid checks if a profile's token can produce a valid access token.
func isTokenValid(ctx context.Context, config *oauth2.Config, profile string) bool {
	tok, err := loadCachedTokenForProfile(profile)
	if err != nil {
		return false
	}
	ts := config.TokenSource(ctx, tok)
	_, err = ts.Token()
	return err == nil
}
