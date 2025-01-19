package storage

import (
	"context"

	"git.defalsify.org/vise.git/db"
	gdbmdb "git.defalsify.org/vise.git/db/gdbm"
	"git.defalsify.org/vise.git/lang"
	"git.defalsify.org/vise.git/logging"
)

var (
	logg = logging.NewVanilla().WithDomain("gdbmstorage")
)

var (
	dbC map[string]chan db.Db
)

type ThreadGdbmDb struct {
	db db.Db
	connStr string
}

func NewThreadGdbmDb() *ThreadGdbmDb {
	if dbC == nil {
		dbC = make(map[string]chan db.Db)
	}
	return &ThreadGdbmDb{}
}

func(tdb *ThreadGdbmDb) Connect(ctx context.Context, connStr string) error {
	var ok bool
	_, ok = dbC[connStr]
	if ok {
		logg.WarnCtxf(ctx, "already registered thread gdbm, skipping", "connStr", connStr)
		return nil
	}
	gdb := gdbmdb.NewGdbmDb()
	err := gdb.Connect(ctx, connStr)
	if err != nil {
		return err
	}
	dbC[connStr] = make(chan db.Db, 1)
	dbC[connStr]<- gdb
	tdb.connStr = connStr
	return nil
}

func(tdb *ThreadGdbmDb) reserve() {
	if tdb.db == nil {
		tdb.db = <-dbC[tdb.connStr]
	}
}

func(tdb *ThreadGdbmDb) release() {
	if tdb.db == nil {
		return
	}
	dbC[tdb.connStr] <- tdb.db
	tdb.db = nil
}

func(tdb *ThreadGdbmDb) SetPrefix(pfx uint8) {
	tdb.reserve()
	tdb.db.SetPrefix(pfx)
}

func(tdb *ThreadGdbmDb) SetSession(sessionId string) {
	tdb.reserve()
	tdb.db.SetSession(sessionId)
}

func(tdb *ThreadGdbmDb) SetLanguage(lng *lang.Language) {
	tdb.reserve()
	tdb.db.SetLanguage(lng)
}

func(tdb *ThreadGdbmDb) Safe() bool {
	tdb.reserve()
	v := tdb.db.Safe()
	tdb.release()
	return v
}

func(tdb *ThreadGdbmDb) Prefix() uint8 {
	tdb.reserve()
	v := tdb.db.Prefix()
	tdb.release()
	return v
}

func(tdb *ThreadGdbmDb) SetLock(typ uint8, locked bool) error {
	tdb.reserve()
	err := tdb.db.SetLock(typ, locked)
	tdb.release()
	return err
}

func(tdb *ThreadGdbmDb) Put(ctx context.Context, key []byte, val []byte) error {
	tdb.reserve()
	err := tdb.db.Put(ctx, key, val)
	tdb.release()
	return err
}

func(tdb *ThreadGdbmDb) Get(ctx context.Context, key []byte) ([]byte, error) {
	tdb.reserve()
	v, err := tdb.db.Get(ctx, key)
	tdb.release()
	return v, err
}

func(tdb *ThreadGdbmDb) Close(ctx context.Context) error {
	tdb.reserve()
	close(dbC[tdb.connStr])
	delete(dbC, tdb.connStr)
	err := tdb.db.Close(ctx)
	tdb.db = nil
	return err
}

func(tdb *ThreadGdbmDb) Dump(ctx context.Context, key []byte) (*db.Dumper, error) {
	tdb.reserve()
	defer tdb.release()
	return tdb.db.Dump(ctx, key)
}

func(tdb *ThreadGdbmDb) DecodeKey(ctx context.Context, key []byte) ([]byte, error) {
	return tdb.db.DecodeKey(ctx, key)
}

func(tdb *ThreadGdbmDb) Abort(ctx context.Context) {
	tdb.db.Abort(ctx)
}

func(tdb *ThreadGdbmDb) Start(ctx context.Context) error {
	return tdb.db.Start(ctx)
}

func(tdb *ThreadGdbmDb) Stop(ctx context.Context) error {
	return tdb.db.Stop(ctx)
}

func(tdb *ThreadGdbmDb) Connection() string {
	return tdb.db.Connection()
}
