package testservice

import (
	"encoding/json"
	"time"

	"git.grassecon.net/urdt/ussd/internal/models"
)

type TestAccountService struct {
}

func (tas *TestAccountService) CreateAccount() (*models.AccountResponse, error) {
	return &models.AccountResponse{
		Ok: true,
		Result: struct {
			CustodialId json.Number `json:"custodialId"`
			PublicKey   string      `json:"publicKey"`
			TrackingId  string      `json:"trackingId"`
		}{
			CustodialId: json.Number("182"),
			PublicKey:   "0x48ADca309b5085852207FAaf2816eD72B52F527C",
			TrackingId:  "28ebe84d-b925-472c-87ae-bbdfa1fb97be",
		},
	}, nil
}

func (tas *TestAccountService) CheckBalance(publicKey string) (*models.BalanceResponse, error) {
	balanceResponse := &models.BalanceResponse{
		Ok: true,
		Result: struct {
			Balance string      `json:"balance"`
			Nonce   json.Number `json:"nonce"`
		}{
			Balance: "0.003 CELO",
			Nonce:   json.Number("0"),
		},
	}

	return balanceResponse, nil
}

func (tas *TestAccountService) CheckAccountStatus(trackingId string) (*models.TrackStatusResponse, error) {
	trackResponse := &models.TrackStatusResponse{
		Ok: true,
		Result: struct {
			Transaction struct {
				CreatedAt     time.Time   "json:\"createdAt\""
				Status        string      "json:\"status\""
				TransferValue json.Number "json:\"transferValue\""
				TxHash        string      "json:\"txHash\""
				TxType        string      "json:\"txType\""
			}
		}{
			Transaction: models.Transaction{
				CreatedAt:     time.Now(),
				Status:        "SUCCESS",
				TransferValue: json.Number("0.5"),
				TxHash:        "0x123abc456def",
				TxType:        "transfer",
			},
		},
	}
	return trackResponse, nil
}
