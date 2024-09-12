package http

import (
	"io/ioutil"
	"net/http"
	"strconv"

	"git.defalsify.org/vise.git/db"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/logging"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"

	"git.grassecon.net/urdt/ussd/internal/handlers"
	"git.grassecon.net/urdt/ussd/internal/handlers/ussd"
	"git.grassecon.net/urdt/ussd/internal/storage"
)

var (
	logg = logging.NewVanilla().WithDomain("httpserver")
)

type DefaultRequestParser struct {
}


func(rp *DefaultRequestParser) GetSessionId(rq any) (string, error) {
	rqv, ok := rq.(*http.Request)
	if !ok {
		return "", handlers.ErrInvalidRequest
	}
	v := rqv.Header.Get("X-Vise-Session")
	if v == "" {
		return "", handlers.ErrSessionMissing
	}
	return v, nil
}

func(rp *DefaultRequestParser) GetInput(rq any) ([]byte, error) {
	rqv, ok := rq.(*http.Request)
	if !ok {
		return nil, handlers.ErrInvalidRequest
	}
	defer rqv.Body.Close()
	v, err := ioutil.ReadAll(rqv.Body)
	if err != nil {
		return nil, err
	}
	return v, nil
}

type SessionHandler struct {
	cfgTemplate engine.Config
	rp handlers.RequestParser
	rs resource.Resource
	hn *ussd.Handlers
	provider storage.StorageProvider
}

func NewSessionHandler(cfg engine.Config, rs resource.Resource, stateDb db.Db, userdataDb db.Db, rp handlers.RequestParser, hn *ussd.Handlers) *SessionHandler {
	return &SessionHandler{
		cfgTemplate: cfg,
		rs: rs,
		hn: hn,
		rp: rp,
		provider: storage.NewSimpleStorageProvider(stateDb, userdataDb),
	}
}

func(f *SessionHandler) writeError(w http.ResponseWriter, code int, err error) {
	s := err.Error()
	w.Header().Set("Content-Length", strconv.Itoa(len(s)))
	w.WriteHeader(code)
	_, err = w.Write([]byte{})
	if err != nil {
		logg.Errorf("error writing error!!", "err", err, "olderr", s)
		w.WriteHeader(500)
	}
	return 
}

func(f* SessionHandler) Shutdown() {
	err := f.provider.Close()
	if err != nil {
		logg.Errorf("handler shutdown error", "err", err)
	}
}

func(f *SessionHandler) GetEngine(cfg engine.Config, rs resource.Resource, pr *persist.Persister) engine.Engine {
	en := engine.NewEngine(cfg, rs)
	en = en.WithPersister(pr)
	return en
}

func(f *SessionHandler) Process(rqs handlers.RequestSession) (handlers.RequestSession, error) {
	var r bool
	var err error
	var ok bool
	
	logg.InfoCtxf(rqs.Ctx, "new request",  rqs)

	rqs.Storage, err = f.provider.Get(rqs.Config.SessionId)
	if err != nil {
		logg.ErrorCtxf(rqs.Ctx, "", "storage error", "err", err)
		return rqs, handlers.ErrStorage
	}

	f.hn = f.hn.WithPersister(rqs.Storage.Persister)
	eni := f.GetEngine(rqs.Config, f.rs, rqs.Storage.Persister)
	en, ok := eni.(*engine.DefaultEngine)
	if !ok {
		return rqs, handlers.ErrEngineType
	}
	en = en.WithFirst(f.hn.Init)
	if rqs.Config.EngineDebug {
		en = en.WithDebug(nil)
	}
	rqs.Engine = en

	r, err = rqs.Engine.Init(rqs.Ctx)
	if err != nil {
		return rqs, err
	}

	if r && len(rqs.Input) > 0 {
		r, err = rqs.Engine.Exec(rqs.Ctx, rqs.Input)
	}
	if err != nil {
		return rqs, err
	}

	_ = r
	return rqs, nil
}

func(f *SessionHandler) Output(rqs handlers.RequestSession) error {
	var err error
	_, err = rqs.Engine.WriteResult(rqs.Ctx, rqs.Writer)
	return err
}

func(f *SessionHandler) Reset(rqs handlers.RequestSession) error {
	defer f.provider.Put(rqs.Config.SessionId, rqs.Storage)
	return rqs.Engine.Finish()
}

func(f *SessionHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var code int
	var err error

	rqs := handlers.RequestSession{
		Ctx: req.Context(),
		Writer: w,
	}

	cfg := f.cfgTemplate
	cfg.SessionId, err = f.rp.GetSessionId(req)
	if err != nil {
		logg.ErrorCtxf(rqs.Ctx, "", "header processing error", err)
		f.writeError(w, 400, err)
	}
	rqs.Config = cfg
	rqs.Input, err = f.rp.GetInput(req)
	if err != nil {
		logg.ErrorCtxf(rqs.Ctx, "", "header processing error", err)
		f.writeError(w, 400, err)
		return
	}

	rqs, err = f.Process(rqs)
	switch err {
	case handlers.ErrStorage:
		code = 500
	case handlers.ErrEngineInit:
		code = 500
	case handlers.ErrEngineExec:
		code = 500
	default:
		code = 200
	}

	if code != 200 {
		f.writeError(w, 500, err)
		return
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/plain")
	err = f.Output(rqs)
	if err != nil {
		f.writeError(w, 500, err)
		return
	}

	err = f.Reset(rqs)
	if err != nil {
		f.writeError(w, 500, err)
		return
	}
}
