package entry

import (
	"context"

	"github.com/grassrootseconomics/go-vise/persist"
	"github.com/grassrootseconomics/go-vise/resource"
)

type EntryHandler interface {
	Init(context.Context, string, []byte) (resource.Result, error) // HandlerFunc
	Exit()
	SetPersister(*persist.Persister)
}
