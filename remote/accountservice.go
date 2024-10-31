package remote

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	dataserviceapi "github.com/grassrootseconomics/ussd-data-service/pkg/api"
	"github.com/grassrootseconomics/eth-custodial/pkg/api"
	"git.grassecon.net/urdt/ussd/config"
	"git.grassecon.net/urdt/ussd/internal/models"
)

var (
)

type AccountServiceInterface interface {
	//CheckBalance(ctx context.Context, publicKey string) (*models.BalanceResponse, error)
	CheckBalance(ctx context.Context, publicKey string) (*models.BalanceResult, error)
	CreateAccount(ctx context.Context) (*models.AccountResult, error)
	//CheckAccountStatus(ctx context.Context, trackingId string) (*models.TrackStatusResponse, error)
	// TODO: poorly named method seems to be checking transaction
	CheckAccountStatus(ctx context.Context, trackingId string) (*models.TrackStatusResult, error)
	TrackAccountStatus(ctx context.Context, publicKey string) (*api.OKResponse, error)
	FetchVouchers(ctx context.Context, publicKey string) ([]dataserviceapi.TokenHoldings, error)
	//FetchVouchers(ctx context.Context, publicKey string) (*api.OKResponse, error)
}

type AccountService struct {
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
func (as *AccountService) CheckAccountStatus(ctx context.Context, trackingId string) (*models.TrackStatusResult, error) {
	var r models.TrackStatusResult

	ep, err := url.JoinPath(config.TrackURL, trackingId)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", ep, nil)
	if err != nil {
		return nil, err
	}

	_, err =  doCustodialRequest(ctx, req, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (as *AccountService) TrackAccountStatus(ctx context.Context, publicKey string) (*api.OKResponse, error) {
	var okResponse  api.OKResponse
	var errResponse api.ErrResponse
	var err error
	
	ep, err := url.JoinPath(config.TrackURL, publicKey)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", ep, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GE-KEY", "xd")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		errResponse.Description = err.Error()
		return nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		err := json.Unmarshal([]byte(body), &errResponse)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(errResponse.Description)
	}
	err = json.Unmarshal([]byte(body), &okResponse)
	if err != nil {
		return nil, err
	}
	if len(okResponse.Result) == 0 {
		return nil, errors.New("Empty api result")
	}
	return &okResponse, nil

}

// CheckBalance retrieves the balance for a given public key from the custodial balance API endpoint.
// Parameters:
//   - publicKey: The public key associated with the account whose balance needs to be checked.
func (as *AccountService) CheckBalance(ctx context.Context, publicKey string) (*models.BalanceResult, error) {
	var balanceResult models.BalanceResult

	ep, err := url.JoinPath(config.BalanceURL,  publicKey)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", ep, nil)
	if err != nil {
		return nil, err
	}

	_, err = doCustodialRequest(ctx, req, &balanceResult)
	return &balanceResult, err
}


// CreateAccount creates a new account in the custodial system.
// Returns:
//   - *models.AccountResponse: A pointer to an AccountResponse struct containing the details of the created account.
//     If there is an error during the request or processing, this will be nil.
//   - error: An error if any occurred during the HTTP request, reading the response, or unmarshalling the JSON data.
//     If no error occurs, this will be nil.
func (as *AccountService) CreateAccount(ctx context.Context) (*models.AccountResult, error) {
	var r models.AccountResult
	// Create a new request
	req, err := http.NewRequest("POST", config.CreateAccountURL, nil)
	if err != nil {
		return nil, err
	}

	_, err =  doCustodialRequest(ctx, req, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

// FetchVouchers retrieves the token holdings for a given public key from the custodial holdings API endpoint
// Parameters:
//   - publicKey: The public key associated with the account.
func (as *AccountService) FetchVouchers(ctx context.Context, publicKey string) ([]dataserviceapi.TokenHoldings, error) {
	var r []dataserviceapi.TokenHoldings

	req, err := http.NewRequest("GET", config.VoucherHoldingsURL, nil)
	if err != nil {
		return nil, err
	}

	_, err =  doDataRequest(ctx, req, r)
	if err != nil {
		return nil, err
	}

	return r, nil

//
//	file, err := os.Open("sample_tokens.json")
//	if err != nil {
//		return nil, err
//	}
//	defer file.Close()
//	var holdings models.VoucherHoldingResponse
//
//	if err := json.NewDecoder(file).Decode(&holdings); err != nil {
//		return nil, err
//	}
//	return &holdings, nil
}

func doRequest(ctx context.Context, req *http.Request, rcpt any) (*api.OKResponse, error) {
	var okResponse  api.OKResponse
	var errResponse api.ErrResponse

	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		errResponse.Description = err.Error()
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		err := json.Unmarshal([]byte(body), &errResponse)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(errResponse.Description)
	}
	err = json.Unmarshal([]byte(body), &okResponse)
	if err != nil {
		return nil, err
	}
	if len(okResponse.Result) == 0 {
		return nil, errors.New("Empty api result")
	}
	return &okResponse, nil

	v, err := json.Marshal(okResponse.Result)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(v, &rcpt)
	return &okResponse, err
}

func doCustodialRequest(ctx context.Context, req *http.Request, rcpt any) (*api.OKResponse, error) {
	req.Header.Set("X-GE-KEY", config.CustodialAPIKey)
	return doRequest(ctx, req, rcpt)
}

func doDataRequest(ctx context.Context, req *http.Request, rcpt any) (*api.OKResponse, error) {
	req.Header.Set("X-GE-KEY", config.DataAPIKey)
	return doRequest(ctx, req, rcpt)
}
