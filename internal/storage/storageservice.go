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
	GetPersister(dbDir string, ctx context.Context) (*persist.Persister, error)
	GetUserdataDb(dbDir string, ctx context.Context) db.Db
	GetResource(resourceDir string, ctx context.Context) (resource.Resource, error)
	EnsureDbDir(dbDir string) error
}

type MenuStorageService struct{}

func (menuStorageService *MenuStorageService) GetPersister(dbDir string, ctx context.Context) (*persist.Persister, error) {
	store := gdbmdb.NewGdbmDb()
	storeFile := path.Join(dbDir, "state.gdbm")
	store.Connect(ctx, storeFile)
	pr := persist.NewPersister(store)
	return pr, nil
}

func (menuStorageService *MenuStorageService) GetUserdataDb(dbDir string, ctx context.Context) db.Db {
	store := gdbmdb.NewGdbmDb()
	storeFile := path.Join(dbDir, "userdata.gdbm")
	store.Connect(ctx, storeFile)

	return store
}

func (menuStorageService *MenuStorageService) GetResource(resourceDir string, ctx context.Context) (resource.Resource, error) {
	store := fsdb.NewFsDb()
	err := store.Connect(ctx, resourceDir)
	if err != nil {
		return nil, err
	}
	rfs := resource.NewDbResource(store)
	return rfs, nil
}

func (menuStorageService *MenuStorageService) GetStateStore(dbDir string, ctx context.Context) (db.Db, error) {
	store := gdbmdb.NewGdbmDb()
	storeFile := path.Join(dbDir, "state.gdbm")
	store.Connect(ctx, storeFile)
	return store, nil
}

func (menuStorageService *MenuStorageService) EnsureDbDir(dbDir string) error {
	err := os.MkdirAll(dbDir, 0700)
	if err != nil {
		return fmt.Errorf("state dir create exited with error: %v\n", err)
	}
	return nil
}
