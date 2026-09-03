package model

import "testing"

func TestStripPin(t *testing.T) {
	cases := []struct {
		in       string
		wantBare string
		wantPin  int
		wantPin2 bool
	}{
		{"UP-001~3", "UP-001", 3, true},
		{"UP-001", "UP-001", 0, false},
		{"UP-001~0", "UP-001", 0, true},
		{"UP-001~abc", "UP-001~abc", 0, false}, // non-numeric suffix is NOT a pin
		{"", "", 0, false},
	}
	for _, c := range cases {
		bare, pin, pinned := StripPin(c.in)
		if bare != c.wantBare || pin != c.wantPin || pinned != c.wantPin2 {
			t.Errorf("StripPin(%q) = (%q, %d, %v), want (%q, %d, %v)",
				c.in, bare, pin, pinned, c.wantBare, c.wantPin, c.wantPin2)
		}
	}
}
