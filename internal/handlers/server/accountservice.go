package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"git.grassecon.net/urdt/ussd/config"
	"git.grassecon.net/urdt/ussd/internal/models"
)

var apiResponse struct {
	Ok          bool   `json:"ok"`
	Description string `json:"description"`
}

type AccountServiceInterface interface {
	CheckBalance(publicKey string) (*models.BalanceResponse, error)
	CreateAccount() (*OKResponse, *ErrResponse)
	CheckAccountStatus(trackingId string) (*models.TrackStatusResponse, error)
	TrackAccountStatus(publicKey string) (*OKResponse, *ErrResponse)
}

type AccountService struct {
}

type TestAccountService struct {
}

// Parameters:
//   - trackingId: A unique identifier for the account.This should be obtained from a previous call to
//     CreateAccount or a similar function that returns an AccountResponse. The `trackingId` field in the
//     AccountResponse struct can be used here to check the account status during a transaction.
//
// Returns:
//   - string: The status of the transaction as a string. If there is an error during the request or processing, this will be an empty string.
//   - error: An error if any occurred during the HTTP request, reading the response, or unmarshalling the JSON data.
//     If no error occurs, this will be nil
func (as *AccountService) CheckAccountStatus(trackingId string) (*models.TrackStatusResponse, error) {
	resp, err := http.Get(config.BalanceURL + trackingId)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var trackResp models.TrackStatusResponse
	err = json.Unmarshal(body, &trackResp)
	if err != nil {
		return nil, err
	}
	return &trackResp, nil

}

func (as *AccountService) TrackAccountStatus(publicKey string) (*OKResponse, *ErrResponse) {
	var errResponse ErrResponse
	var okResponse OKResponse
	var err error
	// Construct the URL with the path parameter
	url := fmt.Sprintf("%s/%s", config.TrackURL, publicKey)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		errResponse.Description = err.Error()
		return nil, &errResponse
	}
	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GE-KEY", "xd")

	// Send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		errResponse.Description = err.Error()
		return nil, &errResponse
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errResponse.Description = err.Error()
		return nil, &errResponse
	}

	// Step 2: Unmarshal into the generic struct
	err = json.Unmarshal([]byte(body), &apiResponse)
	if err != nil {
		errResponse.Description = err.Error()
		return nil, &errResponse
	}
	if apiResponse.Ok {
		err = json.Unmarshal([]byte(body), &okResponse)
		if err != nil {
			errResponse.Description = err.Error()
			return nil, &errResponse
		}
		return &okResponse, nil
	} else {
		err := json.Unmarshal([]byte(body), &errResponse)
		if err != nil {
			errResponse.Description = err.Error()
			return nil, &errResponse
		}
		return nil, &errResponse
	}
}

// CheckBalance retrieves the balance for a given public key from the custodial balance API endpoint.
// Parameters:
//   - publicKey: The public key associated with the account whose balance needs to be checked.
func (as *AccountService) CheckBalance(publicKey string) (*models.BalanceResponse, error) {
	resp, err := http.Get(config.BalanceURL + publicKey)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var balanceResp models.BalanceResponse
	err = json.Unmarshal(body, &balanceResp)
	if err != nil {
		return nil, err
	}
	return &balanceResp, nil
}

// CreateAccount creates a new account in the custodial system.
// Returns:
//   - *models.AccountResponse: A pointer to an AccountResponse struct containing the details of the created account.
//     If there is an error during the request or processing, this will be nil.
//   - error: An error if any occurred during the HTTP request, reading the response, or unmarshalling the JSON data.
//     If no error occurs, this will be nil.
func (as *AccountService) CreateAccount() (*OKResponse, *ErrResponse) {

	var errResponse ErrResponse
	var okResponse OKResponse
	var err error

	// Create a new request
	req, err := http.NewRequest("POST", config.CreateAccountURL, nil)
	if err != nil {
		errResponse.Description = err.Error()
		return nil, &errResponse
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GE-KEY", "xd")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		errResponse.Description = err.Error()
		return nil, &errResponse
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errResponse.Description = err.Error()
		return nil, &errResponse
	}
	err = json.Unmarshal([]byte(body), &apiResponse)
	if err != nil {
		return nil, &errResponse
	}
	if apiResponse.Ok {
		err = json.Unmarshal([]byte(body), &okResponse)
		if err != nil {
			errResponse.Description = err.Error()
			return nil, &errResponse
		}
		return &okResponse, nil
	} else {
		err := json.Unmarshal([]byte(body), &errResponse)
		if err != nil {
			errResponse.Description = err.Error()
			return nil, &errResponse
		}
		return nil, &errResponse
	}
}

func (tas *TestAccountService) CreateAccount() (*OKResponse, *ErrResponse) {
	return &OKResponse{
		Ok:          true,
		Description: "Account creation request received successfully",
		Result:      map[string]any{"publicKey": "0x48ADca309b5085852207FAaf2816eD72B52F527C", "trackingId": "28ebe84d-b925-472c-87ae-bbdfa1fb97be"},
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

func (tas *TestAccountService) TrackAccountStatus(publicKey string) (*OKResponse, *ErrResponse) {
	return &OKResponse{
		Ok:          true,
		Description: "Account creation succeeded",
		Result: map[string]any{
			"active": true,
		},
	}, nil
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
