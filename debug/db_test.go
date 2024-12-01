package debug

import (
	"testing"
)

func TestDebugDbSubKeyInfo(t *testing.T) {
	s := "foo"
	b := []byte(s)
	b = append(b, []byte{0x40, 0x00, 0x02}...)
	r, err := ToKeyInfo(b, s)
	if err != nil {
		t.Fatal(err)
	}
	if r.SessionId != s {
		t.Fatalf("expected %s, got %s", s, r.SessionId)
	}
	if r.Typ != 64 {
		t.Fatalf("expected 64, got %d", r.Typ)
	}
	if r.SubTyp != 2 {
		t.Fatalf("expected 2, got %d", r.SubTyp)
	}
	if DebugCap & 1 > 0 {
		if r.Label != "tracking id" {
			t.Fatalf("expected 'tracking id', got '%s'", r.Label)
		}
	}
}

func TestDebugDbKeyInfo(t *testing.T) {
	s := "bar"
	b := []byte(s)
	b = append(b, []byte{0x20}...)
	r, err := ToKeyInfo(b, s)
	if err != nil {
		t.Fatal(err)
	}
	if r.SessionId != s {
		t.Fatalf("expected %s, got %s", s, r.SessionId)
	}
	if r.Typ != 32 {
		t.Fatalf("expected 64, got %d", r.Typ)
	}
	if DebugCap & 1 > 0 {
		if r.Label != "userdata" {
			t.Fatalf("expected 'userdata', got '%s'", r.Label)
		}
	}
}
