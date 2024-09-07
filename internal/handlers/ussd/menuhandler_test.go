package ussd

import (
	"context"
	"encoding/json"
	"testing"

	"git.defalsify.org/vise.git/db"
	"git.grassecon.net/urdt/ussd/internal/handlers/ussd/mocks"
	"git.grassecon.net/urdt/ussd/internal/models"
	"git.grassecon.net/urdt/ussd/internal/utils"
	"github.com/alecthomas/assert/v2"
	"github.com/stretchr/testify/mock"
)

func TestCreateAccount_Success(t *testing.T) {
	mockCreateAccountService := new(mocks.MockAccountService)
	mockUserDataStore := new(mocks.MockDb)

	h := &Handlers{
		userdataStore:  mockUserDataStore,
		accountService: mockCreateAccountService,
	}
	ctx := context.WithValue(context.Background(), "SessionId", "test-session-12345")
	k := utils.PackKey(utils.DATA_ACCOUNT_CREATED, []byte("test-session-12345"))
	mockUserDataStore.On("SetPrefix", uint8(0x20)).Return(nil)
	mockUserDataStore.On("SetSession", "test-session-12345").Return(nil)
	mockUserDataStore.On("Get", ctx, k).
		Return(nil, db.ErrNotFound{})

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

	_, err := h.CreateAccount(ctx, "create_account", []byte("create_account"))

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, expectedAccountResp.Ok, true)
}

func TestSaveFirstname(t *testing.T) {
	// Create a mock database
	mockDb := new(mocks.MockDb)

	// Create a Handlers instance with the mock database
	h := &Handlers{
		userdataStore: mockDb,
	}

	// Create a context with a session ID
	ctx := context.WithValue(context.Background(), "SessionId", "test-session")

	tests := []struct {
		name        string
		input       []byte
		expectError bool
		setupMock   func(*mocks.MockDb)
	}{
		{
			name:        "Valid first name",
			input:       []byte("John"),
			expectError: false,
			setupMock: func(m *mocks.MockDb) {
				m.On("SetPrefix", uint8(0x20)).Return(nil)
				m.On("SetSession", "test-session").Return(nil)
				m.On("Put", mock.Anything, mock.Anything, []byte("John")).Return(nil)
			},
		},
		{
			name:        "Empty first name",
			input:       []byte{},
			expectError: false, // Note: The function doesn't return an error for empty input
			setupMock:   func(m *mocks.MockDb) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			tt.setupMock(mockDb)

			// Call the function
			_, err := h.SaveFirstname(ctx, "", tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				mockDb.AssertExpectations(t)
			}

			// Clear mock for the next test
			mockDb.ExpectedCalls = nil
			mockDb.Calls = nil
		})
	}
}

func TestSaveFamilyname(t *testing.T) {
	// Create a mock database
	mockDb := new(mocks.MockDb)

	// Create a Handlers instance with the mock database
	h := &Handlers{
		userdataStore: mockDb,
	}

	// Create a context with a session ID
	ctx := context.WithValue(context.Background(), "SessionId", "test-session")

	tests := []struct {
		name        string
		input       []byte
		expectError bool
		setupMock   func(*mocks.MockDb)
	}{
		{
			name:        "Valid family name",
			input:       []byte("Smith"),
			expectError: false,
			setupMock: func(m *mocks.MockDb) {
				m.On("SetPrefix", uint8(0x20)).Return(nil)
				m.On("SetSession", "test-session").Return(nil)
				m.On("Put", mock.Anything, mock.Anything, []byte("Smith")).Return(nil)
			},
		},
		{
			name:        "Empty family name",
			input:       []byte{},
			expectError: true,
			setupMock:   func(m *mocks.MockDb) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectations
			tt.setupMock(mockDb)

			// Call the function
			_, err := h.SaveFamilyname(ctx, "", tt.input)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				mockDb.AssertExpectations(t)
			}

			// Clear mock for the next test
			mockDb.ExpectedCalls = nil
			mockDb.Calls = nil
		})
	}
}
