package storage

import (
	"context"

	"git.defalsify.org/vise.git/db"
	postgres "git.defalsify.org/vise.git/db/postgres"
	"git.defalsify.org/vise.git/lang"
)

type PostgresDb struct {
	db      db.Db
	connStr string
}

func NewPostgresDb() *PostgresDb {
	return &PostgresDb{}
}

func (pdb *PostgresDb) Connect(ctx context.Context, connStr string) error {
	if pdb.db != nil {
		logg.WarnCtxf(ctx, "already connected, skipping", "connStr", connStr)
		return nil
	}
	postgresdb := postgres.NewPgDb().WithSchema("public")
	err := postgresdb.Connect(ctx, connStr)
	if err != nil {
		return err
	}
	pdb.db = postgresdb
	pdb.connStr = connStr
	return nil
}

func (pdb *PostgresDb) SetPrefix(pfx uint8) {
	pdb.db.SetPrefix(pfx)
}

func (pdb *PostgresDb) SetSession(sessionId string) {
	pdb.db.SetSession(sessionId)
}

func (pdb *PostgresDb) SetLanguage(lng *lang.Language) {
	pdb.db.SetLanguage(lng)
}

func (pdb *PostgresDb) Safe() bool {
	return pdb.db.Safe()
}

func (pdb *PostgresDb) Prefix() uint8 {
	return pdb.db.Prefix()
}

func (pdb *PostgresDb) SetLock(typ uint8, locked bool) error {
	return pdb.db.SetLock(typ, locked)
}

func (pdb *PostgresDb) Put(ctx context.Context, key []byte, val []byte) error {
	return pdb.db.Put(ctx, key, val)
}

func (pdb *PostgresDb) Get(ctx context.Context, key []byte) ([]byte, error) {
	return pdb.db.Get(ctx, key)
}

func (pdb *PostgresDb) Close() error {
	if pdb.db == nil {
		return nil
	}
	err := pdb.db.Close()
	pdb.db = nil
	return err
}