package utils

import (
	"context"

	"git.defalsify.org/vise.git/db"
	"git.defalsify.org/vise.git/lang"
)

type DataStore interface {
	SetPrefix(prefix uint8)
	SetSession(sessionId string)
	Get(ctx context.Context, key []byte) ([]byte, error)
	ReadEntry(ctx context.Context, sessionId string, typ DataTyp) ([]byte, error)
	WriteEntry(ctx context.Context, sessionId string, typ DataTyp, value []byte) error
	Connect(ctx context.Context, connStr string) error
	SetLanguage(*lang.Language)
	Close() error
	Prefix() uint8
	Put(ctx context.Context, key []byte, val []byte) error
	Safe() bool
	SetLock(typ uint8, locked bool) error
}

type UserDataStore struct {
	Store db.Db
}

func (store UserDataStore) SetPrefix(prefix uint8) {
	store.Store.SetPrefix(prefix)
}

func (store UserDataStore) SetLanguage(lang *lang.Language) {
	store.Store.SetLanguage(lang)
}

func (store UserDataStore) SetLock(typ uint8, locked bool) error {
	return store.Store.SetLock(typ, locked)
}
func (store UserDataStore) Safe() bool {
	return store.Store.Safe()
}

func (store UserDataStore) Put(ctx context.Context, key []byte, val []byte) error {
	return store.Store.Put(ctx, key, val)
}

func (store UserDataStore) Connect(ctx context.Context, connectionStr string) error {
	return store.Store.Connect(ctx, connectionStr)
}

func (store UserDataStore) Close() error {
	return store.Store.Close()
}

func (store UserDataStore) Prefix() uint8 {
	return store.Store.Prefix()
}

func (store UserDataStore) SetSession(sessionId string) {
	store.Store.SetSession(sessionId)
}

func (store UserDataStore) Get(ctx context.Context, key []byte) ([]byte, error) {
	return store.Store.Get(ctx, key)
}

// ReadEntry retrieves an entry from the store based on the provided parameters.
func (store UserDataStore) ReadEntry(ctx context.Context, sessionId string, typ DataTyp) ([]byte, error) {
	store.Store.SetPrefix(db.DATATYPE_USERDATA)
	store.Store.SetSession(sessionId)
	k := PackKey(typ, []byte(sessionId))
	return store.Get(ctx, k)
}

func (store UserDataStore) WriteEntry(ctx context.Context, sessionId string, typ DataTyp, value []byte) error {
	store.Store.SetPrefix(db.DATATYPE_USERDATA)
	store.Store.SetSession(sessionId)
	k := PackKey(typ, []byte(sessionId))
	return store.Store.Put(ctx, k, value)
}
