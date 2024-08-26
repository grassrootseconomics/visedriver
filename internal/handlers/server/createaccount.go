package server

import (
	"encoding/json"
	"io"
	"net/http"

	"git.grassecon.net/urdt/ussd/config"
	"git.grassecon.net/urdt/ussd/internal/models"
)


func CreateAccount() (*models.AccountResponse, error) {
	resp, err := http.Post(config.CreateAccountURL, "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var accountResp models.AccountResponse
	err = json.Unmarshal(body, &accountResp)
	if err != nil {
		return nil, err
	}

	return &accountResp, nil
}