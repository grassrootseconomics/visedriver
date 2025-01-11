package session

import (
	"git.defalsify.org/vise.git/db"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/logging"
	"git.grassecon.net/grassrootseconomics/visedriver/request"
	"git.grassecon.net/grassrootseconomics/visedriver/storage"
	"git.grassecon.net/grassrootseconomics/visedriver/errors"
	"git.grassecon.net/grassrootseconomics/visedriver/entry"
)

var (
	logg = logging.NewVanilla().WithDomain("visedriver.session")
)

type BaseSessionHandler struct {
	cfgTemplate engine.Config
	rp request.RequestParser
	rs resource.Resource
	hn entry.EntryHandler
	provider storage.StorageProvider
}

//func NewBaseSessionHandler(cfg engine.Config, rs resource.Resource, stateDb db.Db, userdataDb db.Db, rp request.RequestParser, hn *handlers.Handlers) *BaseSessionHandler {
func NewBaseSessionHandler(cfg engine.Config, rs resource.Resource, stateDb db.Db, userdataDb db.Db, rp request.RequestParser, hn entry.EntryHandler) *BaseSessionHandler {
	return &BaseSessionHandler{
		cfgTemplate: cfg,
		rs:          rs,
		hn:          hn,
		rp:          rp,
		provider:    storage.NewSimpleStorageProvider(stateDb, userdataDb),
	}
}

func (f *BaseSessionHandler) Shutdown() {
	err := f.provider.Close()
	if err != nil {
		logg.Errorf("handler shutdown error", "err", err)
	}
}

func (f *BaseSessionHandler) GetEngine(cfg engine.Config, rs resource.Resource, pr *persist.Persister) engine.Engine {
	en := engine.NewEngine(cfg, rs)
	en = en.WithPersister(pr)
	return en
}

func(f *BaseSessionHandler) Process(rqs request.RequestSession) (request.RequestSession, error) {
	var r bool
	var err error
	var ok bool

	logg.InfoCtxf(rqs.Ctx, "new request", "data", rqs)

	rqs.Storage, err = f.provider.Get(rqs.Config.SessionId)
	if err != nil {
		logg.ErrorCtxf(rqs.Ctx, "", "storage get error", err)
		return rqs, errors.ErrStorage
	}

	//f.hn = f.hn.WithPersister(rqs.Storage.Persister)
	f.hn.SetPersister(rqs.Storage.Persister)
	defer func() {
		f.hn.Exit()
	}()
	eni := f.GetEngine(rqs.Config, f.rs, rqs.Storage.Persister)
	en, ok := eni.(*engine.DefaultEngine)
	if !ok {
		perr := f.provider.Put(rqs.Config.SessionId, rqs.Storage)
		rqs.Storage = nil
		if perr != nil {
			logg.ErrorCtxf(rqs.Ctx, "", "storage put error", perr)
		}
		return rqs, errors.ErrEngineType
	}
	en = en.WithFirst(f.hn.Init)
	if rqs.Config.EngineDebug {
		en = en.WithDebug(nil)
	}
	rqs.Engine = en

	r, err = rqs.Engine.Exec(rqs.Ctx, rqs.Input)
	if err != nil {
		perr := f.provider.Put(rqs.Config.SessionId, rqs.Storage)
		rqs.Storage = nil
		if perr != nil {
			logg.ErrorCtxf(rqs.Ctx, "", "storage put error", perr)
		}
		return rqs, err
	}

	rqs.Continue = r
	return rqs, nil
}

func(f *BaseSessionHandler) Output(rqs request.RequestSession) (request.RequestSession,  error) {
	var err error
	_, err = rqs.Engine.Flush(rqs.Ctx, rqs.Writer)
	return rqs, err
}

func(f *BaseSessionHandler) Reset(rqs request.RequestSession) (request.RequestSession, error) {
	defer f.provider.Put(rqs.Config.SessionId, rqs.Storage)
	return rqs, rqs.Engine.Finish()
}

func (f *BaseSessionHandler) GetConfig() engine.Config {
	return f.cfgTemplate
}

func(f *BaseSessionHandler) GetRequestParser() request.RequestParser {
	return f.rp
}
