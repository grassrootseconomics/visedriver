package entry

import (
	"context"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/persist"
)

type EntryHandler interface {
	Init(context.Context, string, []byte) (resource.Result, error) // HandlerFunc
	Exit()
	SetPersister(*persist.Persister)
}
