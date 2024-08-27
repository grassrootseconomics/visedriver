package server

import (
	"encoding/json"
	"io"
	"net/http"

	"git.grassecon.net/urdt/ussd/config"
	"git.grassecon.net/urdt/ussd/internal/models"
)

//CreateAccount creates a new account in the custodial system.
// Returns:
//   - *models.AccountResponse: A pointer to an AccountResponse struct containing the details of the created account.
//     If there is an error during the request or processing, this will be nil.
//   - error: An error if any occurred during the HTTP request, reading the response, or unmarshalling the JSON data.
//     If no error occurs, this will be nil.
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