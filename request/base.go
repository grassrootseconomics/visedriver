package request

import (
	"context"

	"git.defalsify.org/vise.git/db"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
	"git.grassecon.net/grassrootseconomics/visedriver/storage"
	"git.grassecon.net/grassrootseconomics/visedriver/errors"
	"git.grassecon.net/grassrootseconomics/visedriver/entry"
)

type BaseRequestHandler struct {
	cfgTemplate engine.Config
	rp RequestParser
	rs resource.Resource
	hn entry.EntryHandler
	provider storage.StorageProvider
}

//func NewBaseRequestHandler(cfg engine.Config, rs resource.Resource, stateDb db.Db, userdataDb db.Db, rp request.RequestParser, hn *handlers.Handlers) *BaseRequestHandler {
func NewBaseRequestHandler(cfg engine.Config, rs resource.Resource, stateDb db.Db, userdataDb db.Db, rp RequestParser, hn entry.EntryHandler) *BaseRequestHandler {
	return &BaseRequestHandler{
		cfgTemplate: cfg,
		rs:          rs,
		hn:          hn,
		rp:          rp,
		provider:    storage.NewSimpleStorageProvider(stateDb, userdataDb),
	}
}

func (f *BaseRequestHandler) Shutdown(ctx context.Context) {
	err := f.provider.Close(ctx)
	if err != nil {
		logg.Errorf("handler shutdown error", "err", err)
	}
}

func (f *BaseRequestHandler) GetEngine(cfg engine.Config, rs resource.Resource, pr *persist.Persister) engine.Engine {
	en := engine.NewEngine(cfg, rs)
	en = en.WithPersister(pr)
	return en
}

func(f *BaseRequestHandler) Process(rqs RequestSession) (RequestSession, error) {
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

func(f *BaseRequestHandler) Output(rqs RequestSession) (RequestSession,  error) {
	var err error
	_, err = rqs.Engine.Flush(rqs.Ctx, rqs.Writer)
	return rqs, err
}

func(f *BaseRequestHandler) Reset(ctx context.Context, rqs RequestSession) (RequestSession, error) {
	defer f.provider.Put(rqs.Config.SessionId, rqs.Storage)
	return rqs, rqs.Engine.Finish(ctx)
}

func (f *BaseRequestHandler) GetConfig() engine.Config {
	return f.cfgTemplate
}

func(f *BaseRequestHandler) GetRequestParser() RequestParser {
	return f.rp
}
