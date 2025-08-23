package gates

import "testing"

func TestGetGate(t *testing.T) {
	g, err := GetGate("clarity-form-1")
	if err != nil {
		t.Fatalf("GetGate: %v", err)
	}
	if g.ID != "clarity-form-1" || g.Question == "" {
		t.Fatalf("unexpected gate: %#v", g)
	}
}

func TestGetGateUnknown(t *testing.T) {
	if _, err := GetGate("unknown"); err == nil {
		t.Fatalf("expected error for unknown gate")
	}
}
