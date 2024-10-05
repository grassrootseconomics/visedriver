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
	FetchVouchers(publicKey string) (*models.VoucherHoldingResponse, error)
}

type AccountService struct {
}

// CheckAccountStatus retrieves the status of an account transaction based on the provided tracking ID.
//
// Parameters:
//   - trackingId: A unique identifier for the account.This should be obtained from a previous call to
//     CreateAccount or a similar function that returns an AccountResponse. The `trackingId` field in the
//     AccountResponse struct can be used here to check the account status during a transaction.
//
// Returns:
//   - string: The status of the transaction as a string. If there is an error during the request or processing, this will be an empty string.
//   - error: An error if any occurred during the HTTP request, reading the response, or unmarshalling the JSON data.
//     If no error occurs, this will be nil.
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

// CheckBalance retrieves the balance for a given public key from the custodial balance API endpoint.
// Parameters:
//   - publicKey: The public key associated with the account whose balance needs to be checked.
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

// CreateAccount creates a new account in the custodial system.
// Returns:
//   - *models.AccountResponse: A pointer to an AccountResponse struct containing the details of the created account.
//     If there is an error during the request or processing, this will be nil.
//   - error: An error if any occurred during the HTTP request, reading the response, or unmarshalling the JSON data.
//     If no error occurs, this will be nil.
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

// FetchVouchers retrieves the token holdings for a given public key from the custodial holdings API endpoint
// Parameters:
//   - publicKey: The public key associated with the account.
func (as *AccountService) FetchVouchers(publicKey string) (*models.VoucherHoldingResponse, error) {
	// TODO replace with the actual request once ready
	mockJSON := `{
	"ok": true,
	"description": "Token holdings with current balances",
		"result": {
			"holdings": [
				{
					"contractAddress": "0x6CC75A06ac72eB4Db2eE22F781F5D100d8ec03ee",
					"tokenSymbol": "FSPTST",
					"tokenDecimals": "6",
					"balance": "8869964242"
				},
				{
					"contractAddress": "0x724F2910D790B54A39a7638282a45B1D83564fFA",
					"tokenSymbol": "GEO",
					"tokenDecimals": "6",
					"balance": "9884"
				},
				{
					"contractAddress": "0x2105a206B7bec31E2F90acF7385cc8F7F5f9D273",
					"tokenSymbol": "MFNK",
					"tokenDecimals": "6",
					"balance": "19788697"
				},
				{
					"contractAddress": "0x63DE2Ac8D1008351Cc69Fb8aCb94Ba47728a7E83",
					"tokenSymbol": "MILO",
					"tokenDecimals": "6",
					"balance": "75"
				},
				{
					"contractAddress": "0xd4c288865Ce0985a481Eef3be02443dF5E2e4Ea9",
					"tokenSymbol": "SOHAIL",
					"tokenDecimals": "6",
					"balance": "27874115"
				},
				{
					"contractAddress": "0x45d747172e77d55575c197CbA9451bC2CD8F4958",
					"tokenSymbol": "SRQIF",
					"tokenDecimals": "6",
					"balance": "2745987"
				},
				{
					"contractAddress": "0x45d747172e77d55575c197CbA9451bC2CD8F4958",
					"tokenSymbol": "SRFI",
					"tokenDecimals": "6",
					"balance": "2745987"
				},
				{
					"contractAddress": "0x45d747172e77d55575c197CbA9451bC2CD8F4958",
					"tokenSymbol": "SRFU",
					"tokenDecimals": "6",
					"balance": "2745987"
				},
				{
					"contractAddress": "0x45d747172e77d55575c197CbA9451bC2CD8F4958",
					"tokenSymbol": "SRQF",
					"tokenDecimals": "6",
					"balance": "2745987"
				},
				{
					"contractAddress": "0x45d747172e77d55575c197CbA9451bC2CD8F4958",
					"tokenSymbol": "SREF",
					"tokenDecimals": "6",
					"balance": "2745987"
				}
			]
		}
	}`

	// Unmarshal the JSON response
	var holdings models.VoucherHoldingResponse
	err := json.Unmarshal([]byte(mockJSON), &holdings)
	if err != nil {
		return nil, err
	}

	return &holdings, nil
}
