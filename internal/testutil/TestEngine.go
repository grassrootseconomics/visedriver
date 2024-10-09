package testutil

import (
	"context"
	"fmt"
	"os"
	"path"
	"time"

	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/logging"
	"git.defalsify.org/vise.git/resource"
	"git.grassecon.net/urdt/ussd/internal/handlers"
	"git.grassecon.net/urdt/ussd/internal/handlers/server"
	"git.grassecon.net/urdt/ussd/internal/storage"
	testdataloader "github.com/peteole/testdata-loader"
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
		FlagCount:  uint32(16),
	}

	dbDir := ".test_state"
	resourceDir := scriptDir
	menuStorageService := storage.NewMenuStorageService(dbDir, resourceDir)

	err := menuStorageService.EnsureDbDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	rs, err := menuStorageService.GetResource(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	pe, err := menuStorageService.GetPersister(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	userDataStore, err := menuStorageService.GetUserdataDb(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	dbResource, ok := rs.(*resource.DbResource)
	if !ok {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	lhs, err := handlers.NewLocalHandlerService(pfp, true, dbResource, cfg, rs)
	lhs.SetDataStore(&userDataStore)
	lhs.SetPersister(pe)

	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	switch AccountService.(type) {
	case *server.MockAccountService:
		go func() {
			eventChannel <- false
		}()
	case *server.AccountService:
		go func() {
			time.Sleep(5 * time.Second) // Wait for 5 seconds
			eventChannel <- true
		}()
	default:
		panic("Unknown account service type")
	}

	hl, err := lhs.GetHandler(AccountService)
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

	//en = en.WithDebug(nil)
	return en, cleanFn, eventChannel
}
