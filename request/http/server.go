package http

import (
	"net/http"
	"strconv"

	"git.defalsify.org/vise.git/logging"
	"git.grassecon.net/grassrootseconomics/visedriver/request"
	"git.grassecon.net/grassrootseconomics/visedriver/errors"
)

var (
	logg = logging.NewVanilla().WithDomain("visedriver.http.session")
)

// HTTPRequestHandler implements the session handler for HTTP
type HTTPRequestHandler struct {
	request.RequestHandler
}

func (f *HTTPRequestHandler) WriteError(w http.ResponseWriter, code int, err error) {
	s := err.Error()
	w.Header().Set("Content-Length", strconv.Itoa(len(s)))
	w.WriteHeader(code)
	_, err = w.Write([]byte(s))
	if err != nil {
		logg.Errorf("error writing error!!", "err", err, "olderr", s)
		w.WriteHeader(500)
	}
}

func NewHTTPRequestHandler(h request.RequestHandler) *HTTPRequestHandler {
	return &HTTPRequestHandler{
		RequestHandler: h,
	}
}

func (hh *HTTPRequestHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var code int
	var err error
	var perr error

	rqs := request.RequestSession{
		Ctx:    req.Context(),
		Writer: w,
	}

	rp := hh.GetRequestParser()
	cfg := hh.GetConfig()
	cfg.SessionId, err = rp.GetSessionId(req.Context(), req)
	if err != nil {
		logg.ErrorCtxf(rqs.Ctx, "", "header processing error", err)
		hh.WriteError(w, 400, err)
	}
	rqs.Config = cfg
	rqs.Input, err = rp.GetInput(req)
	if err != nil {
		logg.ErrorCtxf(rqs.Ctx, "", "header processing error", err)
		hh.WriteError(w, 400, err)
		return
	}

	rqs, err = hh.Process(rqs)
	switch err {
	case errors.ErrStorage:
		code = 500
	case errors.ErrEngineInit:
		code = 500
	case errors.ErrEngineExec:
		code = 500
	default:
		code = 200
	}

	if code != 200 {
		hh.WriteError(w, 500, err)
		return
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/plain")
	rqs, err = hh.Output(rqs)
	rqs, perr = hh.Reset(rqs.Ctx, rqs)
	if err != nil {
		hh.WriteError(w, 500, err)
		return
	}
	if perr != nil {
		hh.WriteError(w, 500, perr)
		return
	}
}
