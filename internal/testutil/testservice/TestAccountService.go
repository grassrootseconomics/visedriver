package testservice

import (
	"context"
	"encoding/json"
	"time"

	"git.grassecon.net/urdt/ussd/internal/models"
	"github.com/grassrootseconomics/eth-custodial/pkg/api"
)

type TestAccountService struct {
}

func (tas *TestAccountService) CreateAccount(ctx context.Context) (*api.OKResponse, error) {
	return &api.OKResponse{
		Ok:          true,
		Description: "Account creation succeeded",
		Result: map[string]any{
			"trackingId": "075ccc86-f6ef-4d33-97d5-e91cfb37aa0d",
			"publicKey":  "0x623EFAFa8868df4B934dd12a8B26CB3Dd75A7AdD",
		},
	}, nil
}

func (tas *TestAccountService) CheckBalance(ctx context.Context, publicKey string) (*models.BalanceResponse, error) {
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

func (tas *TestAccountService) CheckAccountStatus(ctx context.Context, trackingId string) (*models.TrackStatusResponse, error) {
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

func (tas *TestAccountService) TrackAccountStatus(ctx context.Context, publicKey string) (*api.OKResponse, error) {
	return &api.OKResponse{
		Ok:          true,
		Description: "Account creation succeeded",
		Result: map[string]any{
			"active": true,
		},
	}, nil
}
