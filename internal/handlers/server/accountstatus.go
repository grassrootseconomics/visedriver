package server

import (
	"encoding/json"
	"io"
	"net/http"

	"git.grassecon.net/urdt/ussd/config"
	"git.grassecon.net/urdt/ussd/internal/models"
)

// CheckAccountStatus retrieves the status of an account transaction based on the provided tracking ID.
//
// Parameters:
//   - trackingId: A unique identifier for the account.This should be obtained from a previous call to
//     CreateAccount or a similar function that returns an AccountResponse. The `trackingId` field in the
//     AccountResponse struct can be used here to check the account status during a transaction.
//
//
// Returns:
//   - string: The status of the transaction as a string. If there is an error during the request or processing, this will be an empty string.
//   - error: An error if any occurred during the HTTP request, reading the response, or unmarshalling the JSON data.
//     If no error occurs, this will be nil.
//
func CheckAccountStatus(trackingId string) (string, error) {
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
