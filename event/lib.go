package event

import (
	tevent "github.com/grassrootseconomics/eth-tracker/pkg/event"
	trouter "github.com/grassrootseconomics/eth-tracker/pkg/router"
)

var (
	typeTokenMint = "TOKEN_MINT"
	typeTokenTransfer = "TOKEN_TRANSFER"
	typeFaucetGive = "FAUCET_GIVE"
)

func processFaucetGive(db db.Db, payload map[string]any) {
	var sessionId string
	var tokenString string
	var amount uint64


}
