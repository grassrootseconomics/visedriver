package storage

import (
	"context"
	"fmt"
	"os"
	"path"

	"git.defalsify.org/vise.git/db"
	fsdb "git.defalsify.org/vise.git/db/fs"
	gdbmdb "git.defalsify.org/vise.git/db/gdbm"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
)

type StorageService interface {
	GetPersister(ctx context.Context) (*persist.Persister, error)
	GetUserdataDb(ctx context.Context) db.Db
	GetResource(ctx context.Context) (resource.Resource, error)
	EnsureDbDir() error
}

type MenuStorageService struct{
	dbDir string
	resourceDir string
}

func NewMenuStorageService(dbDir string, resourceDir string) *MenuStorageService {
	return &MenuStorageService{
		dbDir: dbDir,
		resourceDir: resourceDir,
	}
}

func (ms *MenuStorageService) GetPersister(ctx context.Context) (*persist.Persister, error) {
	store := gdbmdb.NewGdbmDb()
	storeFile := path.Join(ms.dbDir, "state.gdbm")
	store.Connect(ctx, storeFile)
	pr := persist.NewPersister(store)
	return pr, nil
}

func (ms *MenuStorageService) GetUserdataDb(ctx context.Context) db.Db {
	store := gdbmdb.NewGdbmDb()
	storeFile := path.Join(ms.dbDir, "userdata.gdbm")
	store.Connect(ctx, storeFile)
	return store
}

func (ms *MenuStorageService) GetResource(ctx context.Context) (resource.Resource, error) {
	store := fsdb.NewFsDb()
	err := store.Connect(ctx, ms.resourceDir)
	if err != nil {
		return nil, err
	}
	rfs := resource.NewDbResource(store)
	return rfs, nil
}

func (ms *MenuStorageService) GetStateStore(ctx context.Context) (db.Db, error) {
	store := gdbmdb.NewGdbmDb()
	storeFile := path.Join(ms.dbDir, "state.gdbm")
	store.Connect(ctx, storeFile)
	return store, nil
}

func (ms *MenuStorageService) EnsureDbDir() error {
	err := os.MkdirAll(ms.dbDir, 0700)
	if err != nil {
		return fmt.Errorf("state dir create exited with error: %v\n", err)
	}
	return nil
}
