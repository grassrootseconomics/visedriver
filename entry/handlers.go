package entry

import (
	"git.defalsify.org/vise.git/persist"
)

type EntryHandler interface {
	Init(context.Context, string, []byte) (*resource.Result, error) // HandlerFunc
	Exit()
}
