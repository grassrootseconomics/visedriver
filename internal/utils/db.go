package utils

import (
	"context"
	"encoding/binary"

	"git.defalsify.org/vise.git/db"
)

type DataTyp uint16

const (
	DATA_ACCOUNT DataTyp = iota
	DATA_ACCOUNT_CREATED
	DATA_TRACKING_ID
	DATA_PUBLIC_KEY
	DATA_CUSTODIAL_ID
)

func typToBytes(typ DataTyp) []byte {
	//var b []byte
	b := make([]byte, 2) 
	binary.BigEndian.PutUint16(b, uint16(typ))
	return b
}

func packKey(typ DataTyp, data []byte) []byte {
	v := typToBytes(typ)
	return append(v, data...)
}

func ReadEntry(ctx context.Context, store db.Db, sessionId string, typ DataTyp) ([]byte, error) {
	store.SetPrefix(db.DATATYPE_USERDATA)
	 store.SetSession(sessionId)
	 k := packKey(typ, []byte(sessionId))
	//k := []byte(sessionId)
	b, err := store.Get(ctx, k)
	if err != nil {
		return nil, err
	}
	return b, nil

}

func WriteEntry(ctx context.Context, store db.Db, sessionId string, typ DataTyp, value []byte) error {
	store.SetPrefix(db.DATATYPE_USERDATA)
	store.SetSession(sessionId)
	//k := packKey(typ, []byte(sessionId))
	k := []byte(sessionId)
	return store.Put(ctx, k, value)
}