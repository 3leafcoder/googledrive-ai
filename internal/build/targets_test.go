package build

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestAssetName(t *testing.T) {
	cases := []struct {
		goos, goarch, want string
	}{
		{"darwin", "amd64", "googledrive-mcp-darwin-amd64"},
		{"darwin", "arm64", "googledrive-mcp-darwin-arm64"},
		{"linux", "amd64", "googledrive-mcp-linux-amd64"},
		{"linux", "arm64", "googledrive-mcp-linux-arm64"},
		{"windows", "amd64", "googledrive-mcp-windows-amd64.exe"},
		{"windows", "arm64", "googledrive-mcp-windows-arm64.exe"},
	}
	for _, c := range cases {
		if got := AssetName(c.goos, c.goarch); got != c.want {
			t.Errorf("AssetName(%q, %q) = %q, want %q", c.goos, c.goarch, got, c.want)
		}
	}
}

// The Makefile PLATFORMS list and SupportedTargets must agree — they are the
// same asset-name scheme shared by build, install, and self-update. This guards
// against the two drifting apart.
func TestMakefilePlatformsMatchSupportedTargets(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "Makefile"))
	if err != nil {
		t.Fatalf("read Makefile: %v", err)
	}

	block := regexp.MustCompile(`(?s)PLATFORMS :=(.*?)\n\n`).FindSubmatch(data)
	if block == nil {
		t.Fatal("could not find PLATFORMS block in Makefile")
	}
	found := map[string]bool{}
	for _, f := range strings.Fields(string(block[1])) {
		if f == "\\" {
			continue
		}
		found[f] = true
	}

	want := map[string]bool{}
	for _, tgt := range SupportedTargets {
		want[tgt.OS+"/"+tgt.Arch] = true
	}

	if len(found) != len(want) {
		t.Fatalf("Makefile has %d platforms, SupportedTargets has %d: %v vs %v", len(found), len(want), found, want)
	}
	for p := range want {
		if !found[p] {
			t.Errorf("SupportedTargets platform %q missing from Makefile PLATFORMS", p)
		}
	}
}
