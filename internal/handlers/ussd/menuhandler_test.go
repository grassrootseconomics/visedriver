package ussd

import (
	"encoding/json"
	"testing"

	"git.grassecon.net/urdt/ussd/internal/handlers/ussd/mocks"
	"git.grassecon.net/urdt/ussd/internal/models"
	"github.com/alecthomas/assert/v2"
	"github.com/stretchr/testify/mock"
)

func TestCreateAccount_Success(t *testing.T) {
	mockAccountFileHandler := new(mocks.MockAccountFileHandler)
	mockCreateAccountService := new(mocks.MockAccountService)

	mockAccountFileHandler.On("EnsureFileExists").Return(nil)

	// Mock that no account data exists
	mockAccountFileHandler.On("ReadAccountData").Return(nil, nil)

	// Define expected account response after api call
	expectedAccountResp := &models.AccountResponse{
		Ok: true,
		Result: struct {
			CustodialId json.Number `json:"custodialId"`
			PublicKey   string      `json:"publicKey"`
			TrackingId  string      `json:"trackingId"`
		}{
			CustodialId: "12",
			PublicKey:   "some-public-key",
			TrackingId:  "some-tracking-id",
		},
	}
	mockCreateAccountService.On("CreateAccount").Return(expectedAccountResp, nil)

	// Mock WriteAccountData to not error
	mockAccountFileHandler.On("WriteAccountData", mock.Anything).Return(nil)

	handlers := &Handlers{
		accountService:     mockCreateAccountService,
	}

	
	actualResponse, err := handlers.accountService.CreateAccount()

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t,expectedAccountResp.Ok,true)
	assert.Equal(t,expectedAccountResp,actualResponse)
	

}
