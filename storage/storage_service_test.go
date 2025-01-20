package storage

import (
	"context"
	"os"
	"testing"

	fsdb "git.defalsify.org/vise.git/db/fs"
)

func TestMenuStorageServiceOneSet(t *testing.T) {
	d, err := os.MkdirTemp("", "visedriver-menustorageservice")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(d)
	conns := NewConns()
	connData, err := ToConnData(d)
	if err != nil {
		t.Fatal(err)
	}
	conns.Set(connData, STORETYPE_STATE)

	ctx := context.Background()
	ms := NewMenuStorageService(conns)
	_, err = ms.GetStateStore(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = ms.GetResource(ctx)
	if err == nil {
		t.Fatalf("expected error getting resource")
	}
	_, err = ms.GetUserdataDb(ctx)
	if err == nil {
		t.Fatalf("expected error getting userdata")
	}
}

func TestMenuStorageServiceExplicit(t *testing.T) {
	d, err := os.MkdirTemp("", "visedriver-menustorageservice")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(d)
	conns := NewConns()
	connData, err := ToConnData(d)
	if err != nil {
		t.Fatal(err)
	}
	conns.Set(connData, STORETYPE_STATE)

	ctx := context.Background()
	d, err = os.MkdirTemp("", "visedriver-menustorageservice")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(d)
	store := fsdb.NewFsDb()
	err = store.Connect(ctx, d)
	if err != nil {
		t.Fatal(err)
	}
	
	ms := NewMenuStorageService(conns)
	ms = ms.WithDb(store, STORETYPE_RESOURCE)
	_, err = ms.GetStateStore(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = ms.GetResource(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = ms.GetUserdataDb(ctx)
	if err == nil {
		t.Fatalf("expected error getting userdata")
	}

}

func TestMenuStorageServiceReuse(t *testing.T) {
	d, err := os.MkdirTemp("", "visedriver-menustorageservice")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(d)
	conns := NewConns()
	connData, err := ToConnData(d)
	if err != nil {
		t.Fatal(err)
	}
	conns.Set(connData, STORETYPE_STATE)
	conns.Set(connData, STORETYPE_USER)

	ctx := context.Background()
	ctx = context.WithValue(ctx, "SessionId", "foo")
	ms := NewMenuStorageService(conns)
	stateStore, err := ms.GetStateStore(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = ms.GetResource(ctx)
	if err == nil {
		t.Fatalf("expected error getting resource")
	}
	userStore, err := ms.GetUserdataDb(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if userStore != stateStore {
		t.Fatalf("expected same store, but they are %p and %p", userStore, stateStore)
	}
}
