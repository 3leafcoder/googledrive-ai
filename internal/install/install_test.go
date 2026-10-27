// Package install holds an integration test for the repo-root install.sh.
// The script itself is POSIX sh; these tests run it under `sh` with stubbed
// curl/sudo and a temp $HOME so `go test ./...` covers its install-dir logic.
package install

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// scriptPath returns the absolute path to install.sh at the repo root.
func scriptPath(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// This test lives in internal/install; the script is two levels up.
	return filepath.Join(wd, "..", "..", "install.sh")
}

// writeStub writes an executable stub script into dir.
func writeStub(t *testing.T, dir, name, body string) {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(body), 0o755); err != nil {
		t.Fatal(err)
	}
}

// runResult captures the outcome of a scripted install.
type runResult struct {
	exitOK   bool
	stdout   string
	stderr   string
	sudoLog  string
	curlURL  string // download URL the stubbed curl received
	homeDir  string
	stubRoot string // real dir holding stub curl/sudo, prepended to PATH
}

// run executes install.sh with a stubbed PATH and the given env overrides.
// It stubs curl (writes a fake versioned binary) and sudo (logs the call to
// $SUDO_LOG, then grants write access under $STUB_ELEVATE_ROOT and execs the
// command — emulating root's ability to write anywhere we own).
//
// pathDirs are prepended to $PATH (after the stub dir), letting a test put the
// install dir on $PATH to exercise the not-on-PATH warning both ways.
func run(t *testing.T, home string, extraEnv map[string]string, pathDirs ...string) runResult {
	t.Helper()
	stubRoot := t.TempDir()
	sudoLog := filepath.Join(stubRoot, "sudo.log")
	curlURLLog := filepath.Join(stubRoot, "curl-url.log")

	writeStub(t, stubRoot, "curl", `#!/bin/sh
out=""
url=""
while [ $# -gt 0 ]; do
	case "$1" in
		-o) out="$2"; shift 2 ;;
		-*) shift ;;
		*) url="$1"; shift ;;
	esac
done
echo "$url" > "$CURL_URL_LOG"
cat > "$out" <<'FAKEBIN'
#!/bin/sh
echo "9.9.9"
FAKEBIN
`)
	writeStub(t, stubRoot, "sudo", `#!/bin/sh
echo "$*" >> "$SUDO_LOG"
[ -n "${STUB_ELEVATE_ROOT:-}" ] && chmod -R u+rwx "$STUB_ELEVATE_ROOT" 2>/dev/null || true
exec "$@"
`)

	sep := string(os.PathListSeparator)
	pathParts := append([]string{stubRoot}, pathDirs...)
	pathValue := strings.Join(pathParts, sep) + sep + os.Getenv("PATH")

	cmd := exec.Command("sh", scriptPath(t))
	env := []string{
		"HOME=" + home,
		"PATH=" + pathValue,
		"SUDO_LOG=" + sudoLog,
		"CURL_URL_LOG=" + curlURLLog,
	}
	for k, v := range extraEnv {
		env = append(env, k+"="+v)
	}
	cmd.Env = env

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	logBytes, _ := os.ReadFile(sudoLog)
	urlBytes, _ := os.ReadFile(curlURLLog)
	return runResult{
		exitOK:   err == nil,
		stdout:   stdout.String(),
		stderr:   stderr.String(),
		sudoLog:  string(logBytes),
		curlURL:  strings.TrimSpace(string(urlBytes)),
		homeDir:  home,
		stubRoot: stubRoot,
	}
}

// TestDefaultInstallDirIsUserOwnedNoSudo covers the default flow: no overrides,
// binary lands in $HOME/.local/bin, and sudo is never invoked.
func TestDefaultInstallDirIsUserOwnedNoSudo(t *testing.T) {
	home := t.TempDir()
	r := run(t, home, nil)

	if !r.exitOK {
		t.Fatalf("expected success, got failure\nstdout: %s\nstderr: %s", r.stdout, r.stderr)
	}
	dest := filepath.Join(home, ".local", "bin", "googledrive-mcp")
	if _, err := os.Stat(dest); err != nil {
		t.Fatalf("binary not installed at %s: %v", dest, err)
	}
	if r.sudoLog != "" {
		t.Fatalf("sudo was invoked in the default flow: %q", r.sudoLog)
	}
	if !strings.Contains(r.stdout, "Installed googledrive-mcp 9.9.9") {
		t.Fatalf("missing version report in output: %s", r.stdout)
	}
}

// TestHomeDirUnwritableErrorsWithoutSudo covers the guarantee that a $HOME path
// that cannot be written errors with an ownership/permission message and never
// escalates to sudo.
func TestHomeDirUnwritableErrorsWithoutSudo(t *testing.T) {
	home := t.TempDir()
	localBin := filepath.Join(home, ".local", "bin")
	if err := os.MkdirAll(localBin, 0o755); err != nil {
		t.Fatal(err)
	}
	// Make the target dir read-only so mv into it must fail.
	if err := os.Chmod(localBin, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(localBin, 0o755) }) // let t.TempDir cleanup succeed

	r := run(t, home, nil)

	if r.exitOK {
		t.Fatalf("expected failure for unwritable $HOME dir, got success\nstdout: %s", r.stdout)
	}
	if r.sudoLog != "" {
		t.Fatalf("sudo must never be invoked for a $HOME path, got: %q", r.sudoLog)
	}
	msg := strings.ToLower(r.stderr)
	if !strings.Contains(msg, "ownership") && !strings.Contains(msg, "permission") && !strings.Contains(msg, "writable") {
		t.Fatalf("error should point at ownership/permissions, got: %s", r.stderr)
	}
}

// TestOverrideOutsideHomeUsesSudoFallback covers GOOGLEDRIVE_MCP_INSTALL_DIR pointing
// outside $HOME at a non-writable dir: the existing sudo fallback is retained.
func TestOverrideOutsideHomeUsesSudoFallback(t *testing.T) {
	home := t.TempDir()
	// An "outside $HOME" root we own but make read-only, so the non-sudo path
	// cannot create the install dir and the script escalates.
	outside := t.TempDir()
	installDir := filepath.Join(outside, "bin")
	if err := os.Chmod(outside, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(outside, 0o755) })

	r := run(t, home, map[string]string{
		"GOOGLEDRIVE_MCP_INSTALL_DIR": installDir,
		"STUB_ELEVATE_ROOT":           outside,
	})

	if !r.exitOK {
		t.Fatalf("expected sudo fallback to succeed, got failure\nstdout: %s\nstderr: %s", r.stdout, r.stderr)
	}
	if !strings.Contains(r.stdout, "requires sudo") {
		t.Fatalf("expected 'requires sudo' message, got: %s", r.stdout)
	}
	if !strings.Contains(r.sudoLog, "mv") {
		t.Fatalf("expected sudo mv to be invoked, sudo log: %q", r.sudoLog)
	}
	if _, err := os.Stat(filepath.Join(installDir, "googledrive-mcp")); err != nil {
		t.Fatalf("binary not installed via sudo fallback: %v", err)
	}
}

// TestDownloadURLResolvesLatestAndPinned covers asset-URL resolution: an unset
// GOOGLEDRIVE_MCP_VERSION downloads from the `latest` path, a set one from the pinned
// release tag. Both request the platform-specific asset name.
func TestDownloadURLResolvesLatestAndPinned(t *testing.T) {
	t.Run("latest", func(t *testing.T) {
		r := run(t, t.TempDir(), nil)
		if !r.exitOK {
			t.Fatalf("expected success: %s / %s", r.stdout, r.stderr)
		}
		if !strings.Contains(r.curlURL, "/releases/latest/download/googledrive-mcp-") {
			t.Fatalf("expected latest download URL, got: %q", r.curlURL)
		}
	})
	t.Run("pinned", func(t *testing.T) {
		r := run(t, t.TempDir(), map[string]string{"GOOGLEDRIVE_MCP_VERSION": "v1.2.3"})
		if !r.exitOK {
			t.Fatalf("expected success: %s / %s", r.stdout, r.stderr)
		}
		if !strings.Contains(r.curlURL, "/releases/download/v1.2.3/googledrive-mcp-") {
			t.Fatalf("expected pinned v1.2.3 download URL, got: %q", r.curlURL)
		}
	})
}

// TestUnsupportedPlatformErrors covers platform detection: an unsupported
// arch/OS aborts before any download with a clear message pointing at the
// releases page (and Windows gets its own pointer).
func TestUnsupportedPlatformErrors(t *testing.T) {
	cases := []struct {
		name, unameS, unameM, wantSubstr string
	}{
		{"arch", "Linux", "riscv64", "unsupported architecture"},
		{"os", "Plan9", "x86_64", "unsupported OS"},
		{"windows", "MINGW64_NT-10.0", "x86_64", "Windows is not supported"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stubRoot := t.TempDir()
			writeStub(t, stubRoot, "uname", `#!/bin/sh
case "$1" in
	-s) echo "`+tc.unameS+`" ;;
	-m) echo "`+tc.unameM+`" ;;
esac
`)
			// A curl stub that fails loudly if reached — detection must abort first.
			writeStub(t, stubRoot, "curl", "#!/bin/sh\nexit 42\n")

			cmd := exec.Command("sh", scriptPath(t))
			cmd.Env = []string{
				"HOME=" + t.TempDir(),
				"PATH=" + stubRoot + string(os.PathListSeparator) + os.Getenv("PATH"),
			}
			var stderr strings.Builder
			cmd.Stderr = &stderr
			if err := cmd.Run(); err == nil {
				t.Fatalf("expected failure for %s/%s", tc.unameS, tc.unameM)
			}
			if !strings.Contains(stderr.String(), tc.wantSubstr) {
				t.Fatalf("stderr missing %q: %s", tc.wantSubstr, stderr.String())
			}
			if !strings.Contains(stderr.String(), "/releases") {
				t.Fatalf("error should point at the releases page: %s", stderr.String())
			}
		})
	}
}

// TestPATHWarning covers the post-install PATH check: when the install dir is
// not a $PATH component the script prints a copy-paste fix to stderr (dir,
// export line, inferred rc file, restart/source reminder) yet still exits 0;
// when it is on $PATH there is no warning; and the match is exact, not a
// substring.
func TestPATHWarning(t *testing.T) {
	t.Run("not on PATH warns to stderr with fix, exits 0", func(t *testing.T) {
		home := t.TempDir()
		r := run(t, home, map[string]string{"SHELL": "/usr/bin/zsh"})
		if !r.exitOK {
			t.Fatalf("install must still succeed (exit 0), got failure\nstderr: %s", r.stderr)
		}
		installDir := filepath.Join(home, ".local", "bin")
		// The post-install version line must use the full binary path so it
		// works before PATH is fixed.
		if !strings.Contains(r.stdout, filepath.Join(installDir, "googledrive-mcp")) {
			t.Fatalf("post-install line should reference the full binary path, stdout: %s", r.stdout)
		}
		w := r.stderr
		if !strings.Contains(w, installDir) {
			t.Fatalf("warning missing the install dir: %s", w)
		}
		if !strings.Contains(w, `export PATH="`+installDir+`:$PATH"`) {
			t.Fatalf("warning missing the exact export line: %s", w)
		}
		if !strings.Contains(w, "~/.zshrc") {
			t.Fatalf("warning should infer ~/.zshrc from a zsh $SHELL: %s", w)
		}
		lw := strings.ToLower(w)
		if !strings.Contains(lw, "restart") && !strings.Contains(lw, "source") {
			t.Fatalf("warning should remind to restart/source the rc file: %s", w)
		}
	})

	t.Run("on PATH: no warning", func(t *testing.T) {
		home := t.TempDir()
		installDir := filepath.Join(home, ".local", "bin")
		r := run(t, home, nil, installDir)
		if !r.exitOK {
			t.Fatalf("expected success: %s", r.stderr)
		}
		if strings.Contains(r.stderr, "not on your PATH") {
			t.Fatalf("no warning expected when the dir is on $PATH, stderr: %s", r.stderr)
		}
	})

	t.Run("exact component match, not substring", func(t *testing.T) {
		home := t.TempDir()
		installDir := filepath.Join(home, ".local", "bin")
		// installDir+"ary" contains installDir as a substring but is a distinct
		// PATH component; it must not suppress the warning.
		r := run(t, home, nil, installDir+"ary")
		if !r.exitOK {
			t.Fatalf("expected success: %s", r.stderr)
		}
		if !strings.Contains(r.stderr, "not on your PATH") {
			t.Fatalf("a substring PATH entry must not count as on-PATH; expected warning, stderr: %s", r.stderr)
		}
	})

	t.Run("bash $SHELL infers the platform rc file", func(t *testing.T) {
		home := t.TempDir()
		r := run(t, home, map[string]string{"SHELL": "/bin/bash"})
		want := "~/.bashrc"
		if runtime.GOOS == "darwin" {
			want = "~/.bash_profile"
		}
		if !strings.Contains(r.stderr, want) {
			t.Fatalf("bash on %s should infer %s, stderr: %s", runtime.GOOS, want, r.stderr)
		}
	})

	t.Run("unknown $SHELL falls back to generic guidance", func(t *testing.T) {
		home := t.TempDir()
		r := run(t, home, map[string]string{"SHELL": "/usr/bin/fish"})
		if !strings.Contains(r.stderr, "your shell config file") {
			t.Fatalf("unknown shell should get generic guidance, stderr: %s", r.stderr)
		}
	})
}

func TestMain(m *testing.M) {
	// Skip on Windows: install.sh is a POSIX shell script.
	if runtime.GOOS == "windows" {
		os.Exit(0)
	}
	os.Exit(m.Run())
}
