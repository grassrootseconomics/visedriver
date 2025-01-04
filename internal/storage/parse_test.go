package storage

import (
	"testing"
)

func TestParseConnStr(t *testing.T) {
	svc := NewMenuStorageService("")
	err := svc.SetConn("postgres://foo:bar@localhost:5432/baz")
	if err != nil {
		t.Fatal(err)	
	}
	err = svc.SetConn("/foo/bar/baz.gdbm")
	if err != nil {
		t.Fatal(err)	
	}
	err = svc.SetConn("foo/bar/baz.gdbm")
	if err == nil {
		t.Fatalf("expected error")
	}
	err = svc.SetConn("http://foo/bar")
	if err == nil {
		t.Fatalf("expected error")
	}
	err = svc.SetConn("foo/bar/baz.txt")
	if err == nil {
		t.Fatalf("expected error")
	}
	err = svc.SetConn("/foo/bar")
	if err == nil {
		t.Fatalf("expected error")
	}
	err = svc.SetConn("foo/bar")
	if err == nil {
		t.Fatalf("expected error")
	}
}
