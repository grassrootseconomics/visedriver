// +build debugdb

package debug

import (
	"git.defalsify.org/vise.git/db"

	"git.grassecon.net/urdt/ussd/common"
	"git.grassecon.net/urdt/ussd/internal/storage"
)

func init() {
	DebugCap |= 1
	dbTypStr[db.DATATYPE_STATE] = "internal_state"
	dbTypStr[db.DATATYPE_USERDATA] = "userdata"
	dbTypStr[storage.DATATYPE_USERSUB + 1 + common.DATA_ACCOUNT] = "account"
	dbTypStr[storage.DATATYPE_USERSUB + 1 + common.DATA_ACCOUNT_CREATED] = "account_created"
	dbTypStr[storage.DATATYPE_USERSUB + 1 + common.DATA_TRACKING_ID] = "tracking id"
	dbTypStr[storage.DATATYPE_USERSUB + 1 + common.DATA_PUBLIC_KEY] = "public key"
	dbTypStr[storage.DATATYPE_USERSUB + 1 + common.DATA_CUSTODIAL_ID] = "custodial id"
	dbTypStr[storage.DATATYPE_USERSUB + 1 + common.DATA_ACCOUNT_PIN] = "account pin"
	dbTypStr[storage.DATATYPE_USERSUB + 1 + common.DATA_ACCOUNT_STATUS] = "account status"
	dbTypStr[storage.DATATYPE_USERSUB + 1 + common.DATA_FIRST_NAME] = "first name"
	dbTypStr[storage.DATATYPE_USERSUB + 1 + common.DATA_FAMILY_NAME] = "family name"
	dbTypStr[storage.DATATYPE_USERSUB + 1 + common.DATA_YOB] = "year of birth"
	dbTypStr[storage.DATATYPE_USERSUB + 1 + common.DATA_LOCATION] = "location"
	dbTypStr[storage.DATATYPE_USERSUB + 1 + common.DATA_GENDER] = "gender"
	dbTypStr[storage.DATATYPE_USERSUB + 1 + common.DATA_OFFERINGS] = "offerings"
	dbTypStr[storage.DATATYPE_USERSUB + 1 + common.DATA_RECIPIENT] = "recipient"
	dbTypStr[storage.DATATYPE_USERSUB + 1 + common.DATA_AMOUNT] = "amount"
	dbTypStr[storage.DATATYPE_USERSUB + 1 + common.DATA_TEMPORARY_VALUE] = "temporary value"
	dbTypStr[storage.DATATYPE_USERSUB + 1 + common.DATA_ACTIVE_SYM] = "active sym"
	dbTypStr[storage.DATATYPE_USERSUB + 1 + common.DATA_ACTIVE_BAL] = "active bal"
	dbTypStr[storage.DATATYPE_USERSUB + 1 + common.DATA_BLOCKED_NUMBER] = "blocked number"
	dbTypStr[storage.DATATYPE_USERSUB + 1 + common.DATA_PUBLIC_KEY_REVERSE] = "public_key_reverse"
	dbTypStr[storage.DATATYPE_USERSUB + 1 + common.DATA_ACTIVE_DECIMAL] = "active decimal"
	dbTypStr[storage.DATATYPE_USERSUB + 1 + common.DATA_ACTIVE_ADDRESS] = "active address"
	dbTypStr[storage.DATATYPE_USERSUB + 1 + common.DATA_TRANSACTIONS] = "transactions"
}
