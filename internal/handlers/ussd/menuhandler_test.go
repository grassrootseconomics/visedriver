package ussd

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"git.defalsify.org/vise.git/resource"
	"git.grassecon.net/urdt/ussd/internal/handlers/ussd/mocks"
	"git.grassecon.net/urdt/ussd/internal/models"
	"git.grassecon.net/urdt/ussd/internal/utils"
	"github.com/alecthomas/assert/v2"
	"github.com/stretchr/testify/mock"
)

// MockAccountCreator implements AccountCreator for testing
type MockAccountCreator struct {
	mockResponse *models.AccountResponse
	mockError    error
}



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
			PublicKey:   "0x8E0XSCSVA",
			TrackingId:  "d95a7e83-196c-4fd0-866fSGAGA",
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



func TestSaveLocation(t *testing.T) {
	// Create a new instance of MockAccountFileHandler
	mockFileHandler := new(mocks.MockAccountFileHandler)

	// Define test cases
	tests := []struct {
		name           string
		input          []byte
		existingData   map[string]string
		writeError     error
		expectedResult resource.Result
		expectedError  error
	}{
		{
			name:           "Successful Save",
			input:          []byte("Mombasa"),
			existingData:   map[string]string{"Location": "Mombasa"},
			writeError:     nil,
			expectedResult: resource.Result{},
			expectedError:  nil,
		},
		{
			name:           "Empty location input",
			input:          []byte{},
			existingData:   map[string]string{"OtherKey": "OtherValue"},
			writeError:     nil,
			expectedResult: resource.Result{},
			expectedError:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the mock expectations
			mockFileHandler.On("ReadAccountData").Return(tt.existingData, tt.expectedError)
			if tt.expectedError == nil && len(tt.input) > 0 {
				mockFileHandler.On("WriteAccountData", mock.MatchedBy(func(data map[string]string) bool {
					return data["Location"] == string(tt.input)
				})).Return(tt.writeError)
			}else if len(tt.input) == 0 {
				// For empty input, no WriteAccountData call should be made
				mockFileHandler.On("WriteAccountData", mock.Anything).Maybe().Return(tt.writeError)
			}

			// Create the Handlers instance with the mock file handler
			h := &Handlers{
				accountFileHandler: mockFileHandler,
			}

			// Call the method
			result, err := h.SaveLocation(context.Background(), "save_location", tt.input)

			// Assert the results
			assert.Equal(t, tt.expectedResult, result)
			assert.Equal(t, tt.expectedError, err)

			// Assert all expectations were met
			mockFileHandler.AssertExpectations(t)
		})
	}
}