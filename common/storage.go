package common

import (
	"context"
	"errors"

	"git.defalsify.org/vise.git/db"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/persist"
	"git.grassecon.net/grassrootseconomics/visedriver/storage"
	dbstorage "git.grassecon.net/grassrootseconomics/visedriver/storage/db"
)

var (
	ToConnData = storage.ToConnData
)

func StoreToDb(store *UserDataStore) db.Db {
	return store.Db
}

func StoreToPrefixDb(store *UserDataStore, pfx []byte) dbstorage.PrefixDb {
	return dbstorage.NewSubPrefixDb(store.Db, pfx)	
}

type StorageServices interface {
	GetPersister(ctx context.Context) (*persist.Persister, error)
	GetUserdataDb(ctx context.Context) (db.Db, error)
	GetResource(ctx context.Context) (resource.Resource, error)
}

type StorageService struct {
	svc *storage.MenuStorageService
}

func NewStorageService(conn storage.ConnData) (*StorageService, error) {
	svc := &StorageService{
		svc: storage.NewMenuStorageService(conn, ""),
	}
	return svc, nil
}

// TODO: simplify enable poresource, conndata instead
func(ss *StorageService) SetResourceDir(resourceDir string) error {
	ss.svc = ss.svc.WithResourceDir(resourceDir)
	return nil
}

func(ss *StorageService) GetPersister(ctx context.Context) (*persist.Persister, error) {
	return ss.svc.GetPersister(ctx)
}
	
func(ss *StorageService) GetUserdataDb(ctx context.Context) (db.Db, error) {
	return ss.svc.GetUserdataDb(ctx)
}

func(ss *StorageService) GetResource(ctx context.Context) (resource.Resource, error) {
	return nil, errors.New("not implemented")
}

func(ss *StorageService) GetStateStore(ctx context.Context) (db.Db, error) {
	return ss.svc.GetStateStore(ctx)
}
