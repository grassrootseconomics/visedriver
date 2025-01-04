package storage

import (
	"context"
	"fmt"
	"os"
	"path"

	"git.defalsify.org/vise.git/db"
	fsdb "git.defalsify.org/vise.git/db/fs"
	"git.defalsify.org/vise.git/db/postgres"
	"git.defalsify.org/vise.git/logging"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
	"git.grassecon.net/urdt/ussd/initializers"
	gdbmstorage "git.grassecon.net/urdt/ussd/internal/storage/db/gdbm"
)

var (
	logg = logging.NewVanilla().WithDomain("storage")
)

type StorageService interface {
	GetPersister(ctx context.Context) (*persist.Persister, error)
	GetUserdataDb(ctx context.Context) db.Db
	GetResource(ctx context.Context) (resource.Resource, error)
	SetConn(connStr string) error
}

type MenuStorageService struct {
	//dbDir         string
	conn connData
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

	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		user, password, host, port, dbName,
	)
	logg.Debugf("pg conn string", "conn", connString)

	return connString
}

func NewMenuStorageService(resourceDir string) *MenuStorageService {
	return &MenuStorageService{
		resourceDir: resourceDir,
	}
}

func (ms *MenuStorageService) SetConn(connStr string) error {
	o, err := toConnData(connStr)
	if err != nil {
		return err
	}
	ms.conn = o
	return nil
}

func (ms *MenuStorageService) getOrCreateDb(ctx context.Context, existingDb db.Db, section string) (db.Db, error) {
	var newDb db.Db
	var err error
//	database, ok := ctx.Value("Database").(string)
//	if !ok {
//		return nil, fmt.Errorf("failed to select the database")
//	}

	if existingDb != nil {
		return existingDb, nil
	}


	connStr := ms.conn.String()
	dbTyp := ms.conn.DbType()
	if dbTyp == DBTYPE_POSTGRES {
		newDb = postgres.NewPgDb()
	} else if dbTyp == DBTYPE_GDBM {
		err = ms.ensureDbDir()
		if err != nil {
			return nil, err
		}
		connStr = path.Join(connStr, section)
		newDb = gdbmstorage.NewThreadGdbmDb()
	} else {
		return nil, fmt.Errorf("unsupported connection string: %s", ms.conn.String())
	}
	logg.DebugCtxf(ctx, "connecting to db", "conn", connStr)
	err = newDb.Connect(ctx, connStr)
	if err != nil {
		return nil, err
	}

	return newDb, nil
}

func (ms *MenuStorageService) GetPersister(ctx context.Context) (*persist.Persister, error) {
	stateStore, err := ms.GetStateStore(ctx)
	if err != nil {
		return nil, err
	}

	pr := persist.NewPersister(stateStore)
	logg.TraceCtxf(ctx, "menu storage service", "persist", pr, "store", stateStore)
	return pr, nil
}

func (ms *MenuStorageService) GetUserdataDb(ctx context.Context) (db.Db, error) {
	if ms.userDataStore != nil {
		return ms.userDataStore, nil
	}

	userDataStore, err := ms.getOrCreateDb(ctx, ms.userDataStore, "userdata.gdbm")
	if err != nil {
		return nil, err
	}

	ms.userDataStore = userDataStore
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
		return ms.stateStore, nil
	}

	stateStore, err := ms.getOrCreateDb(ctx, ms.stateStore, "state.gdbm")
	if err != nil {
		return nil, err
	}

	ms.stateStore = stateStore
	return ms.stateStore, nil
}

func (ms *MenuStorageService) ensureDbDir() error {
	err := os.MkdirAll(ms.conn.String(), 0700)
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
