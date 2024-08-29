package ussd

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"git.defalsify.org/vise.git/resource"
	"git.grassecon.net/urdt/ussd/internal/models"
	"git.grassecon.net/urdt/ussd/internal/utils"
)

// MockAccountCreator implements AccountCreator for testing
type MockAccountCreator struct {
	mockResponse *models.AccountResponse
	mockError    error
}

func (m *MockAccountCreator) CreateAccount() (*models.AccountResponse, error) {
	return m.mockResponse, m.mockError
}

func TestCreateAccount(t *testing.T) {
	// Setup
	tempDir, err := os.MkdirTemp("", "test_create_account")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir) // Clean up after the test run

	sessionID := "07xxxxxxxx"

	// Set up the data file path using the session ID
	accountFilePath := filepath.Join(tempDir, sessionID+"_data")

	// Initialize account file handler
	accountFileHandler := utils.NewAccountFileHandler(accountFilePath)

	// Create a mock account creator
	mockAccountCreator := &MockAccountCreator{
		mockResponse: &models.AccountResponse{
			Ok: true,
			Result: struct {
				CustodialId json.Number `json:"custodialId"`
				PublicKey   string      `json:"publicKey"`
				TrackingId  string      `json:"trackingId"`
			}{
				CustodialId: "test-custodial-id",
				PublicKey:   "test-public-key",
				TrackingId:  "test-tracking-id",
			},
		},
	}

	// Initialize Handlers with mock account creator
	h := &Handlers{
		accountFileHandler: accountFileHandler,
		accountCreator:     mockAccountCreator,
	}

	tests := []struct {
		name           string
		existingData   map[string]string
		expectedResult resource.Result
		expectedData   map[string]string
	}{
		{
			name:         "New account creation",
			existingData: nil,
			expectedResult: resource.Result{
				FlagSet: []uint32{models.USERFLAG_ACCOUNT_CREATED},
			},
			expectedData: map[string]string{
				"TrackingId":  "test-tracking-id",
				"PublicKey":   "test-public-key",
				"CustodialId": "test-custodial-id",
				"Status":      "PENDING",
				"Gender":      "Not provided",
				"YOB":         "Not provided",
				"Location":    "Not provided",
				"Offerings":   "Not provided",
				"FirstName":   "Not provided",
				"FamilyName":  "Not provided",
			},
		},
		{
			name: "Existing account",
			existingData: map[string]string{
				"TrackingId":  "test-tracking-id",
				"PublicKey":   "test-public-key",
				"CustodialId": "test-custodial-id",
				"Status":      "PENDING",
				"Gender":      "Not provided",
				"YOB":         "Not provided",
				"Location":    "Not provided",
				"Offerings":   "Not provided",
				"FirstName":   "Not provided",
				"FamilyName":  "Not provided",
			},
			expectedResult: resource.Result{
				FlagSet: []uint32{models.USERFLAG_ACCOUNT_CREATED},
			},
			expectedData: map[string]string{
				"TrackingId":  "test-tracking-id",
				"PublicKey":   "test-public-key",
				"CustodialId": "test-custodial-id",
				"Status":      "PENDING",
				"Gender":      "Not provided",
				"YOB":         "Not provided",
				"Location":    "Not provided",
				"Offerings":   "Not provided",
				"FirstName":   "Not provided",
				"FamilyName":  "Not provided",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the data file path using the session ID
			accountFilePath := filepath.Join(tempDir, sessionID+"_data")

			// Setup existing data if any
			if tt.existingData != nil {
				data, _ := json.Marshal(tt.existingData)
				err := os.WriteFile(accountFilePath, data, 0644)
				if err != nil {
					t.Fatalf("Failed to write existing data: %v", err)
				}
			}

			// Call the function
			result, err := h.CreateAccount(context.Background(), "", nil)

			// Check for errors
			if err != nil {
				t.Fatalf("CreateAccount returned an error: %v", err)
			}

			// Check the result
			if len(result.FlagSet) != len(tt.expectedResult.FlagSet) {
				t.Errorf("Expected %d flags, got %d", len(tt.expectedResult.FlagSet), len(result.FlagSet))
			}
			for i, flag := range tt.expectedResult.FlagSet {
				if result.FlagSet[i] != flag {
					t.Errorf("Expected flag %d, got %d", flag, result.FlagSet[i])
				}
			}

			// Check the stored data
			data, err := os.ReadFile(accountFilePath)
			if err != nil {
				t.Fatalf("Failed to read account data file: %v", err)
			}

			var storedData map[string]string
			err = json.Unmarshal(data, &storedData)
			if err != nil {
				t.Fatalf("Failed to unmarshal stored data: %v", err)
			}

			for key, expectedValue := range tt.expectedData {
				if storedValue, ok := storedData[key]; !ok || storedValue != expectedValue {
					t.Errorf("Expected %s to be %s, got %s", key, expectedValue, storedValue)
				}
			}
		})
	}
}

func TestSavePin(t *testing.T) {
	// Setup
	tempDir, err := os.MkdirTemp("", "test_save_pin")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	sessionID := "07xxxxxxxx"

	// Set up the data file path using the session ID
	accountFilePath := filepath.Join(tempDir, sessionID+"_data")
	initialAccountData := map[string]string{
		"TrackingId": "test-tracking-id",
		"PublicKey":  "test-public-key",
	}
	data, _ := json.Marshal(initialAccountData)
	err = os.WriteFile(accountFilePath, data, 0644)
	if err != nil {
		t.Fatalf("Failed to write initial account data: %v", err)
	}

	// Create a new AccountFileHandler and set it in the Handlers struct
	accountFileHandler := utils.NewAccountFileHandler(accountFilePath)
	h := &Handlers{
		accountFileHandler: accountFileHandler,
	}

	tests := []struct {
		name           string
		input          []byte
		expectedFlags  []uint32
		expectedData   map[string]string
		expectedErrors bool
	}{
		{
			name:          "Valid PIN",
			input:         []byte("1234"),
			expectedFlags: []uint32{},
			expectedData: map[string]string{
				"TrackingId": "test-tracking-id",
				"PublicKey":  "test-public-key",
				"AccountPIN": "1234",
			},
		},
		{
			name:           "Invalid PIN - non-numeric",
			input:          []byte("12ab"),
			expectedFlags:  []uint32{models.USERFLAG_INCORRECTPIN},
			expectedData:   initialAccountData, // No changes expected
			expectedErrors: false,
		},
		{
			name:           "Invalid PIN - less than 4 digits",
			input:          []byte("123"),
			expectedFlags:  []uint32{models.USERFLAG_INCORRECTPIN},
			expectedData:   initialAccountData, // No changes expected
			expectedErrors: false,
		},
		{
			name:           "Invalid PIN - more than 4 digits",
			input:          []byte("12345"),
			expectedFlags:  []uint32{models.USERFLAG_INCORRECTPIN},
			expectedData:   initialAccountData, // No changes expected
			expectedErrors: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure the file exists before running the test
			err := accountFileHandler.EnsureFileExists()
			if err != nil {
				t.Fatalf("Failed to ensure account file exists: %v", err)
			}

			result, err := h.SavePin(context.Background(), "", tt.input)
			if err != nil && !tt.expectedErrors {
				t.Fatalf("SavePin returned an unexpected error: %v", err)
			}

			if len(result.FlagSet) != len(tt.expectedFlags) {
				t.Errorf("Expected %d flags, got %d", len(tt.expectedFlags), len(result.FlagSet))
			}
			for i, flag := range tt.expectedFlags {
				if result.FlagSet[i] != flag {
					t.Errorf("Expected flag %d, got %d", flag, result.FlagSet[i])
				}
			}

			data, err := os.ReadFile(accountFilePath)
			if err != nil {
				t.Fatalf("Failed to read account data file: %v", err)
			}

			var storedData map[string]string
			err = json.Unmarshal(data, &storedData)
			if err != nil {
				t.Fatalf("Failed to unmarshal stored data: %v", err)
			}

			for key, expectedValue := range tt.expectedData {
				if storedValue, ok := storedData[key]; !ok || storedValue != expectedValue {
					t.Errorf("Expected %s to be %s, got %s", key, expectedValue, storedValue)
				}
			}
		})
	}
}
