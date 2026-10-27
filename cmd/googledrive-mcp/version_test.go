package main

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestVersionCommand builds the binary with and without release ldflags and
// checks that both invocation forms report the embedded version and that a plain
// build reports the "dev" sentinel. This covers the ldflags embedding plus the
// `--version` and `version` subcommand forms in one integration test.
func TestVersionCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping build-based test in -short mode")
	}

	dir := t.TempDir()

	build := func(t *testing.T, name string, ldflags ...string) string {
		out := filepath.Join(dir, name)
		args := []string{"build", "-o", out}
		args = append(args, ldflags...)
		args = append(args, ".")
		cmd := exec.Command("go", args...)
		if o, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("build failed: %v\n%s", err, o)
		}
		return out
	}

	run := func(t *testing.T, bin, arg string) string {
		out, err := exec.Command(bin, arg).Output()
		if err != nil {
			t.Fatalf("running %s %s: %v", bin, arg, err)
		}
		return strings.TrimSpace(string(out))
	}

	// Released build: ldflags set the version; both forms must report it.
	relBin := build(t, "googledrive-mcp-rel", "-ldflags", "-X main.version=1.2.3")
	for _, form := range []string{"--version", "version"} {
		if got := run(t, relBin, form); got != "1.2.3" {
			t.Errorf("%s = %q, want %q", form, got, "1.2.3")
		}
	}

	// Plain build: version falls back to the "dev" sentinel.
	devBin := build(t, "googledrive-mcp-dev")
	if got := run(t, devBin, "--version"); got != "dev" {
		t.Errorf("plain build --version = %q, want %q", got, "dev")
	}
}
