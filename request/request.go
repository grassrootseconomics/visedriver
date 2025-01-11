package request

import (
	"context"
	"io"

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
