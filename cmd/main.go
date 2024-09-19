package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"

	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/logging"
	"git.defalsify.org/vise.git/resource"
	"git.grassecon.net/urdt/ussd/internal/handlers"
	"git.grassecon.net/urdt/ussd/internal/storage"
)

var (
	logg      = logging.NewVanilla()
	scriptDir = path.Join("services", "registration")
)

func main() {
	var dbDir string
	var size uint
	var sessionId string
	var debug bool
	flag.StringVar(&sessionId, "session-id", "075xx2123", "session id")
	flag.StringVar(&dbDir, "dbdir", ".state", "database dir to read from")
	flag.BoolVar(&debug, "d", false, "use engine debug output")
	flag.UintVar(&size, "s", 160, "max size of output")
	flag.Parse()

	logg.Infof("start command", "dbdir", dbDir, "outputsize", size)

	ctx := context.Background()
	ctx = context.WithValue(ctx, "SessionId", sessionId)
	pfp := path.Join(scriptDir, "pp.csv")

	cfg := engine.Config{
		Root:       "root",
		SessionId:  sessionId,
		OutputSize: uint32(size),
		FlagCount:  uint32(16),
	}

	menuStorageService := storage.MenuStorageService{}

	err := menuStorageService.EnsureDbDir(dbDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	rs, err := menuStorageService.GetResource(scriptDir, ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	pe, err := menuStorageService.GetPersister(dbDir, ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	userdatastore := menuStorageService.GetUserdataDb(dbDir, ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	dbResource, ok := rs.(*resource.DbResource)
	if !ok {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	lhs, err := handlers.NewLocalHandlerService(pfp, true, dbResource, pe, userdatastore, cfg, rs)

	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	hl, err := lhs.GetHandler()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	en := lhs.GetEngine()
	en = en.WithFirst(hl.Init)
	if debug {
		en = en.WithDebug(nil)
	}

	_, err = en.Init(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "engine init exited with error: %v\n", err)
		os.Exit(1)
	}

	err = engine.Loop(ctx, en, os.Stdin, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "loop exited with error: %v\n", err)
		os.Exit(1)
	}
}
