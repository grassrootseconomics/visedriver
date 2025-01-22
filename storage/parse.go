package storage

import (
	"fmt"
	"net/url"
	"path"
	"path/filepath"
)

func probePostgres(s string) (string, string, bool) {
	domain := "public"
	v, err := url.Parse(s)
	if err != nil {
		return "", "", false
	}
	if v.Scheme != "postgres" {
		return "", "", false
	}
	vv := v.Query()
	if vv.Has("search_path") {
		domain = vv.Get("search_path")
	}
	return s, domain, true
}

func probeGdbm(s string) (string, string, bool) {
	domain := "public"
	v, err := url.Parse(s)
	if err != nil {
		return "", "", false
	}
	if v.Scheme != "gdbm" {
		return "", "", false
	}
	s = v.Path
	return s, domain, true
}

func probeFs(s string) (string, string, bool) {
	var err error

	v, _ := url.Parse(s)
	if v.Scheme != "" && v.Scheme != "file://" {
		return "", "", false
	}

	if !path.IsAbs(s) {
		s, err = filepath.Abs(s)
		if err != nil {
			panic(err)
		}
	}
	s = path.Clean(s)
	return s, "", true
}

func probeMem(s string) (string, string, bool) {
	if s != "" {
		return "", "", false
	}
	return "", "", true
}

func ToConnDataMode(connStr string, mode DbMode) (ConnData, error) {
	o, err := ToConnData(connStr)
	if err != nil {
		return o, err
	}
	o.mode = mode
	return o, nil
}

func ToConnData(connStr string) (ConnData, error) {
	var o ConnData

	v, domain, ok := probeMem(connStr)
	if ok {
		o.typ = DBTYPE_MEM
		return o, nil
	}

	v, domain, ok = probePostgres(connStr)
	if ok {
		o.typ = DBTYPE_POSTGRES
		o.str = v
		o.domain = domain
		return o, nil
	}

	v, _, ok = probeGdbm(connStr)
	if ok {
		o.typ = DBTYPE_GDBM
		o.str = v
		return o, nil
	}

	v, _, ok = probeFs(connStr)
	if ok {
		o.typ = DBTYPE_FS
		o.str = v
		return o, nil
	}

	return o, fmt.Errorf("invalid connection string: %s", connStr)
}
