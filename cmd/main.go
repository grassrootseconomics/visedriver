package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	"git.defalsify.org/vise.git/asm"
	"git.defalsify.org/vise.git/db"
	fsdb "git.defalsify.org/vise.git/db/fs"
	gdbmdb "git.defalsify.org/vise.git/db/gdbm"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/logging"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
	"git.grassecon.net/urdt/ussd/internal/handlers/ussd"
)

var (
	logg      = logging.NewVanilla()
	scriptDir = path.Join("services", "registration")
)


type LocalHandler struct {
	sessionId string
}	

func NewLocalHandler() *LocalHandler {
	return &LocalHandler{
		sessionId: "",
	}
}

type RequestParser interface {
	GetSessionId(*http.Request) (string, error)
	GetInput(*http.Request) ([]byte, error)
}

type DefaultRequestParser struct {
}

func(rp *DefaultRequestParser) GetSessionId(rq *http.Request) (string, error) {
	v := rq.Header.Get("X-Vise-Session")
	if v == "" {
		return "", fmt.Errorf("no session found")
	}
	return v, nil
}

func(rp *DefaultRequestParser) GetInput(rq *http.Request) ([]byte, error) {
	defer rq.Body.Close()
	v, err := io.ReadAll(rq.Body)
	if err != nil {
		return nil, err
	}
	return v, nil
}

type DefaultSessionHandler struct {
	cfgTemplate engine.Config
	rp RequestParser
	rh *LocalHandler
	dbDir string
	resourceDir string
}

func NewDefaultSessionHandler(dbDir string, resourceDir string, rp RequestParser, outputSize uint32, flagCount uint32) *DefaultSessionHandler {
	rh := NewLocalHandler()
	return &DefaultSessionHandler{
		cfgTemplate: engine.Config{
			OutputSize: outputSize,
			Root: "root",
			FlagCount: flagCount,
		},
		rh: rh,
		rp: rp,
		dbDir: dbDir,
		resourceDir: resourceDir,
	}
}

func getParser(fp string, debug bool) (*asm.FlagParser, error) {
	flagParser := asm.NewFlagParser().WithDebug()
	_, err := flagParser.Load(fp)
	if err != nil {
		return nil, err
	}
	return flagParser, nil
}

func getHandler(appFlags *asm.FlagParser, rs *resource.DbResource, pe *persist.Persister, userdataStore db.Db) (*ussd.Handlers, error) {

	ussdHandlers, err := ussd.NewHandlers(appFlags, pe, userdataStore)
	if err != nil {
		return nil, err
	}
	rs.AddLocalFunc("select_language", ussdHandlers.SetLanguage)
	rs.AddLocalFunc("create_account", ussdHandlers.CreateAccount)
	rs.AddLocalFunc("save_pin", ussdHandlers.SavePin)
	rs.AddLocalFunc("verify_pin", ussdHandlers.VerifyPin)
	rs.AddLocalFunc("check_identifier", ussdHandlers.CheckIdentifier)
	rs.AddLocalFunc("check_account_status", ussdHandlers.CheckAccountStatus)
	rs.AddLocalFunc("authorize_account", ussdHandlers.Authorize)
	rs.AddLocalFunc("quit", ussdHandlers.Quit)
	rs.AddLocalFunc("check_balance", ussdHandlers.CheckBalance)
	rs.AddLocalFunc("validate_recipient", ussdHandlers.ValidateRecipient)
	rs.AddLocalFunc("transaction_reset", ussdHandlers.TransactionReset)
	rs.AddLocalFunc("max_amount", ussdHandlers.MaxAmount)
	rs.AddLocalFunc("validate_amount", ussdHandlers.ValidateAmount)
	rs.AddLocalFunc("reset_transaction_amount", ussdHandlers.ResetTransactionAmount)
	rs.AddLocalFunc("get_recipient", ussdHandlers.GetRecipient)
	rs.AddLocalFunc("get_sender", ussdHandlers.GetSender)
	rs.AddLocalFunc("get_amount", ussdHandlers.GetAmount)
	rs.AddLocalFunc("reset_incorrect", ussdHandlers.ResetIncorrectPin)
	rs.AddLocalFunc("save_firstname", ussdHandlers.SaveFirstname)
	rs.AddLocalFunc("save_familyname", ussdHandlers.SaveFamilyname)
	rs.AddLocalFunc("save_gender", ussdHandlers.SaveGender)
	rs.AddLocalFunc("save_location", ussdHandlers.SaveLocation)
	rs.AddLocalFunc("save_yob", ussdHandlers.SaveYob)
	rs.AddLocalFunc("save_offerings", ussdHandlers.SaveOfferings)
	rs.AddLocalFunc("quit_with_balance", ussdHandlers.QuitWithBalance)
	rs.AddLocalFunc("reset_account_authorized", ussdHandlers.ResetAccountAuthorized)
	rs.AddLocalFunc("reset_allow_update", ussdHandlers.ResetAllowUpdate)
	rs.AddLocalFunc("get_profile_info", ussdHandlers.GetProfileInfo)
	rs.AddLocalFunc("verify_yob", ussdHandlers.VerifyYob)
	rs.AddLocalFunc("reset_incorrect_date_format", ussdHandlers.ResetIncorrectYob)
	rs.AddLocalFunc("set_reset_single_edit", ussdHandlers.SetResetSingleEdit)
	rs.AddLocalFunc("initiate_transaction", ussdHandlers.InitiateTransaction)

	return ussdHandlers, nil
}

func(f *DefaultSessionHandler) getPersister(ctx context.Context) (*persist.Persister, error) {
	err := os.MkdirAll(f.dbDir, 0700)
	if err != nil {
		return nil, fmt.Errorf("state dir create exited with error: %v\n", err)
	}
	store := gdbmdb.NewGdbmDb()
	storeFile := path.Join(f.dbDir, "state.gdbm")
	store.Connect(ctx, storeFile)
	pr := persist.NewPersister(store)
	return pr, nil
}

func(f *DefaultSessionHandler) getUserdataDb(ctx context.Context) db.Db {
	store := gdbmdb.NewGdbmDb()
	storeFile := path.Join(f.dbDir, "userdata.gdbm")
	store.Connect(ctx, storeFile)

	return store
}

func(f *DefaultSessionHandler) getResource(ctx context.Context) (resource.Resource, error) {
	store := fsdb.NewFsDb()
	err := store.Connect(ctx, f.resourceDir)
	if err != nil {
		return nil, err
	}
	rfs := resource.NewDbResource(store)
	return rfs, nil
}

func(f *DefaultSessionHandler) getEngine(rs resource.Resource, pr *persist.Persister, sessionId string) *engine.DefaultEngine {
	cfg := f.cfgTemplate
	cfg.SessionId = sessionId
	en := engine.NewEngine(cfg, rs)
	en = en.WithPersister(pr)
	return en
}

func(f *DefaultSessionHandler) writeError(w http.ResponseWriter, code int, msg string, err error) {
	w.Header().Set("X-Vise", msg + ": " + err.Error())
	w.Header().Set("Content-Length", "0")
	w.WriteHeader(code)
	_, err = w.Write([]byte{})
	if err != nil {
		w.WriteHeader(500)
		w.Header().Set("X-Vise", err.Error())
	}
	return 
}

func(f *DefaultSessionHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var r bool

	sessionId, err := f.rp.GetSessionId(req)
	if err != nil {
		f.writeError(w, 400, "Session missing", err)
		return
	}

	input, err := f.rp.GetInput(req)
	if err != nil {
		f.writeError(w, 400, "Input read fail", err)
		return
	}

	ctx := req.Context()
	ctx = context.WithValue(ctx, "SessionId", sessionId)

	pfp := path.Join(scriptDir, "pp.csv")
	flagParser, err := getParser(pfp, true)
	if err != nil {
		f.writeError(w, 500, "flagParser failed with error:", err)
		return
	}

	rs, err := f.getResource(ctx)
	if err != nil {
		f.writeError(w, 500, "getResource failed with error:", err)
		return
	}

	pr, err := f.getPersister(ctx)
	if err != nil {
		f.writeError(w, 500, "getPersister failed with error:", err)
		return
	}

	store := f.getUserdataDb(ctx)
	
	dbResource, ok := rs.(*resource.DbResource)
	if !ok {
		f.writeError(w, 500, "getHandler exited with error:", err)
		return
	}

	hl, err := getHandler(flagParser, dbResource, pr, store)
	if err != nil {
		f.writeError(w, 500, "getHandler exited with error:", err)
		return
	}

	en := f.getEngine(rs, pr, sessionId)
	en = en.WithFirst(hl.Init)

	if len(input) == 0 {
		r, err = en.Init(ctx)
	} else {
		r, err = en.Exec(ctx, input)
	}
	if err != nil {
		f.writeError(w, 500, "Engine exec fail", err)
		return
	}

	// _, err = en.Init(ctx)
	// if err != nil {
	// 	f.writeError(w, 500, "Engine exec fail", err)
	// 	return
	// }

	// err = engine.Loop(ctx, en, os.Stdin, os.Stdout)
	// if err != nil {
	// 	f.writeError(w, 500, "Loop exec fail", err)
	// 	return
	// }

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/plain")
	_, err = en.WriteResult(ctx, w)
	if err != nil {
		f.writeError(w, 500, "Write result fail", err)
		return
	}

	_ = r
}

func main() {
	var host string
	var port string
	var dbDir string
	var resourceDir string
	var size uint
	var flagCount uint
	var sessionId string
	var debug bool
	flag.StringVar(&host, "h", "127.0.0.1", "http host")
	flag.StringVar(&port, "p", "7123", "http port")
	flag.StringVar(&sessionId, "session-id", "075xx2123", "session id")
	flag.StringVar(&dbDir, "dbdir", ".state", "database dir to read from")
	flag.StringVar(&resourceDir, "resourcedir", path.Join("services", "registration"), "resource dir")
	flag.BoolVar(&debug, "d", false, "use engine debug output")
	flag.UintVar(&size, "s", 160, "max size of output")
	flag.UintVar(&flagCount, "f", 16, "flag count")
	flag.Parse()

	logg.Infof("starting server", "dbdir", dbDir, "resourcedir", resourceDir, "outputsize", size, "flagCount", flagCount)

	rp := &DefaultRequestParser{}
	h := NewDefaultSessionHandler(dbDir, resourceDir, rp, uint32(size), uint32(flagCount))
	s := &http.Server{
		Addr: fmt.Sprintf("%s:%s", host, port),
		Handler: h,
	}

	err := s.ListenAndServe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %s", err)
		os.Exit(1)
	}
}
