package common

import (
	"git.grassecon.net/urdt/ussd/internal/handlers"
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
