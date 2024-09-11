package http

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"git.defalsify.org/vise.git/db"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/logging"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
)

var (
	logg = logging.NewVanilla().WithDomain("httpserver")
)

type RequestParser interface {
	GetSessionId(rq *http.Request) (string, error)
	GetInput(rq *http.Request) ([]byte, error)
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
	v, err := ioutil.ReadAll(rq.Body)
	if err != nil {
		return nil, err
	}
	return v, nil
}

type SessionHandler struct {
	cfgTemplate engine.Config
	rp RequestParser
	rs resource.Resource
	first resource.EntryFunc
	provider StorageProvider
}

func NewSessionHandler(cfg engine.Config, rs resource.Resource, stateDb db.Db, userdataDb db.Db, rp RequestParser, first resource.EntryFunc) *SessionHandler {
	return &SessionHandler{
		cfgTemplate: cfg,
		rs: rs,
		first: first,
		rp: rp,
		provider: NewSimpleStorageProvider(stateDb, userdataDb),
	}
}

func(f *SessionHandler) writeError(w http.ResponseWriter, code int, msg string, err error) {
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

func(f* SessionHandler) Shutdown() {
	err := f.provider.Close()
	if err != nil {
		logg.Errorf("handler shutdown error", "err", err)
	}
}

func(f *SessionHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
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
	cfg := f.cfgTemplate
	cfg.SessionId = sessionId

	logg.InfoCtxf(ctx, "new request",  "session", cfg.SessionId, "input", input)

	storage, err := f.provider.Get(cfg.SessionId)
	if err != nil {
		f.writeError(w, 500, "Storage retrieval fail", err)
		return
	}
	defer f.provider.Put(cfg.SessionId, storage)
	en := getEngine(cfg, f.rs, storage.Persister)
	en = en.WithFirst(f.first)
	if cfg.EngineDebug {
		en = en.WithDebug(nil)
	}

	r, err = en.Init(ctx)
	if err != nil {
		f.writeError(w, 500, "Engine init fail", err)
		return
	}
	if r && len(input) > 0 {
		r, err = en.Exec(ctx, input)
	}
	if err != nil {
		f.writeError(w, 500, "Engine exec fail", err)
		return
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/plain")
	_, err = en.WriteResult(ctx, w)
	if err != nil {
		f.writeError(w, 500, "Write result fail", err)
		return
	}
	err = en.Finish()
	if err != nil {
		f.writeError(w, 500, "Engine finish fail", err)
		return
	}

	_ = r
}

func getEngine(cfg engine.Config, rs resource.Resource, pr *persist.Persister) *engine.DefaultEngine {
	en := engine.NewEngine(cfg, rs)
	en = en.WithPersister(pr)
	return en
}
