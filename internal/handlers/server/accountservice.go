package server

import (
	"encoding/json"
	"io"
	"net/http"

	"git.grassecon.net/urdt/ussd/config"
	"git.grassecon.net/urdt/ussd/internal/models"
)

type AccountServiceInterface interface {
	CheckBalance(publicKey string) (string, error)
	CreateAccount() (*models.AccountResponse, error)
	CheckAccountStatus(trackingId string) (string, error)
}

type AccountService struct {
}

func (as *AccountService) CheckAccountStatus(trackingId string) (string, error) {
	resp, err := http.Get(config.TrackStatusURL + trackingId)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var trackResp models.TrackStatusResponse
	err = json.Unmarshal(body, &trackResp)
	if err != nil {
		return "", err
	}

	status := trackResp.Result.Transaction.Status

	return status, nil
}

func (as *AccountService) CheckBalance(publicKey string) (string, error) {

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

func (as *AccountService) CreateAccount() (*models.AccountResponse, error) {
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
