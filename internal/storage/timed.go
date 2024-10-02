package storage

import (
	"context"
	"time"
	"encoding/binary"

	"git.defalsify.org/vise.git/db"
)

type TimedDb struct {
	db.Db
	tdb *SubPrefixDb
	ttl time.Duration
	parentPfx uint8
}


func NewTimedDb(db db.Db, ttl time.Duration) *TimedDb {
	var b [2]byte
	binary.BigEndian.PutUint16(b[:], SUBPREFIX_TIME)
	sdb := NewSubPrefixDb(db, b[:])
	return &TimedDb{
		Db: db,
		tdb: sdb,
		ttl: ttl,
	}
}

func(tib *TimedDb) SetPrefix(pfx uint8) {
	tib.Db.SetPrefix(pfx)
	tib.parentPfx = pfx
}

func(tib *TimedDb) Put(ctx context.Context, key []byte, val []byte) error {
	t := time.Now()	
	b, err := t.MarshalBinary()
	if err != nil {
		return err
	}
	err = tib.Db.Put(ctx, key, val)
	if err != nil {
		return err
	}
	k := append([]byte{tib.parentPfx}, key...)
	defer func() {
		tib.parentPfx = 0
	}()
	err = tib.tdb.Put(ctx, k, b)
	if err != nil {
		logg.ErrorCtxf(ctx, "failed to update timestamp of record", err)
	}
	return nil
}

func(tib *TimedDb) Get(ctx context.Context, key []byte) ([]byte, error) {
	v, err := tib.Db.Get(ctx, key)
	return v, err
}

func(tib *TimedDb) Stale(ctx context.Context, pfx uint8, key []byte) bool {
	b := append([]byte{pfx}, key...)
	v, err := tib.tdb.Get(ctx, b)
	if err != nil {
		logg.ErrorCtxf(ctx, "no time entry", "key", key)
		return false
	}
	t_now := time.Now()	
	t_then := time.Time{}
	err = t_then.UnmarshalBinary(v)
	if err != nil {
		return false
	}
	return t_now.After(t_then.Add(tib.ttl))
}
