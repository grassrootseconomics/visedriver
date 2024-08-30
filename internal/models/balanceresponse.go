package models

import "encoding/json"


type BalanceResponse struct {
	Ok     bool `json:"ok"`
	Result struct {
		Balance string      `json:"balance"`
		Nonce   json.Number `json:"nonce"`
	} `json:"result"`
}
