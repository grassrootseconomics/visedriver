package storage

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	
	"github.com/jackc/pgx/v5/pgxpool"
	"git.defalsify.org/vise.git/db"
	fsdb "git.defalsify.org/vise.git/db/fs"
	memdb "git.defalsify.org/vise.git/db/mem"
	"git.defalsify.org/vise.git/db/postgres"
	"git.defalsify.org/vise.git/lang"
	"git.defalsify.org/vise.git/logging"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
	gdbmstorage "git.grassecon.net/grassrootseconomics/visedriver/storage/db/gdbm"
)

var (
	logg = logging.NewVanilla().WithDomain("storage")
)

type StorageService interface {
	GetPersister(ctx context.Context) (*persist.Persister, error)
	GetUserdataDb(ctx context.Context) (db.Db, error)
	GetResource(ctx context.Context) (resource.Resource, error)
}

type MenuStorageService struct {
	conns Conns
	poResource    resource.Resource
	store map[int8]db.Db
}

func NewMenuStorageService(conn Conns) *MenuStorageService {
	return &MenuStorageService{
		conns: conn,
		store: make(map[int8]db.Db),
	}
}

func (ms *MenuStorageService) WithDb(store db.Db, typ int8) *MenuStorageService {
	var err error
	if ms.store[typ] != nil {
		panic(fmt.Errorf("db already set for typ: %d", typ))
	}
	ms.store[typ] = store
	ms.conns[typ], err = ToConnData(store.Connection())
	if err != nil {
		panic(err)
	}
	return ms
}

func (ms *MenuStorageService) checkDb(ctx context.Context,typ int8) db.Db {
	store := ms.store[typ]
	if store != nil {
		return store
	}
	connData := ms.conns[typ]
	v := ms.conns.Have(&connData)
	if v == -1 {
		return nil
	}
	src := ms.store[v]
	if src == nil {
		return nil
	}
	ms.store[typ] = ms.store[v]
	logg.DebugCtxf(ctx, "found existing db", "typ", typ, "srctyp", v, "store", ms.store[typ], "srcstore", ms.store[v])
	return ms.store[typ]
}

func (ms *MenuStorageService) getOrCreateDb(ctx context.Context, section string, typ int8) (db.Db, error) {
	var err error

	newDb := ms.checkDb(ctx, typ)
	if newDb != nil {
		logg.InfoCtxf(ctx, "using existing db", "typ", typ, "db", newDb)
		return newDb, nil
	}

	connData := ms.conns[typ]
	connStr := connData.String()
	dbTyp := connData.DbType()
	if dbTyp == DBTYPE_POSTGRES {
		// TODO: move to vise
		err = ensureSchemaExists(ctx, connData)
		if err != nil {
			return nil, err
		}
		newDb = postgres.NewPgDb().WithSchema(connData.Domain())
	} else if dbTyp == DBTYPE_GDBM {
		err = ms.ensureDbDir(connStr)
		if err != nil {
			return nil, err
		}
		connStr = path.Join(connStr, section)
		newDb = gdbmstorage.NewThreadGdbmDb()
	} else if dbTyp == DBTYPE_FS {
		err = ms.ensureDbDir(connStr)
		if err != nil {
			return nil, err
		}
		newDb = fsdb.NewFsDb().WithBinary()
	} else if dbTyp == DBTYPE_MEM {
		logg.WarnCtxf(ctx, "using volatile storage (memdb)")
		newDb = memdb.NewMemDb()
	} else {
		return nil, fmt.Errorf("unsupported connection string: '%s'\n", connData.String())
	}
	logg.InfoCtxf(ctx, "connecting to db", "conn", connData, "typ", typ)
	err = newDb.Connect(ctx, connStr)
	if err != nil {
		return nil, err
	}
	ms.store[typ] = newDb

	return newDb, nil
}

// WithGettext triggers use of gettext for translation of templates and menus.
//
// The first language in `lns` will be used as default language, to resolve node keys to 
// language strings.
//
// If `lns` is an empty array, gettext will not be used.
func (ms *MenuStorageService) WithGettext(path string, lns []lang.Language) *MenuStorageService {
	if len(lns) == 0 {
		logg.Warnf("Gettext requested but no languages supplied")
		return ms
	}
	rs := resource.NewPoResource(lns[0], path)

	for _, ln := range(lns) {
		rs = rs.WithLanguage(ln)
	}

	ms.poResource = rs

	return ms
}

// ensureSchemaExists creates a new schema if it does not exist
func ensureSchemaExists(ctx context.Context, conn ConnData) error {
	h, err := pgxpool.New(ctx, conn.Path())
	if err != nil {
		return fmt.Errorf("failed to connect to the database: %w", err)
	}
	defer h.Close()

	query := fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", conn.Domain())
	_, err = h.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

func applySession(ctx context.Context, store db.Db) error {
	sessionId, ok := ctx.Value("SessionId").(string)
	if !ok {
		return fmt.Errorf("missing session to apply to store: %v", store)
	}
	store.SetSession(sessionId)
	return nil
}

func (ms *MenuStorageService) GetPersister(ctx context.Context) (*persist.Persister, error) {
	stateStore, err := ms.GetStateStore(ctx)
	if err != nil {
		return nil, err
	}
	err = applySession(ctx, stateStore)
	if err != nil {
		return nil, err
	}

	pr := persist.NewPersister(stateStore)
	logg.TraceCtxf(ctx, "menu storage service", "persist", pr, "store", stateStore)
	return pr, nil
}

func (ms *MenuStorageService) GetUserdataDb(ctx context.Context) (db.Db, error) {
	userStore, err := ms.getOrCreateDb(ctx, "userdata.gdbm", STORETYPE_USER)
	if err != nil {
		return nil, err
	}

	err = applySession(ctx, userStore)
	if err != nil {
		return nil, err
	}
	return userStore, nil
}

func (ms *MenuStorageService) GetResource(ctx context.Context) (resource.Resource, error) {
	store, err := ms.getOrCreateDb(ctx, "resource.gdbm", STORETYPE_RESOURCE)
	if err != nil {
		return nil, err
	}
	rfs := resource.NewDbResource(store)
	if ms.poResource != nil {
		logg.InfoCtxf(ctx, "using poresource for menu and template")
		rfs.WithMenuGetter(ms.poResource.GetMenu)
		rfs.WithTemplateGetter(ms.poResource.GetTemplate)
	}
	return rfs, nil
}

func (ms *MenuStorageService) GetStateStore(ctx context.Context) (db.Db, error) {
	return ms.getOrCreateDb(ctx, "state.gdbm", STORETYPE_STATE)
}

func (ms *MenuStorageService) ensureDbDir(path string) error {
	err := os.MkdirAll(path, 0700)
	if err != nil {
		return fmt.Errorf("store dir create exited with error: %v\n", err)
	}
	return nil
}

// TODO: how to handle persister here?
func (ms *MenuStorageService) Close(ctx context.Context) error {
	var errs []error
	var haveErr bool
	for i := range(_STORETYPE_MAX) {
		err := ms.store[int8(i)].Close(ctx)
		if err != nil {
			haveErr = true
		}
		errs = append(errs, err) 
	}
	if haveErr {
		errStr := ""
		for i, err := range(errs) {
			errStr += fmt.Sprintf("(%d: %v)", i, err)
		}
		return errors.New(errStr)
	}
	return nil
}
