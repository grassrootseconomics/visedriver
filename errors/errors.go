package errors

import (
	"git.grassecon.net/grassrootseconomics/visedriver/internal/handlers"
)

var (
	ErrInvalidRequest = handlers.ErrInvalidRequest
	ErrSessionMissing = handlers.ErrSessionMissing
	ErrInvalidInput = handlers.ErrInvalidInput
	ErrStorage = handlers.ErrStorage
	ErrEngineType = handlers.ErrEngineType
	ErrEngineInit = handlers.ErrEngineInit
	ErrEngineExec = handlers.ErrEngineExec
)
