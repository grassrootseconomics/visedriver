package storage

import (
	"bytes"
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
	parentSession []byte
	matchPfx map[uint8][][]byte
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

func(tib *TimedDb) WithMatch(pfx uint8, keyPart []byte) *TimedDb {
	if tib.matchPfx == nil {
		tib.matchPfx = make(map[uint8][][]byte)
	}
	tib.matchPfx[pfx] = append(tib.matchPfx[pfx], keyPart)
	return tib
}

func(tib *TimedDb) checkPrefix(pfx uint8, key []byte) bool {
	var v []byte
	if tib.matchPfx == nil {
		return true
	}
	for _, v = range(tib.matchPfx[pfx]) {
		l := len(v)
		if l > len(key) {
			continue
		}
		if bytes.Equal(v, key[:l]) {
			return true	
		}
	}
	return false
}

func(tib *TimedDb) SetPrefix(pfx uint8) {
	tib.Db.SetPrefix(pfx)
	tib.parentPfx = pfx
}

func(tib *TimedDb) SetSession(session string) {
	tib.Db.SetSession(session)
	tib.parentSession = []byte(session)
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
	defer func() {
		tib.parentPfx = 0
		tib.parentSession = nil
	}()
	if tib.checkPrefix(tib.parentPfx, key) {
		tib.tdb.SetSession("")
		k := db.ToSessionKey(tib.parentPfx, []byte(tib.parentSession), key)
		k = append([]byte{tib.parentPfx}, k...)
		err = tib.tdb.Put(ctx, k, b)
		if err != nil {
			logg.ErrorCtxf(ctx, "failed to update timestamp of record", err)
		}
	}
	return nil
}

func(tib *TimedDb) Stale(ctx context.Context, pfx uint8, sessionId string, key []byte) bool {
	tib.tdb.SetSession("")
	b := db.ToSessionKey(pfx, []byte(sessionId), key)
	b = append([]byte{pfx}, b...)
	v, err := tib.tdb.Get(ctx, b)
	if err != nil {
		logg.WarnCtxf(ctx, "no time entry", "key", key, "b", b)
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
