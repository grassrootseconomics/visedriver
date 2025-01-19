package storage

import (
	"context"

	"git.defalsify.org/vise.git/db"
	"git.defalsify.org/vise.git/persist"
)

const (
	DATATYPE_EXTEND = 128
)

type Storage struct {
	Persister *persist.Persister
	UserdataDb db.Db	
}

func (s *Storage) Close(ctx context.Context) error {
	return s.UserdataDb.Close(ctx)
}

type StorageProvider interface {
	Get(ctx context.Context, sessionId string) (*Storage, error)
	Put(ctx context.Context, sessionId string, storage *Storage) error
	Close(ctx context.Context) error
}

type SimpleStorageProvider struct {
	*Storage
}

func NewSimpleStorageProvider(stateStore db.Db, userdataStore db.Db) StorageProvider {
	pe := persist.NewPersister(stateStore)
	pe = pe.WithFlush()
	return &SimpleStorageProvider{
		Storage: &Storage{
			Persister: pe,
			UserdataDb: userdataStore,
		},
	}
}

func (p *SimpleStorageProvider) Get(ctx context.Context, sessionId string) (*Storage, error) {
	p.Storage.UserdataDb.Start(ctx)
	return p.Storage, nil
}

func (p *SimpleStorageProvider) Put(ctx context.Context, sessionId string, storage *Storage) error {
	storage.UserdataDb.Stop(ctx)
	return nil
}

func (p *SimpleStorageProvider) Close(ctx context.Context) error {
	return p.Storage.Close(ctx)
}
