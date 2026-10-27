// Package update replaces the running googledrive-mcp binary in place with a freshly
// downloaded version, atomically, so a self-update never leaves the user with a
// half-written binary.
package update

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

// TargetPath returns the real on-disk location of the running binary. It starts
// from os.Executable and resolves every symlink with filepath.EvalSymlinks, so
// the swap targets the actual file rather than a symlink pointing at it.
func TargetPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("locating executable: %w", err)
	}
	return resolveTarget(exe)
}

// resolveTarget resolves executable through symlinks to its real path.
func resolveTarget(executable string) (string, error) {
	resolved, err := filepath.EvalSymlinks(executable)
	if err != nil {
		return "", fmt.Errorf("resolving symlinks: %w", err)
	}
	return resolved, nil
}

// Apply replaces the running binary with the bytes produced by download and
// prints the before → after version on success. The staging temp file is created
// in the target directory *before* download is called, so a non-writable install
// dir fails fast (with an actionable message) instead of after a wasted
// download. The new bytes are staged next to the target and swapped in with a
// rename, so the on-disk binary is never partially written.
func Apply(download func() (io.Reader, error), oldVersion, newVersion string, out io.Writer) error {
	target, err := TargetPath()
	if err != nil {
		return err
	}
	return replaceBinary(runtime.GOOS, target, download, oldVersion, newVersion, out)
}

// replaceBinary creates a staging temp file in the same directory as target (so
// the final os.Rename is an atomic same-filesystem move), then calls download to
// obtain the new bytes, and swaps them into place. Creating the temp file first
// means a non-writable install dir fails before any bytes are fetched. goos
// selects the platform strategy. On any failure the temp file is removed and
// target is left untouched unless the swap actually completed.
func replaceBinary(goos, target string, download func() (io.Reader, error), oldVersion, newVersion string, out io.Writer) (err error) {
	dir := filepath.Dir(target)
	// Create the temp file up front: this is the fail-fast gate. A root-owned
	// install dir (e.g. a legacy sudo install into /usr/local/bin) fails here,
	// before the download starts, with an actionable message.
	tmp, err := os.CreateTemp(dir, ".googledrive-mcp-update-*")
	if err != nil {
		if os.IsPermission(err) {
			return permissionError(dir)
		}
		return fmt.Errorf("cannot create temp file: %w", err)
	}
	tmpName := tmp.Name()

	// Remove the temp file on every failure path. Cleared once the swap lands,
	// by which point the temp file has already been renamed away.
	cleanup := true
	defer func() {
		if cleanup {
			os.Remove(tmpName)
		}
	}()

	// Start the download only now that the temp file exists — a permission
	// problem has already been ruled out above.
	r, err := download()
	if err != nil {
		tmp.Close()
		return err
	}
	if _, err := io.Copy(tmp, r); err != nil {
		tmp.Close()
		return fmt.Errorf("writing new binary: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing temp file: %w", err)
	}

	// Keep the installed binary executable on Unix. Windows ignores the mode.
	if goos != "windows" {
		if err := os.Chmod(tmpName, 0o755); err != nil {
			return fmt.Errorf("setting permissions: %w", err)
		}
	}

	if goos == "windows" {
		// A running binary cannot be overwritten in place on Windows, so move
		// the old one aside to free the target, then rename the new one in.
		old := target + ".old"
		os.Remove(old)
		if err := os.Rename(target, old); err != nil {
			return fmt.Errorf("moving old binary aside: %w", err)
		}
		if err := os.Rename(tmpName, target); err != nil {
			os.Rename(old, target) // best-effort restore
			return fmt.Errorf("installing new binary: %w", err)
		}
	} else if err := os.Rename(tmpName, target); err != nil {
		return fmt.Errorf("installing new binary: %w", err)
	}

	cleanup = false
	fmt.Fprintf(out, "Updated successfully: %s → %s\n", oldVersion, newVersion)
	return nil
}

// installDocsURL points at the README install section, the convention-aligned
// fix referenced by the permission-error message.
const installDocsURL = "https://github.com/3leafcoder/googledrive-ai#install"

// permissionError explains that the install dir is not writable and offers both
// the immediate workaround (re-run with sudo) and the convention-aligned fix
// (reinstall to a user-owned dir for sudo-free updates). The binary never
// escalates to sudo itself — that decision is left to the user.
func permissionError(dir string) error {
	return fmt.Errorf(
		"cannot write to %s: permission denied.\n"+
			"googledrive-mcp is installed in a directory you don't own, so it can't update itself.\n"+
			"  • Quick fix: re-run with sudo:  sudo googledrive-mcp update\n"+
			"  • Better fix: reinstall to a user-owned directory (e.g. ~/.local/bin) so future\n"+
			"    updates need no sudo. See %s",
		dir, installDocsURL,
	)
}
