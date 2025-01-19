package httpmocks

import (
	"context"

	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
	"git.grassecon.net/grassrootseconomics/visedriver/request"
)

// MockRequestHandler implements request.RequestHandler interface for testing
type MockRequestHandler struct {
	ProcessFunc          func(request.RequestSession) (request.RequestSession, error)
	GetConfigFunc        func() engine.Config
	GetEngineFunc        func(cfg engine.Config, rs resource.Resource, pe *persist.Persister) engine.Engine
	OutputFunc           func(rs request.RequestSession) (request.RequestSession, error)
	ResetFunc            func(ctx context.Context, rs request.RequestSession) (request.RequestSession, error)
	ShutdownFunc         func(ctx context.Context)
	GetRequestParserFunc func() request.RequestParser
}

func (m *MockRequestHandler) Process(rqs request.RequestSession) (request.RequestSession, error) {
	return m.ProcessFunc(rqs)
}

func (m *MockRequestHandler) GetConfig() engine.Config {
	return m.GetConfigFunc()
}

func (m *MockRequestHandler) GetEngine(cfg engine.Config, rs resource.Resource, pe *persist.Persister) engine.Engine {
	return m.GetEngineFunc(cfg, rs, pe)
}

func (m *MockRequestHandler) Output(rs request.RequestSession) (request.RequestSession, error) {
	return m.OutputFunc(rs)
}

func (m *MockRequestHandler) Reset(ctx context.Context, rs request.RequestSession) (request.RequestSession, error) {
	return m.ResetFunc(ctx, rs)
}

func (m *MockRequestHandler) Shutdown(ctx context.Context) {
	m.ShutdownFunc(ctx)
}

func (m *MockRequestHandler) GetRequestParser() request.RequestParser {
	return m.GetRequestParserFunc()
}
