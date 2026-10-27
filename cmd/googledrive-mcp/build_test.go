package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/pariwork/googledrive-ai/internal/build"
)

// TestBuildAllCrossCompiles drives `make build-all` end to end: it must produce
// one correctly-named artifact per supported platform, and each binary must
// report the version injected via ldflags. This validates the naming scheme,
// the ldflags embedding, and that the CGO_ENABLED=0 cross-compile succeeds.
func TestBuildAllCrossCompiles(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping cross-compile build in -short mode")
	}
	if _, err := exec.LookPath("make"); err != nil {
		t.Skip("make not available")
	}

	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}

	const wantVersion = "9.9.9-test"
	cmd := exec.Command("make", "build-all", "VERSION="+wantVersion)
	cmd.Dir = repoRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("make build-all failed: %v\n%s", err, out)
	}

	distDir := filepath.Join(repoRoot, "dist")
	for _, tgt := range build.SupportedTargets {
		asset := build.AssetName(tgt.OS, tgt.Arch)
		if _, err := os.Stat(filepath.Join(distDir, asset)); err != nil {
			t.Errorf("missing expected artifact %s: %v", asset, err)
		}
	}

	// The host-platform binary must run and report the injected version.
	hostBin := filepath.Join(distDir, build.AssetName(runtime.GOOS, runtime.GOARCH))
	out, err := exec.Command(hostBin, "--version").Output()
	if err != nil {
		t.Fatalf("running %s --version: %v", hostBin, err)
	}
	if got := strings.TrimSpace(string(out)); got != wantVersion {
		t.Errorf("--version = %q, want %q", got, wantVersion)
	}
}
