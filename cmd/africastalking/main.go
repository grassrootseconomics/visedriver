package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strconv"
	"syscall"

	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/logging"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/lang"

	"git.grassecon.net/urdt/ussd/config"
	"git.grassecon.net/urdt/ussd/initializers"
	"git.grassecon.net/urdt/ussd/internal/handlers"
	"git.grassecon.net/urdt/ussd/internal/http/at"
	httpserver "git.grassecon.net/urdt/ussd/internal/http/at"
	"git.grassecon.net/urdt/ussd/internal/storage"
	"git.grassecon.net/urdt/ussd/remote"
	"git.grassecon.net/urdt/ussd/internal/args"
)

var (
	logg          = logging.NewVanilla().WithDomain("AfricasTalking").WithContextKey("at-session-id")
	scriptDir     = path.Join("services", "registration")
	build         = "dev"
	menuSeparator = ": "
)

func init() {
	initializers.LoadEnvVariables()
}
func main() {
	config.LoadConfig()

	var dbDir string
	var resourceDir string
	var size uint
	var database string
	var engineDebug bool
	var host string
	var port uint
	var gettextDir string
	var langs args.LangVar
	flag.StringVar(&dbDir, "dbdir", ".state", "database dir to read from")
	flag.StringVar(&resourceDir, "resourcedir", path.Join("services", "registration"), "resource dir")
	flag.StringVar(&database, "db", "gdbm", "database to be used")
	flag.BoolVar(&engineDebug, "d", false, "use engine debug output")
	flag.UintVar(&size, "s", 160, "max size of output")
	flag.StringVar(&host, "h", initializers.GetEnv("HOST", "127.0.0.1"), "http host")
	flag.UintVar(&port, "p", initializers.GetEnvUint("PORT", 7123), "http port")
	flag.StringVar(&gettextDir, "gettext", "", "use gettext translations from given directory")
	flag.Var(&langs, "language", "add symbol resolution for language")
	flag.Parse()

	logg.Infof("start command", "build", build, "dbdir", dbDir, "resourcedir", resourceDir, "outputsize", size)

	ctx := context.Background()
	ctx = context.WithValue(ctx, "Database", database)
	ln, err := lang.LanguageFromCode(config.DefaultLanguage)
	if err != nil {
		fmt.Fprintf(os.Stderr, "default language set error: %v", err)
		os.Exit(1)
	}
	ctx = context.WithValue(ctx, "Language", ln)

	pfp := path.Join(scriptDir, "pp.csv")

	cfg := engine.Config{
		Root:          "root",
		OutputSize:    uint32(size),
		FlagCount:     uint32(128),
		MenuSeparator: menuSeparator,
	}

	if engineDebug {
		cfg.EngineDebug = true
	}

	menuStorageService := storage.NewMenuStorageService(dbDir, resourceDir)
	rs, err := menuStorageService.GetResource(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	err = menuStorageService.EnsureDbDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	userdataStore, err := menuStorageService.GetUserdataDb(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	defer userdataStore.Close()

	dbResource, ok := rs.(*resource.DbResource)
	if !ok {
		os.Exit(1)
	}

	lhs, err := handlers.NewLocalHandlerService(ctx, pfp, true, dbResource, cfg, rs)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	lhs.SetDataStore(&userdataStore)

	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	accountService := remote.AccountService{}
	hl, err := lhs.GetHandler(&accountService)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	stateStore, err := menuStorageService.GetStateStore(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	defer stateStore.Close()

	rp := &at.ATRequestParser{
		Context: ctx,
	}
	bsh := handlers.NewBaseSessionHandler(cfg, rs, stateStore, userdataStore, rp, hl)
	sh := httpserver.NewATSessionHandler(bsh)

	mux := http.NewServeMux()
	mux.Handle(initializers.GetEnv("AT_ENDPOINT", "/"), sh)

	s := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", host, strconv.Itoa(int(port))),
		Handler: mux,
	}
	s.RegisterOnShutdown(sh.Shutdown)

	cint := make(chan os.Signal)
	cterm := make(chan os.Signal)
	signal.Notify(cint, os.Interrupt, syscall.SIGINT)
	signal.Notify(cterm, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case _ = <-cint:
		case _ = <-cterm:
		}
		s.Shutdown(ctx)
	}()
	err = s.ListenAndServe()
	if err != nil {
		logg.Infof("Server closed with error", "err", err)
	}
}
