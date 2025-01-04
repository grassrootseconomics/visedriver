package storage

import (
	"fmt"
	"net/url"
	"path"
)

const (
	DBTYPE_MEM = iota
	DBTYPE_GDBM
	DBTYPE_POSTGRES
)

type connData struct {
	typ int
	str string
}

func (cd *connData) DbType() int {
	return cd.typ
}

func (cd *connData) String() string {
	return cd.str
}

func probePostgres(s string) (string, bool) {
	v, err := url.Parse(s)
	if err != nil {
		return "", false
	}
	if v.Scheme != "postgres" {
		return "", false
	}
	return s, true
}

func probeGdbm(s string) (string, bool) {
	if !path.IsAbs(s) {
		return "", false
	}
	if path.Ext(s) != ".gdbm" {
		return "", false
	}
	s = path.Clean(s)
	return s, true
}

func toConnData(connStr string) (connData, error) {
	var o connData

	if connStr == "" {
		return o, nil
	}

	v, ok := probePostgres(connStr)
	if ok {
		o.typ = DBTYPE_POSTGRES
		o.str = v
		return o, nil
	}

	v, ok = probeGdbm(connStr)
	if ok {
		o.typ = DBTYPE_GDBM
		o.str = v
		return o, nil
	}

	return o, fmt.Errorf("invalid connection string: %s", connStr)
}
