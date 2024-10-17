package storage

import (
	"context"

	"git.defalsify.org/vise.git/db"
	postgres "git.defalsify.org/vise.git/db/postgres"
	"git.defalsify.org/vise.git/lang"
)

var (
	pdbC map[string]chan db.Db
)

type ThreadPostgresDb struct {
	db      db.Db
	connStr string
}

func NewThreadPostgresDb() *ThreadPostgresDb {
	if pdbC == nil {
		pdbC = make(map[string]chan db.Db)
	}
	return &ThreadPostgresDb{}
}

func (tpdb *ThreadPostgresDb) Connect(ctx context.Context, connStr string) error {
	var ok bool
	_, ok = pdbC[connStr]
	if ok {
		logg.WarnCtxf(ctx, "already registered thread postgres, skipping", "connStr", connStr)
		return nil
	}
	postgresdb := postgres.NewPgDb().WithSchema("public")
	err := postgresdb.Connect(ctx, connStr)
	if err != nil {
		return err
	}
	pdbC[connStr] = make(chan db.Db, 1)
	pdbC[connStr] <- postgresdb
	tpdb.connStr = connStr
	return nil
}

func (tpdb *ThreadPostgresDb) reserve() {
	if tpdb.db == nil {
		tpdb.db = <-pdbC[tpdb.connStr]
	}
}

func (tpdb *ThreadPostgresDb) release() {
	if tpdb.db == nil {
		return
	}
	pdbC[tpdb.connStr] <- tpdb.db
	tpdb.db = nil
}

func (tpdb *ThreadPostgresDb) SetPrefix(pfx uint8) {
	tpdb.reserve()
	tpdb.db.SetPrefix(pfx)
}

func (tpdb *ThreadPostgresDb) SetSession(sessionId string) {
	tpdb.reserve()
	tpdb.db.SetSession(sessionId)
}

func (tpdb *ThreadPostgresDb) SetLanguage(lng *lang.Language) {
	tpdb.reserve()
	tpdb.db.SetLanguage(lng)
}

func (tpdb *ThreadPostgresDb) Safe() bool {
	tpdb.reserve()
	v := tpdb.db.Safe()
	tpdb.release()
	return v
}

func (tpdb *ThreadPostgresDb) Prefix() uint8 {
	tpdb.reserve()
	v := tpdb.db.Prefix()
	tpdb.release()
	return v
}

func (tpdb *ThreadPostgresDb) SetLock(typ uint8, locked bool) error {
	tpdb.reserve()
	err := tpdb.db.SetLock(typ, locked)
	tpdb.release()
	return err
}

func (tpdb *ThreadPostgresDb) Put(ctx context.Context, key []byte, val []byte) error {
	tpdb.reserve()
	err := tpdb.db.Put(ctx, key, val)
	tpdb.release()
	return err
}

func (tpdb *ThreadPostgresDb) Get(ctx context.Context, key []byte) ([]byte, error) {
	tpdb.reserve()
	v, err := tpdb.db.Get(ctx, key)
	tpdb.release()
	return v, err
}

func (tpdb *ThreadPostgresDb) Close() error {
	tpdb.reserve()
	close(pdbC[tpdb.connStr])
	err := tpdb.db.Close()
	tpdb.db = nil
	return err
}
