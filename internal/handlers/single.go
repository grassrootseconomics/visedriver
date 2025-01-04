package handlers

import (
	"errors"

	"git.defalsify.org/vise.git/logging"
)

var (
	ErrInvalidRequest = errors.New("invalid request for context")
	ErrSessionMissing = errors.New("missing session")
	ErrInvalidInput = errors.New("invalid input")
	ErrStorage = errors.New("storage retrieval fail")
	ErrEngineType = errors.New("incompatible engine")
	ErrEngineInit = errors.New("engine init fail")
	ErrEngineExec = errors.New("engine exec fail")
)

var (
	logg = logging.NewVanilla().WithDomain("handlers")
)

