package dubai

import (
	"testing"
	"time"
)

func TestValidDate(t *testing.T) {
	good := []string{"2026-11-18", "2028-02-29"}
	bad := []string{"", "2026-13-01", "2026-02-30", "2027-02-29", "18-11-2026", "2026-1-5", "tomorrow"}
	for _, s := range good {
		if !ValidDate(s) {
			t.Errorf("ValidDate(%q) = false, want true", s)
		}
	}
	for _, s := range bad {
		if ValidDate(s) {
			t.Errorf("ValidDate(%q) = true, want false", s)
		}
	}
}

func TestTodayIsAValidDate(t *testing.T) {
	if !ValidDate(Today()) {
		t.Fatalf("Today() = %q is not a valid date", Today())
	}
}

func TestDateOffset(t *testing.T) {
	if DateOffset(0) != Today() {
		t.Errorf("DateOffset(0) = %q, want today %q", DateOffset(0), Today())
	}
	if !(DateOffset(-1) < Today() && Today() < DateOffset(1)) {
		t.Errorf("offsets are not in order: %s %s %s", DateOffset(-1), Today(), DateOffset(1))
	}
}

func TestUntilMidnight(t *testing.T) {
	d := UntilMidnight()
	if d <= 0 || d > 24*time.Hour {
		t.Fatalf("UntilMidnight() = %v, want between 0 and 24h", d)
	}
}

func TestLocationIsDubai(t *testing.T) {
	_, off := time.Date(2026, 1, 1, 12, 0, 0, 0, Location()).Zone()
	if off != 4*3600 {
		t.Fatalf("Dubai offset = %d seconds, want 14400 (UTC+4, no daylight saving)", off)
	}
}
