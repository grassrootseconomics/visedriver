package testutil

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/logging"
	"git.defalsify.org/vise.git/resource"
	"git.grassecon.net/grassrootseconomics/visedriver/handlers"
	"git.grassecon.net/grassrootseconomics/visedriver/storage"
	"git.grassecon.net/grassrootseconomics/visedriver/internal/testutil/testservice"
	"git.grassecon.net/grassrootseconomics/visedriver/internal/testutil/testtag"
	testdataloader "github.com/peteole/testdata-loader"
	"git.grassecon.net/grassrootseconomics/visedriver/remote"
)

var (
	baseDir   = testdataloader.GetBasePath()
	logg      = logging.NewVanilla()
	scriptDir = path.Join(baseDir, "services", "registration")
)

func TestEngine(sessionId string) (engine.Engine, func(), chan bool) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "SessionId", sessionId)
	pfp := path.Join(scriptDir, "pp.csv")

	var eventChannel = make(chan bool)

	cfg := engine.Config{
		Root:       "root",
		SessionId:  sessionId,
		OutputSize: uint32(160),
		FlagCount:  uint32(128),
	}

	connStr, err := filepath.Abs(".test_state/state.gdbm")
	if err != nil {
		fmt.Fprintf(os.Stderr, "connstr err: %v", err)
		os.Exit(1)
	}
	conn, err := storage.ToConnData(connStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connstr parse err: %v", err)
		os.Exit(1)
	}
	resourceDir := scriptDir
	menuStorageService := storage.NewMenuStorageService(conn, resourceDir)

	rs, err := menuStorageService.GetResource(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "resource error: %v", err)
		os.Exit(1)
	}

	pe, err := menuStorageService.GetPersister(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "persister error: %v", err)
		os.Exit(1)
	}

	userDataStore, err := menuStorageService.GetUserdataDb(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "userdb error: %v", err)
		os.Exit(1)
	}

	dbResource, ok := rs.(*resource.DbResource)
	if !ok {
		fmt.Fprintf(os.Stderr, "dbresource cast error")
		os.Exit(1)
	}

	lhs, err := handlers.NewLocalHandlerService(ctx, pfp, true, dbResource, cfg, rs)
	lhs.SetDataStore(&userDataStore)
	lhs.SetPersister(pe)

	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	if testtag.AccountService == nil {
		testtag.AccountService = &remote.AccountService{}
	}

	switch testtag.AccountService.(type) {
	case *testservice.TestAccountService:
		go func() {
			eventChannel <- false
		}()
	case *remote.AccountService:
		go func() {
			time.Sleep(5 * time.Second) // Wait for 5 seconds
			eventChannel <- true
		}()
	default:
		panic("Unknown account service type")
	}

	hl, err := lhs.GetHandler(testtag.AccountService)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	en := lhs.GetEngine()
	en = en.WithFirst(hl.Init)
	cleanFn := func() {
		err := en.Finish()
		if err != nil {
			logg.Errorf(err.Error())
		}

		err = menuStorageService.Close()
		if err != nil {
			logg.Errorf(err.Error())
		}
		logg.Infof("testengine storage closed")
	}
	return en, cleanFn, eventChannel
}
