package gates

import "testing"

func TestGetGate(t *testing.T) {
	ids := []string{"clarity-form-1", "consistency-1"}
	for _, id := range ids {
		g, err := GetGate(id)
		if err != nil {
			t.Fatalf("GetGate(%s): %v", id, err)
		}
		if g.ID != id || g.Question == "" {
			t.Fatalf("unexpected gate for %s: %#v", id, g)
		}
	}
}

func TestGetGateUnknown(t *testing.T) {
	if _, err := GetGate("unknown"); err == nil {
		t.Fatalf("expected error for unknown gate")
	}
}
