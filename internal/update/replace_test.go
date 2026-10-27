package update

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// writeBinary creates a fake "installed binary" at path with the given content
// and mode, returning the path.
func writeBinary(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o755); err != nil {
		t.Fatalf("seeding binary: %v", err)
	}
	return p
}

// tempLeftovers returns any staging temp files left behind in dir.
func tempLeftovers(t *testing.T, dir string) []string {
	t.Helper()
	m, err := filepath.Glob(filepath.Join(dir, ".googledrive-mcp-update-*"))
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	return m
}

func TestReplaceBinary_UnixAtomicSwap(t *testing.T) {
	dir := t.TempDir()
	target := writeBinary(t, dir, "googledrive-mcp", "OLD")

	var out bytes.Buffer
	if err := replaceBinary("linux", target, reader("NEW"), "1.2.0", "1.3.0", &out); err != nil {
		t.Fatalf("replaceBinary: %v", err)
	}

	got, _ := os.ReadFile(target)
	if string(got) != "NEW" {
		t.Errorf("target content = %q, want %q", got, "NEW")
	}
	// Installed binary must stay executable.
	info, _ := os.Stat(target)
	if info.Mode().Perm() != 0o755 {
		t.Errorf("target mode = %v, want 0755", info.Mode().Perm())
	}
	// No staging temp file may survive a successful swap.
	if leftover := tempLeftovers(t, dir); len(leftover) != 0 {
		t.Errorf("temp files left behind: %v", leftover)
	}
	if want := "Updated successfully: 1.2.0 → 1.3.0\n"; out.String() != want {
		t.Errorf("output = %q, want %q", out.String(), want)
	}
}

func TestReplaceBinary_WindowsMovesOldAside(t *testing.T) {
	dir := t.TempDir()
	target := writeBinary(t, dir, "googledrive-mcp.exe", "OLD")

	var out bytes.Buffer
	if err := replaceBinary("windows", target, reader("NEW"), "1.2.0", "1.3.0", &out); err != nil {
		t.Fatalf("replaceBinary: %v", err)
	}

	if got, _ := os.ReadFile(target); string(got) != "NEW" {
		t.Errorf("target content = %q, want %q", got, "NEW")
	}
	// The prior binary is preserved as <path>.old rather than deleted.
	if got, err := os.ReadFile(target + ".old"); err != nil || string(got) != "OLD" {
		t.Errorf(".old = %q (err %v), want %q", got, err, "OLD")
	}
	if leftover := tempLeftovers(t, dir); len(leftover) != 0 {
		t.Errorf("temp files left behind: %v", leftover)
	}
}

// reader returns a download thunk that yields content, the common success case.
func reader(content string) func() (io.Reader, error) {
	return func() (io.Reader, error) { return strings.NewReader(content), nil }
}

// errReader fails partway through the copy so we can exercise the failure path.
type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("boom") }

func TestReplaceBinary_FailureLeavesTargetUntouched(t *testing.T) {
	dir := t.TempDir()
	target := writeBinary(t, dir, "googledrive-mcp", "OLD")

	var out bytes.Buffer
	err := replaceBinary("linux", target, func() (io.Reader, error) { return errReader{}, nil }, "1.2.0", "1.3.0", &out)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	// Existing binary is untouched because the swap never happened.
	if got, _ := os.ReadFile(target); string(got) != "OLD" {
		t.Errorf("target content = %q, want %q (untouched)", got, "OLD")
	}
	// Temp file is cleaned up on the failure path.
	if leftover := tempLeftovers(t, dir); len(leftover) != 0 {
		t.Errorf("temp files left behind on failure: %v", leftover)
	}
	if out.Len() != 0 {
		t.Errorf("no success line expected on failure, got %q", out.String())
	}
}

func TestReplaceBinary_PermissionErrorFailsBeforeDownload(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("dir-mode permission bits are not enforced the same way on Windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permission checks")
	}
	dir := t.TempDir()
	target := writeBinary(t, dir, "googledrive-mcp", "OLD")
	// Strip write permission from the install dir so CreateTemp fails.
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0o755) })

	downloaded := false
	var out bytes.Buffer
	err := replaceBinary("linux", target, func() (io.Reader, error) {
		downloaded = true
		return strings.NewReader("NEW"), nil
	}, "1.2.0", "1.3.0", &out)

	if err == nil {
		t.Fatal("expected permission error, got nil")
	}
	if downloaded {
		t.Error("download was started; it must fail before downloading")
	}
	// The message must name the dir and offer both remedies.
	msg := err.Error()
	for _, want := range []string{dir, "sudo googledrive-mcp update", "~/.local/bin", installDocsURL} {
		if !strings.Contains(msg, want) {
			t.Errorf("permission message missing %q\ngot: %s", want, msg)
		}
	}
	// The generic temp-file wording must NOT be used for a permission error.
	if strings.Contains(msg, "cannot create temp file") {
		t.Errorf("permission error should use the actionable message, not the generic one:\n%s", msg)
	}
	if got, _ := os.ReadFile(target); string(got) != "OLD" {
		t.Errorf("target content = %q, want %q (untouched)", got, "OLD")
	}
}

func TestReplaceBinary_NonPermissionTempFailureSurfacesCause(t *testing.T) {
	// A nonexistent install dir is a non-permission CreateTemp failure.
	target := filepath.Join(t.TempDir(), "no-such-dir", "googledrive-mcp")

	downloaded := false
	var out bytes.Buffer
	err := replaceBinary("linux", target, func() (io.Reader, error) {
		downloaded = true
		return strings.NewReader("NEW"), nil
	}, "1.2.0", "1.3.0", &out)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if downloaded {
		t.Error("download must not start when the temp file cannot be created")
	}
	if !strings.HasPrefix(err.Error(), "cannot create temp file: ") {
		t.Errorf("error = %q, want prefix %q", err.Error(), "cannot create temp file: ")
	}
}

func TestResolveTarget_FollowsSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink semantics differ on Windows")
	}
	dir := t.TempDir()
	real := writeBinary(t, dir, "googledrive-mcp-real", "BIN")
	link := filepath.Join(dir, "googledrive-mcp")
	if err := os.Symlink(real, link); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	got, err := resolveTarget(link)
	if err != nil {
		t.Fatalf("resolveTarget: %v", err)
	}
	// EvalSymlinks may canonicalize the temp dir (e.g. /var → /private/var on
	// macOS), so compare against the equally-resolved real path.
	want, _ := filepath.EvalSymlinks(real)
	if got != want {
		t.Errorf("resolved = %q, want %q (real file, not the symlink)", got, want)
	}
}
