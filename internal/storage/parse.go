package storage

const (
	DBTYPE_MEM = iota
	DBTYPE_GDBM
	DBTYPE_POSTGRES
)

type connData struct {
	typ int
	str string
}

func toConnData(s string) connData {
	var o connData

	if s == "" {
		return o
	}
	return o
}
