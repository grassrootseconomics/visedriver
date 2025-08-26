package http

import (
	"net/http"
	"strconv"

	"git.defalsify.org/vise.git/logging"
	"git.grassecon.net/grassrootseconomics/visedriver/errors"
	"git.grassecon.net/grassrootseconomics/visedriver/request"
	"github.com/uptrace/bunrouter"
)

var (
	logg = logging.NewVanilla().WithDomain("visedriver.http.session")
)

// HTTPRequestHandler implements the session handler for HTTP
type HTTPRequestHandler struct {
	request.RequestHandler
	router *bunrouter.Router
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
	handler := &HTTPRequestHandler{
		RequestHandler: h,
		router:         bunrouter.New(),
	}

	// Route all requests to existing vise logic
	handler.router.GET("/*path", handler.handleViseRequest)
	handler.router.POST("/*path", handler.handleViseRequest)
	handler.router.PUT("/*path", handler.handleViseRequest)
	handler.router.DELETE("/*path", handler.handleViseRequest)

	return handler
}

func (hh *HTTPRequestHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	hh.router.ServeHTTP(w, req)
}

func (hh *HTTPRequestHandler) handleViseRequest(w http.ResponseWriter, req bunrouter.Request) error {
	var code int
	var err error
	var perr error

	rqs := request.RequestSession{
		Ctx:    req.Context(),
		Writer: w,
	}

	rp := hh.GetRequestParser()
	cfg := hh.GetConfig()
	cfg.SessionId, err = rp.GetSessionId(req.Context(), req.Request)
	if err != nil {
		logg.ErrorCtxf(rqs.Ctx, "", "header processing error", err)
		hh.WriteError(w, 400, err)
		return err
	}
	rqs.Config = cfg
	rqs.Input, err = rp.GetInput(req.Request)
	if err != nil {
		logg.ErrorCtxf(rqs.Ctx, "", "header processing error", err)
		hh.WriteError(w, 400, err)
		return err
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
		return err
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/plain")
	rqs, err = hh.Output(rqs)
	rqs, perr = hh.Reset(rqs.Ctx, rqs)
	if err != nil {
		hh.WriteError(w, 500, err)
		return err
	}
	if perr != nil {
		hh.WriteError(w, 500, perr)
		return perr
	}
	return nil
}
