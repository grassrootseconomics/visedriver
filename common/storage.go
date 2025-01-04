package common

import (
	"context"
	"errors"

	"git.defalsify.org/vise.git/db"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/persist"
	"git.grassecon.net/urdt/ussd/internal/storage"
	dbstorage "git.grassecon.net/urdt/ussd/internal/storage/db"
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
	SetConn(connStr string) error
}

type StorageService struct {
	svc *storage.MenuStorageService
}

func NewStorageService(connStr string) (*StorageService, error) {
	svc := &StorageService{
		svc: storage.NewMenuStorageService(""),
	}
	err := svc.SetConn(connStr)
	return svc, err
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

func(ss *StorageService) SetConn(connStr string) error {
	return ss.svc.SetConn(connStr)
}
