package common

import (
	"encoding/binary"
	"errors"

	"git.defalsify.org/vise.git/logging"
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
	DATA_TEMPORARY_VALUE
	DATA_ACTIVE_SYM
	DATA_ACTIVE_BAL
	DATA_BLOCKED_NUMBER
	DATA_PUBLIC_KEY_REVERSE
	DATA_ACTIVE_DECIMAL
	DATA_ACTIVE_ADDRESS
	DATA_TRANSACTIONS
)

var (
	logg = logging.NewVanilla().WithDomain("urdt-common")
)

func typToBytes(typ DataTyp) []byte {
	var b [2]byte
	binary.BigEndian.PutUint16(b[:], uint16(typ))
	return b[:]
}

func PackKey(typ DataTyp, data []byte) []byte {
	v := typToBytes(typ)
	return append(v, data...)
}

func StringToDataTyp(str string) (DataTyp, error) {
	switch str {
	case "DATA_FIRST_NAME":
		return DATA_FIRST_NAME, nil
	case "DATA_FAMILY_NAME":
		return DATA_FAMILY_NAME, nil
	case "DATA_YOB":
		return DATA_YOB, nil
	case "DATA_LOCATION":
		return DATA_LOCATION, nil
	case "DATA_GENDER":
		return DATA_GENDER, nil
	case "DATA_OFFERINGS":
		return DATA_OFFERINGS, nil

	default:
		return 0, errors.New("invalid DataTyp string")
	}
}
