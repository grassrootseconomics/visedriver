// +build debugdb

package debug

import (
	"git.defalsify.org/vise.git/db"

	"git.grassecon.net/urdt/ussd/common"
)

var (
	dbTypStr map[common.DataTyp]string = map[common.DataTyp]string {
		db.DATATYPE_USERDATA: "userdata",
		common.DATA_ACCOUNT: "account",
		common.DATA_ACCOUNT_CREATED: "account created",
		common.DATA_TRACKING_ID: "tracking id",
		common.DATA_PUBLIC_KEY: "public key",
		common.DATA_CUSTODIAL_ID: "custodial id",
		common.DATA_ACCOUNT_PIN: "account pin",
		common.DATA_ACCOUNT_STATUS: "account status",
		common.DATA_FIRST_NAME: "first name",
		common.DATA_FAMILY_NAME: "family name",
		common.DATA_YOB: "year of birth",
		common.DATA_LOCATION: "location",
		common.DATA_GENDER: "gender", 
		common.DATA_OFFERINGS: "offerings",
		common.DATA_RECIPIENT: "recipient",
		common.DATA_AMOUNT: "amount",
		common.DATA_TEMPORARY_VALUE: "temporary value",
		common.DATA_ACTIVE_SYM: "active sym",
		common.DATA_ACTIVE_BAL: "active bal",
		common.DATA_BLOCKED_NUMBER: "blocked number",
		common.DATA_PUBLIC_KEY_REVERSE: "public key reverse",
		common.DATA_ACTIVE_DECIMAL: "active decimal",
		common.DATA_ACTIVE_ADDRESS: "active address",
		common.DATA_TRANSACTIONS: "transactions",
	}
)

func init() {
	DebugCap |= 1
}

func typToString(v common.DataTyp) string {
	return dbTypStr[v]
}
