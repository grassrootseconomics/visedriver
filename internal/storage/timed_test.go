package storage

import (
	"context"
	"testing"
	"time"

	"git.defalsify.org/vise.git/db"
	memdb "git.defalsify.org/vise.git/db/mem"
)

func TestStaleDb(t *testing.T) {
	ctx := context.Background()
	mdb := memdb.NewMemDb()
	err := mdb.Connect(ctx, "")
	if err != nil {
		t.Fatal(err)
	}

	tdb := NewTimedDb(mdb, time.Duration(time.Millisecond))
	tdb.SetPrefix(db.DATATYPE_USERDATA)
	k := []byte("foo")
	err = tdb.Put(ctx, k, []byte("bar"))
	if err != nil {
		t.Fatal(err)
	}

	if tdb.Stale(ctx, db.DATATYPE_USERDATA, "", k) {
		t.Fatal("expected not stale")
	}
	time.Sleep(time.Millisecond)
	if !tdb.Stale(ctx, db.DATATYPE_USERDATA, "", k) {
		t.Fatal("expected stale")
	}
}

func TestFilteredStaleDb(t *testing.T) {
	ctx := context.Background()
	mdb := memdb.NewMemDb()
	err := mdb.Connect(ctx, "")
	if err != nil {
		t.Fatal(err)
	}

	k := []byte("foo")
	tdb := NewTimedDb(mdb, time.Duration(time.Millisecond))
	tdb = tdb.WithMatch(db.DATATYPE_STATE, []byte("in"))
	tdb.SetPrefix(db.DATATYPE_USERDATA)
	tdb.SetSession("inky")
	err = tdb.Put(ctx, k, []byte("bar"))
	if err != nil {
		t.Fatal(err)
	}
	tdb.SetPrefix(db.DATATYPE_STATE)
	tdb.SetSession("inky")
	err = tdb.Put(ctx, k, []byte("pinky"))
	if err != nil {
		t.Fatal(err)
	}
	tdb.SetSession("blinky")
	err = tdb.Put(ctx, k, []byte("clyde"))
	if err != nil {
		t.Fatal(err)
	}

	if tdb.Stale(ctx, db.DATATYPE_USERDATA, "inky", k) {
		t.Fatal("expected not stale")
	}
	if tdb.Stale(ctx, db.DATATYPE_STATE, "inky", k) {
		t.Fatal("expected not stale")
	}
	if tdb.Stale(ctx, db.DATATYPE_STATE, "blinky", k) {
		t.Fatal("expected not stale")
	}
	time.Sleep(time.Millisecond)
	if tdb.Stale(ctx, db.DATATYPE_USERDATA, "inky", k) {
		t.Fatal("expected not stale")
	}
	if !tdb.Stale(ctx, db.DATATYPE_STATE, "inky", k) {
		t.Fatal("expected stale")
	}
	if tdb.Stale(ctx, db.DATATYPE_STATE, "blinky", k) {
		t.Fatal("expected not stale")
	}
}
