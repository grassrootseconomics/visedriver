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
	DATA_ACCOUNT_PIN
	DATA_ACCOUNT_STATUS
	DATA_FIRST_NAME
	DATA_FAMILY_NAME
	DATA_YOB
	DATA_LOCATION
	DATA_GENDER
	DATA_OFFERINGS
	DATA_RECIPIENT
	DATA_AMOUNT
)

func typToBytes(typ DataTyp) []byte {
	var b [2]byte
	binary.BigEndian.PutUint16(b[:], uint16(typ))
	return b[:]
}

func packKey(typ DataTyp, data []byte) []byte {
	v := typToBytes(typ)
	return append(v, data...)
}

func ReadEntry(ctx context.Context, store db.Db, sessionId string, typ DataTyp) ([]byte, error) {
	store.SetPrefix(db.DATATYPE_USERDATA)
	store.SetSession(sessionId)
	k := packKey(typ, []byte(sessionId))
	b, err := store.Get(ctx, k)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func WriteEntry(ctx context.Context, store db.Db, sessionId string, typ DataTyp, value []byte) error {
	store.SetPrefix(db.DATATYPE_USERDATA)
	store.SetSession(sessionId)
	k := packKey(typ, []byte(sessionId))
	return store.Put(ctx, k, value)
}
