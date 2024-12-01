package debug

import (
	"fmt"
	"encoding/binary"

	"git.grassecon.net/urdt/ussd/internal/storage"
	"git.grassecon.net/urdt/ussd/common"
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
		o.Label = typToString(o.SubTyp)
		k = k[2:]
	}

	if len(k) != 0 {
		return o, fmt.Errorf("excess key information")
	}

	return o, nil
}
