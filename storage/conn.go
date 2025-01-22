package storage

import (
	"fmt"
	"net/url"
)

type DbMode uint8

const (
	DBTYPE_NONE = iota
	DBTYPE_MEM
	DBTYPE_FS
	DBTYPE_GDBM
	DBTYPE_POSTGRES
)

const (
	DBMODE_BINARY DbMode = iota
	DBMODE_TEXT
)

const (
	STORETYPE_STATE = iota
	STORETYPE_RESOURCE
	STORETYPE_USER
	_STORETYPE_MAX
)

type Conns map[int8]ConnData

func NewConns() Conns {
	c := make(Conns)
	return c
}

func (c Conns) Set(conn ConnData, typ int8) {
	if typ < 0 || typ >= _STORETYPE_MAX {
		panic(fmt.Errorf("invalid store type: %d", typ))
	}
	c[typ] = conn
}

func (c Conns) Have(conn *ConnData) int8 {
	for i := range _STORETYPE_MAX {
		ii := int8(i)
		v, ok := c[ii]
		if !ok {
			continue
		}
		if v.String() == conn.String() {
			return ii
		}
	}
	return -1
}

type ConnData struct {
	typ    int
	str    string
	domain string
	mode DbMode
}

func (cd *ConnData) DbType() int {
	return cd.typ
}

func (cd *ConnData) String() string {
	return cd.str
}

func (cd *ConnData) Domain() string {
	return cd.domain
}

func (cd *ConnData) Mode() DbMode {
	return cd.mode
}

func (cd *ConnData) Path() string {
	v, _ := url.Parse(cd.str)
	v.RawQuery = ""
	return v.String()
}
