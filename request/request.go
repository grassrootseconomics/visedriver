package request

import (
	"context"
	"io"
	"net/http"
	"strconv"

	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/logging"
	"git.grassecon.net/grassrootseconomics/visedriver/storage"
)

var (
	logg = logging.NewVanilla().WithDomain("visedriver.request")
)

type RequestSession struct {
	Ctx context.Context
	Config engine.Config
	Engine engine.Engine
	Input []byte
	Storage *storage.Storage
	Writer io.Writer
	Continue bool
}


// TODO: seems like can remove this.
type RequestParser interface {
	GetSessionId(ctx context.Context, rq any) (string, error)
	GetInput(rq any) ([]byte, error)
}

type RequestHandler interface {
	GetConfig() engine.Config
	GetRequestParser() RequestParser
	GetEngine(cfg engine.Config, rs resource.Resource, pe *persist.Persister) engine.Engine 
	Process(rs RequestSession) (RequestSession, error)
	Output(rs RequestSession) (RequestSession, error)
	Reset(rs RequestSession) (RequestSession, error)
	Shutdown()
}
type SessionHandler struct {
	RequestHandler
}

func (f *SessionHandler) WriteError(w http.ResponseWriter, code int, err error) {
	s := err.Error()
	w.Header().Set("Content-Length", strconv.Itoa(len(s)))
	w.WriteHeader(code)
	_, err = w.Write([]byte(s))
	if err != nil {
		logg.Errorf("error writing error!!", "err", err, "olderr", s)
		w.WriteHeader(500)
	}
}

func ToSessionHandler(h RequestHandler) *SessionHandler {
	return &SessionHandler{
		RequestHandler: h,
	}
}
