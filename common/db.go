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
	// Start the sub prefix data at 256 (0x0100)
	DATA_VOUCHER_SYMBOLS DataTyp = 256 + iota
	DATA_VOUCHER_BALANCES
	DATA_VOUCHER_DECIMALS
	DATA_VOUCHER_ADDRESSES
	DATA_PREFIX_TX_SENDERS
	DATA_PREFIX_TX_RECIPIENTS
	DATA_PREFIX_TX_VALUES
	DATA_PREFIX_TX_ADDRESSES
	DATA_PREFIX_TX_HASHES
	DATA_PREFIX_TX_DATES
	DATA_PREFIX_TX_SYMBOLS
	DATA_PREFIX_TX_DECIMALS
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

// Convert DataTyp to []byte
func (d DataTyp) ToBytes() []byte {
	bytes := make([]byte, 2)
	binary.BigEndian.PutUint16(bytes, uint16(d))
	return bytes
}

// Convert int to []byte
func IntToBytes(value int) []byte {
	bytes := make([]byte, 2)
	binary.BigEndian.PutUint16(bytes, uint16(value))
	return bytes
}
