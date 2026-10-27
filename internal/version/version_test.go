package version

import "testing"

func TestCurrentDefaultAndSet(t *testing.T) {
	if Current != "dev" {
		t.Errorf("default Current = %q, want %q", Current, "dev")
	}
	t.Cleanup(func() { Current = "dev" })
	SetCurrent("4.5.6")
	if Current != "4.5.6" {
		t.Errorf("after SetCurrent, Current = %q, want %q", Current, "4.5.6")
	}
}

func TestParse(t *testing.T) {
	cases := []struct {
		in   string
		want Version
	}{
		{"1.2.3", Version{1, 2, 3}},
		{"v1.2.3", Version{1, 2, 3}},      // leading v stripped
		{"1.2.3-dirty", Version{1, 2, 3}}, // suffix after - ignored
		{"v1.2.3-rc1", Version{1, 2, 3}},
		{"1.2", Version{1, 2, 0}}, // missing component -> 0
		{"1", Version{1, 0, 0}},
		{"dev", Version{0, 0, 0}}, // non-numeric -> zero version, no panic
		{"", Version{0, 0, 0}},
		{"v1.x.3", Version{1, 0, 0}}, // non-numeric component stops parsing
		{"1.10.0", Version{1, 10, 0}},
	}
	for _, c := range cases {
		if got := Parse(c.in); got != c.want {
			t.Errorf("Parse(%q) = %+v, want %+v", c.in, got, c.want)
		}
	}
}

func TestIsNewer(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"1.10.0", "1.9.9", true}, // numeric, not lexical
		{"2.0.0", "1.9.9", true},
		{"1.1.0", "1.0.0", true},
		{"1.0.1", "1.0.0", true},
		{"1.0.0", "1.0.0", false}, // equal -> not newer
		{"1.0.0", "1.1.0", false},
		{"1.9.9", "2.0.0", false},
		{"v1.2.3", "1.2.2", true},       // v-prefix accepted on both sides
		{"1.2.3-dirty", "1.2.3", false}, // suffix ignored -> equal
		{"1.0.0", "dev", true},          // any release beats a dev/zero version
		{"dev", "1.0.0", false},
	}
	for _, c := range cases {
		if got := IsNewer(c.a, c.b); got != c.want {
			t.Errorf("IsNewer(%q, %q) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}
