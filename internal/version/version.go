// Package version implements semantic, component-wise version comparison for the
// self-update decision. Versions are compared numerically (so 1.10.0 > 1.9.9),
// never lexically, and an update only happens when the remote version is
// strictly newer than the running one.
package version

import (
	"strconv"
	"strings"
)

// Current is the running binary's version string. Package main injects its
// ldflags-embedded value (-X main.version=...) here at startup via SetCurrent so
// that packages needing the version (self-update, user-agent headers) can read
// it without importing main. It defaults to "dev" for an un-versioned local
// build, matching the sentinel other behavior keys off of.
var Current = "dev"

// SetCurrent records the running binary's version string in Current. Package
// main calls it once at startup with its ldflags-injected value.
func SetCurrent(v string) { Current = v }

// Version is a parsed version as its numeric major/minor/patch components.
type Version struct {
	Major, Minor, Patch int
}

// Parse turns a version string into numeric components. It accepts an optional
// leading "v" (both "1.2.3" and "v1.2.3" parse the same) and ignores any
// pre-release/build suffix after the first "-" ("1.2.3-dirty" → 1.2.3).
// Non-numeric or malformed input (e.g. "dev") parses to the zero Version rather
// than panicking, so an un-versioned dev build always compares as oldest. A
// missing component is treated as 0 ("1.2" → 1.2.0).
func Parse(s string) Version {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "v")
	// Drop anything from the first "-" onward (pre-release / build metadata).
	if i := strings.IndexByte(s, '-'); i >= 0 {
		s = s[:i]
	}

	var v Version
	parts := strings.Split(s, ".")
	dst := []*int{&v.Major, &v.Minor, &v.Patch}
	for i := 0; i < len(parts) && i < len(dst); i++ {
		n, err := strconv.Atoi(parts[i])
		if err != nil {
			// Non-numeric component: leave this and later components at 0.
			break
		}
		*dst[i] = n
	}
	return v
}

// IsNewer reports whether a is strictly newer than b, comparing major, then
// minor, then patch numerically. Equal versions return false. Both arguments
// accept the same forms as Parse.
func IsNewer(a, b string) bool {
	return Parse(a).newerThan(Parse(b))
}

// newerThan reports whether v is strictly greater than o, component-wise.
func (v Version) newerThan(o Version) bool {
	switch {
	case v.Major != o.Major:
		return v.Major > o.Major
	case v.Minor != o.Minor:
		return v.Minor > o.Minor
	default:
		return v.Patch > o.Patch
	}
}
