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

func (rp *DefaultRequestParser) GetSessionId(rq any) (string, error) {
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

func (rp *DefaultRequestParser) GetInput(rq any) ([]byte, error) {
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
	handlers.RequestHandler
}

func ToSessionHandler(h handlers.RequestHandler) *SessionHandler {
	return &SessionHandler{
		RequestHandler: h,
	}
}

func (f *SessionHandler) writeError(w http.ResponseWriter, code int, err error) {
	s := err.Error()
	w.Header().Set("Content-Length", strconv.Itoa(len(s)))
	w.WriteHeader(code)
	_, err = w.Write([]byte(s))
	if err != nil {
		logg.Errorf("error writing error!!", "err", err, "olderr", s)
		w.WriteHeader(500)
	}
}

func (f *SessionHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var code int
	var err error
	var perr error

	rqs := handlers.RequestSession{
		Ctx:    req.Context(),
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
	rqs, err = f.Output(rqs)
	rqs, perr = f.Reset(rqs)
	if err != nil {
		f.writeError(w, 500, err)
		return
	}
	if perr != nil {
		f.writeError(w, 500, perr)
		return
	}
}
