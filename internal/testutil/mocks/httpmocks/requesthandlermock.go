package httpmocks

import (
	"git.defalsify.org/vise.git/engine"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
	"git.grassecon.net/urdt/ussd/request"
)

// MockRequestHandler implements request.RequestHandler interface for testing
type MockRequestHandler struct {
	ProcessFunc          func(request.RequestSession) (request.RequestSession, error)
	GetConfigFunc        func() engine.Config
	GetEngineFunc        func(cfg engine.Config, rs resource.Resource, pe *persist.Persister) engine.Engine
	OutputFunc           func(rs request.RequestSession) (request.RequestSession, error)
	ResetFunc            func(rs request.RequestSession) (request.RequestSession, error)
	ShutdownFunc         func()
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

func (m *MockRequestHandler) Reset(rs request.RequestSession) (request.RequestSession, error) {
	return m.ResetFunc(rs)
}

func (m *MockRequestHandler) Shutdown() {
	m.ShutdownFunc()
}

func (m *MockRequestHandler) GetRequestParser() request.RequestParser {
	return m.GetRequestParserFunc()
}
