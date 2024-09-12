package server

import (
	"testing"

	"github.com/alecthomas/assert/v2"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/recorder"
)

func TestCheckBalance(t *testing.T) {
	r, err := recorder.New("custodial/balance")
	if err != nil {
		t.Fatal(err)
	}
	defer r.Stop()

	client := r.GetDefaultClient()

	as := AccountService{
		Client: client,
	}
	tests := []struct {
		name      string
		balance   string
		publicKey string
	}{
		{
			name:      "Test check balance with correct public key",
			publicKey: "0x216a4A64E1e699F9d65Dd9CbD0058dAB21DeF002",
			balance:   "3.06000000003 CELO",
		},
		{
			name:      "Test check balance with public key that doesn't exist in the custodial system",
			balance:   "",
			publicKey: "0x216a4A64E1e699F9d65Dd9CbD0058dAB21DeF00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			balance, err := as.CheckBalance(tt.publicKey)
			if err != nil {
				t.Fatalf("Failed to get balance with error %s", err)
			}
			if err != nil {

				return
			}
			assert.NoError(t, err)
			assert.Equal(t, balance, tt.balance, "Expected balance and actual balance should be equal")

		})
	}

}

func TestCheckAccountStatus(t *testing.T) {
	r, err := recorder.New("custodial/status")
	if err != nil {
		t.Fatal(err)
	}
	defer r.Stop()

	client := r.GetDefaultClient()

	as := AccountService{
		Client: client,
	}
	tests := []struct {
		name       string
		status     string
		trackingId string
	}{
		{
			name:       "Test check status with tracking id that exists in the custodial system",
			trackingId: "bb23945b-65cd-4110-ac2e-a5df40572e18",
			status:     "SUCCESS",
		},
		{
			name:       "Test check status with tracking id that doesn't exist in the custodial system",
			status:     "",
			trackingId: "bb23945b-65cd-4110-ac2e-a5df40572e1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := as.CheckAccountStatus(tt.trackingId)
			if err != nil {
				t.Fatalf("Failed to account status with error %s", err)
			}
			if err != nil {
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, status, tt.status, "Expected status and actual status should be equal")

		})
	}

}

func TestCreateAccount(t *testing.T) {
	r, err := recorder.New("custodial/create")
	if err != nil {
		t.Fatal(err)
	}
	defer r.Stop()

	client := r.GetDefaultClient()

	as := AccountService{
		Client: client,
	}
	accountRes, err := as.CreateAccount()

	if err != nil {
		t.Fatalf("Failed to create an account with error %s", err)
	}
	if err != nil {
		return
	}
	assert.NoError(t, err)
	assert.Equal(t, accountRes.Ok, true, "account response status is true")

}
