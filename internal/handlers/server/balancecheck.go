package server

import (
	"encoding/json"
	"io"
	"net/http"

	"git.grassecon.net/urdt/ussd/config"
	"git.grassecon.net/urdt/ussd/internal/models"
)

func CheckBalance(publicKey string) (string, error) {

	resp, err := http.Get(config.BalanceURL + publicKey)
	if err != nil {
		return "0.0", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "0.0", err
	}

	var balanceResp models.BalanceResponse
	err = json.Unmarshal(body, &balanceResp)
	if err != nil {
		return "0.0", err
	}

	balance := balanceResp.Result.Balance
	return balance, nil
}
