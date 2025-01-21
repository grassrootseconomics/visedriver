package entry

import (
	"context"

	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
)

type EntryHandler interface {
	Init(context.Context, string, []byte) (resource.Result, error) // HandlerFunc
	Exit()
	SetPersister(*persist.Persister)
}
