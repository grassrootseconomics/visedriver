package request

import (
	"context"
	"io"

	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/logging"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
	"git.grassecon.net/grassrootseconomics/visedriver/storage"
)

var (
	logg = logging.NewVanilla().WithDomain("visedriver.request")
)

type RequestSession struct {
	Ctx      context.Context
	Config   engine.Config
	Engine   engine.Engine
	Input    []byte
	Storage  *storage.Storage
	Writer   io.Writer
	Continue bool
}

// TODO: seems like can remove this.
type RequestParser interface {
	GetSessionId(context.Context, any) (string, error)
	GetInput(any) ([]byte, error)
}

type RequestHandler interface {
	GetConfig() engine.Config
	GetRequestParser() RequestParser
	GetEngine(engine.Config, resource.Resource, *persist.Persister) engine.Engine
	Process(RequestSession) (RequestSession, error)
	Output(RequestSession) (RequestSession, error)
	Reset(context.Context, RequestSession) (RequestSession, error)
	Shutdown(ctx context.Context)
}
