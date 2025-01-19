package storage

import (
	"testing"
)

func TestParseConnStr(t *testing.T) {
	v, err := ToConnData("postgres://foo:bar@localhost:5432/baz")
	if err != nil {
		t.Fatal(err)	
	}
	if v.DbType() != DBTYPE_POSTGRES {
		t.Fatalf("expected type %v, got %v", DBTYPE_POSTGRES, v.DbType())
	}
	v, err = ToConnData("gdbm:///foo/bar")
	if err != nil {
		t.Fatal(err)	
	}
	if v.DbType() != DBTYPE_GDBM {
		t.Fatalf("expected type %v, got %v", DBTYPE_GDBM, v.DbType())
	}
	v, err = ToConnData("/foo/bar")
	if err != nil {
		t.Fatal(err)	
	}
	if v.DbType() != DBTYPE_FS {
		t.Fatalf("expected type %v, got %v", DBTYPE_FS, v.DbType())
	}
	v, err = ToConnData("/foo/bar/")
	if err != nil {
		t.Fatal(err)	
	}
	if v.DbType() != DBTYPE_FS {
		t.Fatalf("expected type %v, got %v", DBTYPE_FS, v.DbType())
	}
	v, err = ToConnData("foo/bar")
	if err != nil {
		t.Fatal(err)
	}
	if v.DbType() != DBTYPE_FS {
		t.Fatalf("expected type %v, got %v", DBTYPE_FS, v.DbType())
	}
	v, err = ToConnData("")
	if err != nil {
		t.Fatal(err)
	}
	if v.DbType() != DBTYPE_MEM {
		t.Fatalf("expected type %v, got %v", DBTYPE_MEM, v.DbType())
	}
	v, err = ToConnData("http://foo/bar")
	if err == nil {
		t.Fatalf("expected error")
	}
	if v.DbType() != DBTYPE_NONE {
		t.Fatalf("expected type %v, got %v", DBTYPE_NONE, v.DbType())
	}
}
