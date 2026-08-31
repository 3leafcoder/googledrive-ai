package auth

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// stdinReader is the io.Reader used for the manual code paste prompt.
// Tests can replace this to avoid blocking on os.Stdin.
var stdinReader io.Reader = os.Stdin

const configDir = "googledrive-ai"

var Scopes = []string{
	"https://www.googleapis.com/auth/drive",
	"https://www.googleapis.com/auth/spreadsheets",
	"https://www.googleapis.com/auth/documents",
}

// ConfigDir returns the path to ~/.config/googledrive-ai/.
func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, ".config", configDir), nil
}

// credentialsPath returns the full path to credentials.json.
func credentialsPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "credentials.json"), nil
}

// TokenPathForProfile returns the full path to token.json for the given profile.
// An empty profile name uses the default location (~/.config/googledrive-ai/token.json).
func TokenPathForProfile(profile string) (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}
	if profile == "" {
		return filepath.Join(dir, "token.json"), nil
	}
	return filepath.Join(dir, "profiles", profile, "token.json"), nil
}

// LoadCredentials reads the OAuth2 client credentials from credentials.json.
// Returns a clear error with setup instructions if the file is missing.
func LoadCredentials() (*oauth2.Config, error) {
	cp, err := credentialsPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(cp)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf(
				"credentials.json not found at %s\n\n"+
					"Setup steps:\n"+
					"  1. Go to https://console.cloud.google.com/\n"+
					"  2. Create a project and enable Drive API, Sheets API, and Docs API\n"+
					"  3. Create OAuth2 credentials (Desktop app type)\n"+
					"  4. Download the credentials JSON file\n"+
					"  5. Save it to: %s",
				cp, cp,
			)
		}
		return nil, fmt.Errorf("reading credentials.json: %w", err)
	}

	config, err := google.ConfigFromJSON(data, Scopes...)
	if err != nil {
		return nil, fmt.Errorf("parsing credentials.json: %w", err)
	}
	return config, nil
}

// loadCachedTokenForProfile attempts to read a cached token for the given profile.
func loadCachedTokenForProfile(profile string) (*oauth2.Token, error) {
	tp, err := TokenPathForProfile(profile)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(tp)
	if err != nil {
		return nil, err
	}
	var tok oauth2.Token
	if err := json.Unmarshal(data, &tok); err != nil {
		return nil, err
	}
	return &tok, nil
}

// saveTokenForProfile writes the token for the given profile.
func saveTokenForProfile(profile string, tok *oauth2.Token) error {
	tp, err := TokenPathForProfile(profile)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(tp), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(tok, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(tp, data, 0600)
}

// DeleteTokenForProfile removes the cached token for the given profile.
func DeleteTokenForProfile(profile string) error {
	tp, err := TokenPathForProfile(profile)
	if err != nil {
		return err
	}
	err = os.Remove(tp)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// GetTokenSource returns an oauth2.TokenSource for the default profile that
// automatically refreshes and caches tokens. On first run, it launches a
// browser-based consent flow.
func GetTokenSource(ctx context.Context, config *oauth2.Config) (oauth2.TokenSource, error) {
	return GetTokenSourceForProfile(ctx, config, "")
}

// GetTokenSourceForProfile returns an oauth2.TokenSource for the given profile.
func GetTokenSourceForProfile(ctx context.Context, config *oauth2.Config, profile string) (oauth2.TokenSource, error) {
	tok, err := loadCachedTokenForProfile(profile)
	if err == nil {
		ts := config.TokenSource(ctx, tok)
		return &cachingTokenSource{ts: ts, profile: profile}, nil
	}

	tok, err = browserAuthFlow(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("authorization flow failed: %w", err)
	}
	if err := saveTokenForProfile(profile, tok); err != nil {
		return nil, fmt.Errorf("saving token: %w", err)
	}
	ts := config.TokenSource(ctx, tok)
	return &cachingTokenSource{ts: ts, profile: profile}, nil
}

// cachingTokenSource wraps a TokenSource and saves refreshed tokens to disk.
type cachingTokenSource struct {
	ts      oauth2.TokenSource
	profile string
}

func (c *cachingTokenSource) Token() (*oauth2.Token, error) {
	tok, err := c.ts.Token()
	if err != nil {
		return nil, err
	}
	// Best-effort save on each token retrieval (handles refresh).
	_ = saveTokenForProfile(c.profile, tok)
	return tok, nil
}

// callbackPage is the HTML returned after the OAuth2 redirect. It shows the
// authorization code on screen so the user can copy-paste it in headless
// environments where the callback cannot reach the CLI.
const callbackPage = `<!DOCTYPE html>
<html><head><title>Authorization</title></head><body>
<h2>Authorization successful!</h2>
<p>If the terminal accepted the code automatically, you can close this window.</p>
<p>Otherwise, copy the code below and paste it into your terminal:</p>
<pre style="background:#f0f0f0;padding:12px;font-size:16px;user-select:all">%s</pre>
</body></html>`

// browserAuthFlow starts a local HTTP server, opens the browser for consent,
// and races the HTTP callback against a stdin paste prompt so that headless
// users (SSH, Docker, WSL, port-blocked) can still authenticate.
func browserAuthFlow(ctx context.Context, config *oauth2.Config) (*oauth2.Token, error) {
	// Start a listener on a random port.
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("starting local server: %w", err)
	}
	defer listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	config.RedirectURL = fmt.Sprintf("http://localhost:%d/callback", port)

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		if code == "" {
			errCh <- fmt.Errorf("no authorization code in callback")
			fmt.Fprintln(w, "Authorization failed. You can close this window.")
			return
		}
		// Send code to the race channel (non-blocking — first writer wins).
		select {
		case codeCh <- code:
		default:
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, callbackPage, code)
	})

	srv := &http.Server{Handler: mux}
	go func() { _ = srv.Serve(listener) }()
	defer srv.Shutdown(ctx)

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Fprintf(os.Stderr, "Opening browser for authorization...\nIf the browser doesn't open, visit:\n%s\n", authURL)
	openBrowser(authURL)

	// Start stdin reader in a goroutine so headless users can paste the code.
	fmt.Fprintf(os.Stderr, "Paste authorization code here: ")
	go func() {
		scanner := bufio.NewScanner(stdinReader)
		if scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				select {
				case codeCh <- line:
				default:
				}
			}
		}
	}()

	select {
	case code := <-codeCh:
		return config.Exchange(ctx, code)
	case err := <-errCh:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	default:
		cmd = exec.Command("cmd", "/c", "start", url)
	}
	_ = cmd.Start()
}

// RunAuthFlow runs the browser OAuth flow for a profile and saves the token.
// If force is true, deletes the existing token first.
// Returns the token source on success.
func RunAuthFlow(ctx context.Context, config *oauth2.Config, profile string, force bool) (oauth2.TokenSource, error) {
	if force {
		if err := DeleteTokenForProfile(profile); err != nil {
			return nil, fmt.Errorf("deleting existing token: %w", err)
		}
	}

	tok, err := browserAuthFlow(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("authorization flow failed: %w", err)
	}
	if err := saveTokenForProfile(profile, tok); err != nil {
		return nil, fmt.Errorf("saving token: %w", err)
	}
	if err := saveProfileEmail(ctx, config, profile, tok); err != nil {
		// Non-fatal: token is saved, email metadata is best-effort.
		fmt.Fprintf(os.Stderr, "warning: could not fetch user email: %v\n", err)
	}
	return config.TokenSource(ctx, tok), nil
}

// saveProfileEmail fetches the user's email and saves it as profile metadata.
func saveProfileEmail(ctx context.Context, config *oauth2.Config, profile string, tok *oauth2.Token) error {
	ts := config.TokenSource(ctx, tok)
	email, err := FetchUserEmail(ctx, ts)
	if err != nil {
		return err
	}
	return saveProfileInfo(profile, email)
}

// FetchUserEmail retrieves the authenticated user's email via the Drive API.
func FetchUserEmail(ctx context.Context, ts oauth2.TokenSource) (string, error) {
	client := oauth2.NewClient(ctx, ts)
	resp, err := client.Get("https://www.googleapis.com/drive/v3/about?fields=user(emailAddress)")
	if err != nil {
		return "", fmt.Errorf("fetching user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("user info request failed: %s", resp.Status)
	}

	var result struct {
		User struct {
			EmailAddress string `json:"emailAddress"`
		} `json:"user"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("parsing user info: %w", err)
	}
	return result.User.EmailAddress, nil
}

// profileInfoPath returns the path to profile.json for the given profile.
func profileInfoPath(profile string) (string, error) {
	tp, err := TokenPathForProfile(profile)
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(tp), "profile.json"), nil
}

type profileInfo struct {
	Email string `json:"email"`
}

func saveProfileInfo(profile string, email string) error {
	pp, err := profileInfoPath(profile)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(pp), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(profileInfo{Email: email}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(pp, data, 0600)
}

func loadProfileEmail(profile string) (string, error) {
	pp, err := profileInfoPath(profile)
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(pp)
	if err != nil {
		return "", err
	}
	var info profileInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return "", err
	}
	return info.Email, nil
}

// ProfileEntry represents a profile with its name and email.
type ProfileEntry struct {
	Name  string
	Email string
}

// ListProfiles returns all configured profiles with their emails.
func ListProfiles() ([]ProfileEntry, error) {
	dir, err := ConfigDir()
	if err != nil {
		return nil, err
	}

	var profiles []ProfileEntry

	// Check default profile.
	if _, err := os.Stat(filepath.Join(dir, "token.json")); err == nil {
		email, _ := loadProfileEmail("")
		profiles = append(profiles, ProfileEntry{Name: "default", Email: email})
	}

	// Scan profiles/ directory.
	profilesDir := filepath.Join(dir, "profiles")
	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return profiles, nil
		}
		return nil, err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		tokenFile := filepath.Join(profilesDir, e.Name(), "token.json")
		if _, err := os.Stat(tokenFile); err != nil {
			continue
		}
		email, _ := loadProfileEmail(e.Name())
		profiles = append(profiles, ProfileEntry{Name: e.Name(), Email: email})
	}

	return profiles, nil
}
