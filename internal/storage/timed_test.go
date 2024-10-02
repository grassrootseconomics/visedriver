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

	if tdb.Stale(ctx, db.DATATYPE_USERDATA, k) {
		t.Fatal("expected not stale")
	}
	time.Sleep(time.Millisecond)
	if !tdb.Stale(ctx, db.DATATYPE_USERDATA, k) {
		t.Fatal("expected stale")
	}
}
