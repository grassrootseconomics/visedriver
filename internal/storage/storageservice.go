package storage

import (
	"context"
	"fmt"
	"os"
	"path"

	"git.defalsify.org/vise.git/db"
	fsdb "git.defalsify.org/vise.git/db/fs"
	"git.defalsify.org/vise.git/logging"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
	"git.grassecon.net/urdt/ussd/initializers"
)

var (
	logg = logging.NewVanilla().WithDomain("storage")
)

type StorageService interface {
	GetPersister(ctx context.Context) (*persist.Persister, error)
	GetUserdataDb(ctx context.Context) db.Db
	GetResource(ctx context.Context) (resource.Resource, error)
	EnsureDbDir() error
}

type MenuStorageService struct {
	dbDir         string
	resourceDir   string
	resourceStore db.Db
	stateStore    db.Db
	userDataStore db.Db
}

func buildConnStr() string {
	host := initializers.GetEnv("DB_HOST", "localhost")
	user := initializers.GetEnv("DB_USER", "postgres")
	password := initializers.GetEnv("DB_PASSWORD", "")
	dbName := initializers.GetEnv("DB_NAME", "")
	port := initializers.GetEnv("DB_PORT", "5432")

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		user, password, host, port, dbName,
	)
}

func NewMenuStorageService(dbDir string, resourceDir string) *MenuStorageService {
	return &MenuStorageService{
		dbDir:       dbDir,
		resourceDir: resourceDir,
	}
}

func (ms *MenuStorageService) GetPersister(ctx context.Context) (*persist.Persister, error) {
	ms.stateStore = NewThreadGdbmDb()
	storeFile := path.Join(ms.dbDir, "state.gdbm")
	err := ms.stateStore.Connect(ctx, storeFile)
	if err != nil {
		return nil, err
	}
	pr := persist.NewPersister(ms.stateStore)
	logg.TraceCtxf(ctx, "menu storage service", "persist", pr, "store", ms.stateStore)
	return pr, nil
}

func (ms *MenuStorageService) GetUserdataDb(ctx context.Context) (db.Db, error) {
	database, ok := ctx.Value("Database").(string)
	if !ok {
		return nil, fmt.Errorf("failed to select the database")
	}

	if database == "postgres" {
		ms.userDataStore = NewThreadPostgresDb()
		connStr := buildConnStr()
		err := ms.userDataStore.Connect(ctx, connStr)
		if err != nil {
			return nil, err
		}
	} else {
		ms.userDataStore = NewThreadGdbmDb()
		storeFile := path.Join(ms.dbDir, "userdata.gdbm")
		err := ms.userDataStore.Connect(ctx, storeFile)
		if err != nil {
			return nil, err
		}
	}

	return ms.userDataStore, nil
}

func (ms *MenuStorageService) GetResource(ctx context.Context) (resource.Resource, error) {
	ms.resourceStore = fsdb.NewFsDb()
	err := ms.resourceStore.Connect(ctx, ms.resourceDir)
	if err != nil {
		return nil, err
	}
	rfs := resource.NewDbResource(ms.resourceStore)
	return rfs, nil
}

func (ms *MenuStorageService) GetStateStore(ctx context.Context) (db.Db, error) {
	if ms.stateStore != nil {
		panic("set up store when already exists")
	}
	ms.stateStore = NewThreadGdbmDb()
	storeFile := path.Join(ms.dbDir, "state.gdbm")
	err := ms.stateStore.Connect(ctx, storeFile)
	if err != nil {
		return nil, err
	}
	return ms.stateStore, nil
}

func (ms *MenuStorageService) EnsureDbDir() error {
	err := os.MkdirAll(ms.dbDir, 0700)
	if err != nil {
		return fmt.Errorf("state dir create exited with error: %v\n", err)
	}
	return nil
}

func (ms *MenuStorageService) Close() error {
	errA := ms.stateStore.Close()
	errB := ms.userDataStore.Close()
	errC := ms.resourceStore.Close()
	if errA != nil || errB != nil || errC != nil {
		return fmt.Errorf("%v %v %v", errA, errB, errC)
	}
	return nil
}
