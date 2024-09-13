package http

import (
	"io/ioutil"
	"net/http"
	"strconv"

	"git.defalsify.org/vise.git/logging"

	"git.grassecon.net/urdt/ussd/internal/handlers"
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



type SessionHandlerOption func(*SessionHandler)

func WithAtOutput() SessionHandlerOption {
	return func(sh *SessionHandler) {
		sh.useAtOutput = true
	}
}

type SessionHandler struct {
	handlers.RequestHandler
	useAtOutput bool
}

func ToSessionHandler(h handlers.RequestHandler, opts ...SessionHandlerOption) *SessionHandler {
	sh := &SessionHandler{
		RequestHandler: h,
		useAtOutput: false,
	}
	for _, opt := range opts {
		opt(sh)
	}
	return sh
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

func(f *SessionHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var code int
	var err error

	rqs := handlers.RequestSession{
		Ctx: req.Context(),
		Writer: w,
	}

	rp := f.GetRequestParser()
	cfg := f.GetConfig()
	cfg.SessionId, err = rp.GetSessionId(req)
	if err != nil {
		logg.ErrorCtxf(rqs.Ctx, "", "header processing error", err)
		f.writeError(w, 400, err)
	}
	rqs.Config = cfg
	rqs.Input, err = rp.GetInput(req)
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
	if f.useAtOutput {
		rqs, err = f.AtOutput(rqs)
	} else {
		rqs, err = f.Output(rqs)
	}
	if err != nil {
		f.writeError(w, 500, err)
		return
	}

	rqs, err = f.Reset(rqs)
	if err != nil {
		f.writeError(w, 500, err)
		return
	}
}
