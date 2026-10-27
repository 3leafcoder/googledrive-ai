// Package build holds the canonical cross-compile target list and the
// release-asset naming scheme. This is the single source of truth shared by the
// Makefile build-all target, the install script, and self-update, so all three
// agree on which platforms exist and what each binary is called.
package build

import "runtime"

// Target is one OS/architecture combination googledrive-mcp ships a binary for.
type Target struct {
	OS   string
	Arch string
}

// SupportedTargets is the canonical list of platforms googledrive-mcp cross-compiles
// for. The Makefile PLATFORMS list is kept in sync with this via
// TestMakefilePlatformsMatchSupportedTargets.
var SupportedTargets = []Target{
	{OS: "darwin", Arch: "amd64"},
	{OS: "darwin", Arch: "arm64"},
	{OS: "linux", Arch: "amd64"},
	{OS: "linux", Arch: "arm64"},
	{OS: "windows", Arch: "amd64"},
}

// AssetName returns the release-asset filename for a given OS/arch:
// googledrive-mcp-<goos>-<goarch>, with a .exe suffix on Windows.
func AssetName(goos, goarch string) string {
	name := "googledrive-mcp-" + goos + "-" + goarch
	if goos == "windows" {
		name += ".exe"
	}
	return name
}

// CurrentAssetName returns the asset name for the running platform.
func CurrentAssetName() string {
	return AssetName(runtime.GOOS, runtime.GOARCH)
}
