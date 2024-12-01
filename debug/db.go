package debug

import (
	"fmt"
	"encoding/binary"

	"git.grassecon.net/urdt/ussd/internal/storage"
	"git.grassecon.net/urdt/ussd/common"
)

var (
	dbTypStr map[common.DataTyp]string = make(map[common.DataTyp]string)
)

type KeyInfo struct {
	SessionId string
	Typ uint8
	SubTyp common.DataTyp
	Label string
	Description string
}

func ToKeyInfo(k []byte, sessionId string) (KeyInfo, error) {
	o := KeyInfo{}
	b := []byte(sessionId)

	if len(k) <= len(b) {
		return o, fmt.Errorf("storage key missing")
	}

	o.SessionId = sessionId

	k = k[len(b):]
	o.Typ = k[0]
	k = k[1:]

	if o.Typ == storage.DATATYPE_USERSUB {
		if len(k) == 0 {
			return o, fmt.Errorf("missing subtype key")
		}
		v := binary.BigEndian.Uint16(k[:2])
		o.SubTyp = common.DataTyp(v)
		o.Label = subTypToString(o.SubTyp)
		k = k[2:]
	} else {
		o.Label = typToString(o.Typ)
	}

	if len(k) != 0 {
		return o, fmt.Errorf("excess key information")
	}

	return o, nil
}

func subTypToString(v common.DataTyp) string {
	return dbTypStr[v + storage.DATATYPE_USERSUB + 1]
}

func typToString(v uint8) string {
	return dbTypStr[common.DataTyp(uint16(v))]
}
