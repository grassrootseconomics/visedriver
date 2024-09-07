package ussd

import (
	"context"
	"testing"

	"git.grassecon.net/urdt/ussd/internal/handlers/ussd/mocks"
	"github.com/alecthomas/assert/v2"
	"github.com/stretchr/testify/mock"
)

// MockDb is a mock implementation of the db.Db interface
//type MockDb struct {
// 	mock.Mock
// }

// func (m *MockDb) SetPrefix(prefix uint8) {
// 	m.Called(prefix)
// }

// func (m *MockDb) Prefix() uint8 {
// 	args := m.Called()
// 	return args.Get(0).(uint8)
// }

// func (m *MockDb) Safe() bool {
// 	args := m.Called()
// 	return args.Get(0).(bool)
// }

// func (m *MockDb) SetLanguage(language *lang.Language) {
// 	m.Called(language)
// }

// func (m *MockDb) SetLock(uint8, bool) error {
// 	args := m.Called()
// 	return args.Error(0)
// }

// func (m *MockDb) Connect(ctx context.Context, connectionStr string) error {
// 	args := m.Called(ctx, connectionStr)
// 	return args.Error(0)
// }

// func (m *MockDb) SetSession(sessionId string) {
// 	m.Called(sessionId)
// }

// func (m *MockDb) Put(ctx context.Context, key, value []byte) error {
// 	args := m.Called(ctx, key, value)
// 	return args.Error(0)
// }

// func (m *MockDb) Get(ctx context.Context, key []byte) ([]byte, error) {
// 	args := m.Called(ctx, key)
// 	return nil, args.Error(0)
// }

// func (m *MockDb) Close() error {
// 	args := m.Called(nil)
// 	return args.Error(0)
// }

// func TestCreateAccount(t *testing.T) {
// 	// Setup
// 	tempDir, err := os.MkdirTemp("", "test_create_account")
// 	if err != nil {
// 		t.Fatalf("Failed to create temp directory: %v", err)
// 	}
// 	defer os.RemoveAll(tempDir) // Clean up after the test run

// 	sessionID := "07xxxxxxxx"

// 	// Set up the data file path using the session ID
// 	accountFilePath := filepath.Join(tempDir, sessionID+"_data")

// 	// Initialize account file handler
// 	accountFileHandler := utils.NewAccountFileHandler(accountFilePath)

// 	// Create a mock account service
// 	mockAccountService := &MockAccountService{}
// 	mockAccountResponse := &models.AccountResponse{
// 		Ok: true,
// 		Result: struct {
// 			CustodialId json.Number `json:"custodialId"`
// 			PublicKey   string      `json:"publicKey"`
// 			TrackingId  string      `json:"trackingId"`
// 		}{
// 			CustodialId: "test-custodial-id",
// 			PublicKey:   "test-public-key",
// 			TrackingId:  "test-tracking-id",
// 		},
// 	}

// 	// Set up expectations for the mock account service
// 	mockAccountService.On("CreateAccount").Return(mockAccountResponse, nil)

// 	mockParser := new(MockFlagParser)

// 	flag_account_created := uint32(1)
// 	flag_account_creation_failed := uint32(2)

// 	mockParser.On("GetFlag", "flag_account_created").Return(flag_account_created, nil)
// 	mockParser.On("GetFlag", "flag_account_creation_failed").Return(flag_account_creation_failed, nil)

// 	// Initialize Handlers with mock account service
// 	h := &Handlers{
// 		fs:                 &FSData{Path: accountFilePath},
// 		accountFileHandler: accountFileHandler,
// 		accountService:     mockAccountService,
// 		parser:             mockParser,
// 	}

// 	tests := []struct {
// 		name           string
// 		existingData   map[string]string
// 		expectedResult resource.Result
// 		expectedData   map[string]string
// 	}{
// 		{
// 			name:         "New account creation",
// 			existingData: nil,
// 			expectedResult: resource.Result{
// 				FlagSet: []uint32{flag_account_created},
// 			},
// 			expectedData: map[string]string{
// 				"TrackingId":  "test-tracking-id",
// 				"PublicKey":   "test-public-key",
// 				"CustodialId": "test-custodial-id",
// 				"Status":      "PENDING",
// 				"Gender":      "Not provided",
// 				"YOB":         "Not provided",
// 				"Location":    "Not provided",
// 				"Offerings":   "Not provided",
// 				"FirstName":   "Not provided",
// 				"FamilyName":  "Not provided",
// 			},
// 		},
// 		{
// 			name: "Existing account",
// 			existingData: map[string]string{
// 				"TrackingId":  "test-tracking-id",
// 				"PublicKey":   "test-public-key",
// 				"CustodialId": "test-custodial-id",
// 				"Status":      "PENDING",
// 				"Gender":      "Not provided",
// 				"YOB":         "Not provided",
// 				"Location":    "Not provided",
// 				"Offerings":   "Not provided",
// 				"FirstName":   "Not provided",
// 				"FamilyName":  "Not provided",
// 			},
// 			expectedResult: resource.Result{},
// 			expectedData: map[string]string{
// 				"TrackingId":  "test-tracking-id",
// 				"PublicKey":   "test-public-key",
// 				"CustodialId": "test-custodial-id",
// 				"Status":      "PENDING",
// 				"Gender":      "Not provided",
// 				"YOB":         "Not provided",
// 				"Location":    "Not provided",
// 				"Offerings":   "Not provided",
// 				"FirstName":   "Not provided",
// 				"FamilyName":  "Not provided",
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Set up the data file path using the session ID
// 			accountFilePath := filepath.Join(tempDir, sessionID+"_data")

// 			// Setup existing data if any
// 			if tt.existingData != nil {
// 				data, _ := json.Marshal(tt.existingData)
// 				err := os.WriteFile(accountFilePath, data, 0644)
// 				if err != nil {
// 					t.Fatalf("Failed to write existing data: %v", err)
// 				}
// 			}

// 			// Call the function
// 			result, err := h.CreateAccount(context.Background(), "", nil)

// 			// Check for errors
// 			if err != nil {
// 				t.Fatalf("CreateAccount returned an error: %v", err)
// 			}

// 			// Check the result
// 			if len(result.FlagSet) != len(tt.expectedResult.FlagSet) {
// 				t.Errorf("Expected %d flags, got %d", len(tt.expectedResult.FlagSet), len(result.FlagSet))
// 			}
// 			for i, flag := range tt.expectedResult.FlagSet {
// 				if result.FlagSet[i] != flag {
// 					t.Errorf("Expected flag %d, got %d", flag, result.FlagSet[i])
// 				}
// 			}

// 			// Check the stored data
// 			data, err := os.ReadFile(accountFilePath)
// 			if err != nil {
// 				t.Fatalf("Failed to read account data file: %v", err)
// 			}

// 			var storedData map[string]string
// 			err = json.Unmarshal(data, &storedData)
// 			if err != nil {
// 				t.Fatalf("Failed to unmarshal stored data: %v", err)
// 			}

// 			for key, expectedValue := range tt.expectedData {
// 				if storedValue, ok := storedData[key]; !ok || storedValue != expectedValue {
// 					t.Errorf("Expected %s to be %s, got %s", key, expectedValue, storedValue)
// 				}
// 			}
// 		})
// 	}
// }

// func TestCreateAccount_Success(t *testing.T) {
// 	mockAccountFileHandler := new(mocks.MockAccountFileHandler)
// 	mockCreateAccountService := new(mocks.MockAccountService)

// 	mockAccountFileHandler.On("EnsureFileExists").Return(nil)

// 	// Mock that no account data exists
// 	mockAccountFileHandler.On("ReadAccountData").Return(nil, nil)

// 	// Define expected account response after api call
// 	expectedAccountResp := &models.AccountResponse{
// 		Ok: true,
// 		Result: struct {
// 			CustodialId json.Number `json:"custodialId"`
// 			PublicKey   string      `json:"publicKey"`
// 			TrackingId  string      `json:"trackingId"`
// 		}{
// 			CustodialId: "12",
// 			PublicKey:   "0x8E0XSCSVA",
// 			TrackingId:  "d95a7e83-196c-4fd0-866fSGAGA",
// 		},
// 	}
// 	mockCreateAccountService.On("CreateAccount").Return(expectedAccountResp, nil)

// 	// Mock WriteAccountData to not error
// 	mockAccountFileHandler.On("WriteAccountData", mock.Anything).Return(nil)

// 	handlers := &Handlers{
// 		accountService: mockCreateAccountService,
// 	}

// 	actualResponse, err := handlers.accountService.CreateAccount()

// 	// Assert results
// 	assert.NoError(t, err)
// 	assert.Equal(t, expectedAccountResp.Ok, true)
// 	assert.Equal(t, expectedAccountResp, actualResponse)
// }

// func TestSavePin(t *testing.T) {
// 	// Setup
// 	tempDir, err := os.MkdirTemp("", "test_save_pin")
// 	if err != nil {
// 		t.Fatalf("Failed to create temp directory: %v", err)
// 	}
// 	defer os.RemoveAll(tempDir)

// 	sessionID := "07xxxxxxxx"

// 	// Set up the data file path using the session ID
// 	accountFilePath := filepath.Join(tempDir, sessionID+"_data")
// 	initialAccountData := map[string]string{
// 		"TrackingId": "test-tracking-id",
// 		"PublicKey":  "test-public-key",
// 	}
// 	data, _ := json.Marshal(initialAccountData)
// 	err = os.WriteFile(accountFilePath, data, 0644)
// 	if err != nil {
// 		t.Fatalf("Failed to write initial account data: %v", err)
// 	}

// 	// Create a new AccountFileHandler and set it in the Handlers struct
// 	accountFileHandler := utils.NewAccountFileHandler(accountFilePath)
// 	mockParser := new(MockFlagParser)

// 	h := &Handlers{
// 		accountFileHandler: accountFileHandler,
// 		parser:             mockParser,
// 	}

// 	flag_incorrect_pin := uint32(1)
// 	mockParser.On("GetFlag", "flag_incorrect_pin").Return(flag_incorrect_pin, nil)

// 	tests := []struct {
// 		name           string
// 		input          []byte
// 		expectedFlags  []uint32
// 		expectedData   map[string]string
// 		expectedErrors bool
// 	}{
// 		{
// 			name:          "Valid PIN",
// 			input:         []byte("1234"),
// 			expectedFlags: []uint32{},
// 			expectedData: map[string]string{
// 				"TrackingId": "test-tracking-id",
// 				"PublicKey":  "test-public-key",
// 				"AccountPIN": "1234",
// 			},
// 		},
// 		{
// 			name:           "Invalid PIN - non-numeric",
// 			input:          []byte("12ab"),
// 			expectedFlags:  []uint32{flag_incorrect_pin},
// 			expectedData:   initialAccountData, // No changes expected
// 			expectedErrors: false,
// 		},
// 		{
// 			name:           "Invalid PIN - less than 4 digits",
// 			input:          []byte("123"),
// 			expectedFlags:  []uint32{flag_incorrect_pin},
// 			expectedData:   initialAccountData, // No changes expected
// 			expectedErrors: false,
// 		},
// 		{
// 			name:           "Invalid PIN - more than 4 digits",
// 			input:          []byte("12345"),
// 			expectedFlags:  []uint32{flag_incorrect_pin},
// 			expectedData:   initialAccountData, // No changes expected
// 			expectedErrors: false,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			err := accountFileHandler.EnsureFileExists()
// 			if err != nil {
// 				t.Fatalf("Failed to ensure account file exists: %v", err)
// 			}

// 			result, err := h.SavePin(context.Background(), "", tt.input)
// 			if err != nil && !tt.expectedErrors {
// 				t.Fatalf("SavePin returned an unexpected error: %v", err)
// 			}

// 			if len(result.FlagSet) != len(tt.expectedFlags) {
// 				t.Errorf("Expected %d flags, got %d", len(tt.expectedFlags), len(result.FlagSet))
// 			}
// 			for i, flag := range tt.expectedFlags {
// 				if result.FlagSet[i] != flag {
// 					t.Errorf("Expected flag %d, got %d", flag, result.FlagSet[i])
// 				}
// 			}

// 			data, err := os.ReadFile(accountFilePath)
// 			if err != nil {
// 				t.Fatalf("Failed to read account data file: %v", err)
// 			}

// 			var storedData map[string]string
// 			err = json.Unmarshal(data, &storedData)
// 			if err != nil {
// 				t.Fatalf("Failed to unmarshal stored data: %v", err)
// 			}

// 			for key, expectedValue := range tt.expectedData {
// 				if storedValue, ok := storedData[key]; !ok || storedValue != expectedValue {
// 					t.Errorf("Expected %s to be %s, got %s", key, expectedValue, storedValue)
// 				}
// 			}
// 		})
// 	}
// }

// func TestSaveLocation(t *testing.T) {
// 	// Create a new instance of MockAccountFileHandler
// 	mockFileHandler := new(mocks.MockAccountFileHandler)

// 	// Define test cases
// 	tests := []struct {
// 		name           string
// 		input          []byte
// 		existingData   map[string]string
// 		writeError     error
// 		expectedResult resource.Result
// 		expectedError  error
// 	}{
// 		{
// 			name:           "Successful Save",
// 			input:          []byte("Mombasa"),
// 			existingData:   map[string]string{"Location": "Mombasa"},
// 			writeError:     nil,
// 			expectedResult: resource.Result{},
// 			expectedError:  nil,
// 		},
// 		{
// 			name:           "Empty location input",
// 			input:          []byte{},
// 			existingData:   map[string]string{"OtherKey": "OtherValue"},
// 			writeError:     nil,
// 			expectedResult: resource.Result{},
// 			expectedError:  nil,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Set up the mock expectations
// 			mockFileHandler.On("ReadAccountData").Return(tt.existingData, tt.expectedError)
// 			if tt.expectedError == nil && len(tt.input) > 0 {
// 				mockFileHandler.On("WriteAccountData", mock.MatchedBy(func(data map[string]string) bool {
// 					return data["Location"] == string(tt.input)
// 				})).Return(tt.writeError)
// 			} else if len(tt.input) == 0 {
// 				// For empty input, no WriteAccountData call should be made
// 				mockFileHandler.On("WriteAccountData", mock.Anything).Maybe().Return(tt.writeError)
// 				return
// 			}

// 			// Create the Handlers instance with the mock file handler
// 			h := &Handlers{
// 				accountFileHandler: mockFileHandler,
// 			}

// 			// Call Save Location
// 			result, err := h.SaveLocation(context.Background(), "save_location", tt.input)

// 			if err != nil {
// 				t.Fatalf("Failed to save location with error: %v", err)
// 			}

// 			savedData, err := h.accountFileHandler.ReadAccountData()
// 			if err == nil {
// 				//Assert that the input provided is what was saved into the file
// 				assert.Equal(t, string(tt.input), savedData["Location"])
// 			}

// 			// Assert the results
// 			assert.Equal(t, tt.expectedResult, result)
// 			assert.Equal(t, tt.expectedError, err)

// 			// Assert all expectations were met
// 			mockFileHandler.AssertExpectations(t)
// 		})
// 	}
// }

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

// func TestSaveYOB(t *testing.T) {
// 	// Create a new instance of MockAccountFileHandler
// 	mockFileHandler := new(mocks.MockAccountFileHandler)

// 	// Define test cases
// 	tests := []struct {
// 		name           string
// 		input          []byte
// 		existingData   map[string]string
// 		writeError     error
// 		expectedResult resource.Result
// 		expectedError  error
// 	}{
// 		{
// 			name:           "Successful Save",
// 			input:          []byte("2006"),
// 			existingData:   map[string]string{"": ""},
// 			writeError:     nil,
// 			expectedResult: resource.Result{},
// 			expectedError:  nil,
// 		},
// 		{
// 			name:          "YOB less than 4 digits(invalid date entry)",
// 			input:         []byte{},
// 			existingData:  map[string]string{"": ""},
// 			writeError:    nil,
// 			expectedError: nil,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Set up the mock expectations
// 			mockFileHandler.On("ReadAccountData").Return(tt.existingData, tt.expectedError)
// 			if tt.expectedError == nil && len(tt.input) > 0 {
// 				mockFileHandler.On("WriteAccountData", mock.MatchedBy(func(data map[string]string) bool {
// 					return data["YOB"] == string(tt.input)
// 				})).Return(tt.writeError)
// 			} else if len(tt.input) != 4 {
// 				// For input whose input is not a valid yob, no WriteAccountData call should be made
// 				mockFileHandler.On("WriteAccountData", mock.Anything).Maybe().Return(tt.writeError)
// 				return
// 			}

// 			// Create the Handlers instance with the mock file handler
// 			h := &Handlers{
// 				accountFileHandler: mockFileHandler,
// 			}

// 			// Call save yob
// 			result, err := h.SaveYob(context.Background(), "save_yob", tt.input)

// 			if err != nil {
// 				t.Fatalf("Failed to save family name with error: %v", err)
// 			}
// 			savedData, err := h.accountFileHandler.ReadAccountData()
// 			if err == nil {
// 				//Assert that the input provided is what was saved into the file
// 				assert.Equal(t, string(tt.input), savedData["YOB"])
// 			}

// 			// Assert the results
// 			assert.Equal(t, tt.expectedResult, result)
// 			assert.Equal(t, tt.expectedError, err)

// 			// Assert all expectations were met
// 			mockFileHandler.AssertExpectations(t)
// 		})
// 	}
// }

// func TestSaveOfferings(t *testing.T) {
// 	// Create a new instance of MockAccountFileHandler
// 	mockFileHandler := new(mocks.MockAccountFileHandler)

// 	// Define test cases
// 	tests := []struct {
// 		name           string
// 		input          []byte
// 		existingData   map[string]string
// 		writeError     error
// 		expectedResult resource.Result
// 		expectedError  error
// 	}{
// 		{
// 			name:           "Successful Save",
// 			input:          []byte("Bananas"),
// 			existingData:   map[string]string{"": ""},
// 			writeError:     nil,
// 			expectedResult: resource.Result{},
// 			expectedError:  nil,
// 		},
// 		{
// 			name:          "Empty input",
// 			input:         []byte{},
// 			existingData:  map[string]string{"": ""},
// 			writeError:    nil,
// 			expectedError: nil,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Set up the mock expectations
// 			mockFileHandler.On("ReadAccountData").Return(tt.existingData, tt.expectedError)
// 			if tt.expectedError == nil && len(tt.input) > 0 {
// 				mockFileHandler.On("WriteAccountData", mock.MatchedBy(func(data map[string]string) bool {
// 					return data["Offerings"] == string(tt.input)
// 				})).Return(tt.writeError)
// 			} else if len(tt.input) != 4 {
// 				// For input whose input is not a valid yob, no WriteAccountData call should be made
// 				mockFileHandler.On("WriteAccountData", mock.Anything).Maybe().Return(tt.writeError)
// 				return
// 			}

// 			// Create the Handlers instance with the mock file handler
// 			h := &Handlers{
// 				accountFileHandler: mockFileHandler,
// 			}

// 			// Call save yob
// 			result, err := h.SaveOfferings(context.Background(), "save_offerings", tt.input)

// 			if err != nil {
// 				t.Fatalf("Failed to save offerings with error: %v", err)
// 			}
// 			savedData, err := h.accountFileHandler.ReadAccountData()
// 			if err == nil {
// 				//Assert that the input provided is what was saved into the file
// 				assert.Equal(t, string(tt.input), savedData["Offerings"])
// 			}

// 			// Assert the results
// 			assert.Equal(t, tt.expectedResult, result)
// 			assert.Equal(t, tt.expectedError, err)

// 			// Assert all expectations were met
// 			mockFileHandler.AssertExpectations(t)
// 		})
// 	}
// }

// func TestSaveGender(t *testing.T) {
// 	// Create a new instance of MockAccountFileHandler
// 	mockFileHandler := new(mocks.MockAccountFileHandler)

// 	// Define test cases
// 	tests := []struct {
// 		name           string
// 		input          []byte
// 		existingData   map[string]string
// 		writeError     error
// 		expectedResult resource.Result
// 		expectedError  error
// 		expectedGender string
// 	}{
// 		{
// 			name:           "Successful Save - Male",
// 			input:          []byte("1"),
// 			existingData:   map[string]string{"OtherKey": "OtherValue"},
// 			writeError:     nil,
// 			expectedResult: resource.Result{},
// 			expectedError:  nil,
// 			expectedGender: "Male",
// 		},
// 		{
// 			name:           "Successful Save - Female",
// 			input:          []byte("2"),
// 			existingData:   map[string]string{"OtherKey": "OtherValue"},
// 			writeError:     nil,
// 			expectedResult: resource.Result{},
// 			expectedError:  nil,
// 			expectedGender: "Female",
// 		},
// 		{
// 			name:           "Successful Save - Unspecified",
// 			input:          []byte("3"),
// 			existingData:   map[string]string{"OtherKey": "OtherValue"},
// 			writeError:     nil,
// 			expectedResult: resource.Result{},
// 			expectedError:  nil,
// 			expectedGender: "Unspecified",
// 		},

// 		{
// 			name:           "Empty Input",
// 			input:          []byte{},
// 			existingData:   map[string]string{"OtherKey": "OtherValue"},
// 			writeError:     nil,
// 			expectedResult: resource.Result{},
// 			expectedError:  nil,
// 			expectedGender: "",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Set up the mock expectations
// 			mockFileHandler.On("ReadAccountData").Return(tt.existingData, tt.expectedError)
// 			if tt.expectedError == nil && len(tt.input) > 0 {
// 				mockFileHandler.On("WriteAccountData", mock.MatchedBy(func(data map[string]string) bool {
// 					return data["Gender"] == tt.expectedGender
// 				})).Return(tt.writeError)
// 			} else if len(tt.input) == 0 {
// 				// For empty input, no WriteAccountData call should be made
// 				mockFileHandler.On("WriteAccountData", mock.Anything).Maybe().Return(tt.writeError)
// 			}

// 			// Create the Handlers instance with the mock file handler
// 			h := &Handlers{
// 				accountFileHandler: mockFileHandler,
// 			}

// 			// Call the method
// 			result, err := h.SaveGender(context.Background(), "save_gender", tt.input)

// 			// Assert the results
// 			assert.Equal(t, tt.expectedResult, result)
// 			assert.Equal(t, tt.expectedError, err)

// 			// Verify WriteAccountData was called with the expected data
// 			if len(tt.input) > 0 && tt.expectedError == nil {
// 				mockFileHandler.AssertCalled(t, "WriteAccountData", mock.MatchedBy(func(data map[string]string) bool {
// 					return data["Gender"] == tt.expectedGender
// 				}))
// 			}

// 			// Assert all expectations were met
// 			mockFileHandler.AssertExpectations(t)
// 		})
// 	}
// }

// func TestGetSender(t *testing.T) {
// 	mockAccountFileHandler := new(mocks.MockAccountFileHandler)
// 	h := &Handlers{
// 		accountFileHandler: mockAccountFileHandler,
// 	}

// 	tests := []struct {
// 		name           string
// 		expectedResult resource.Result
// 		accountData    map[string]string
// 	}{
// 		{
// 			name: "Valid public key",
// 			expectedResult: resource.Result{
// 				Content: "test-public-key",
// 			},
// 			accountData: map[string]string{
// 				"PublicKey": "test-public-key",
// 			},
// 		},
// 		{
// 			name: "Missing public key",
// 			expectedResult: resource.Result{
// 				Content: "",
// 			},
// 			accountData: map[string]string{},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Reset the mock state
// 			mockAccountFileHandler.Mock = mock.Mock{}

// 			mockAccountFileHandler.On("ReadAccountData").Return(tt.accountData, nil)

// 			result, err := h.GetSender(context.Background(), "", nil)

// 			if err != nil {
// 				t.Fatalf("Error occurred: %v", err)
// 			}

// 			assert.Equal(t, tt.expectedResult.Content, result.Content)
// 			mockAccountFileHandler.AssertCalled(t, "ReadAccountData")
// 		})
// 	}
// }

// func TestGetAmount(t *testing.T) {
// 	mockAccountFileHandler := new(mocks.MockAccountFileHandler)
// 	h := &Handlers{
// 		accountFileHandler: mockAccountFileHandler,
// 	}

// 	tests := []struct {
// 		name           string
// 		expectedResult resource.Result
// 		accountData    map[string]string
// 	}{
// 		{
// 			name: "Valid amount",
// 			expectedResult: resource.Result{
// 				Content: "0.003",
// 			},
// 			accountData: map[string]string{
// 				"Amount": "0.003",
// 			},
// 		},
// 		{
// 			name:           "Missing amount",
// 			expectedResult: resource.Result{},
// 			accountData: map[string]string{
// 				"Amount": "",
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Reset the mock state
// 			mockAccountFileHandler.Mock = mock.Mock{}

// 			mockAccountFileHandler.On("ReadAccountData").Return(tt.accountData, nil)

// 			result, err := h.GetAmount(context.Background(), "", nil)

// 			assert.NoError(t, err)
// 			assert.Equal(t, tt.expectedResult.Content, result.Content)

// 			mockAccountFileHandler.AssertCalled(t, "ReadAccountData")
// 		})
// 	}
// }

// func TestGetProfileInfo(t *testing.T) {
// 	tests := []struct {
// 		name           string
// 		accountData    map[string]string
// 		readError      error
// 		expectedResult resource.Result
// 		expectedError  error
// 	}{
// 		{
// 			name: "Complete Profile",
// 			accountData: map[string]string{
// 				"FirstName":  "John",
// 				"FamilyName": "Doe",
// 				"Gender":     "Male",
// 				"YOB":        "1980",
// 				"Location":   "Mombasa",
// 				"Offerings":  "Product A",
// 			},
// 			readError: nil,
// 			expectedResult: resource.Result{
// 				Content: fmt.Sprintf(
// 					"Name: %s %s\nGender: %s\nAge: %d\nLocation: %s\nYou provide: %s\n",
// 					"John", "Doe", "Male", 44, "Mombasa", "Product A",
// 				),
// 			},
// 			expectedError: nil,
// 		},
// 		{
// 			name: "Profile with Not Provided Fields",
// 			accountData: map[string]string{
// 				"FirstName":  "Not provided",
// 				"FamilyName": "Doe",
// 				"Gender":     "Female",
// 				"YOB":        "1995",
// 				"Location":   "Not provided",
// 				"Offerings":  "Service B",
// 			},
// 			readError: nil,
// 			expectedResult: resource.Result{
// 				Content: fmt.Sprintf(
// 					"Name: %s\nGender: %s\nAge: %d\nLocation: %s\nYou provide: %s\n",
// 					"Not provided", "Female", 29, "Not provided", "Service B",
// 				),
// 			},
// 			expectedError: nil,
// 		},
// 		{
// 			name: "Profile with YOB as Not provided",
// 			accountData: map[string]string{
// 				"FirstName":  "Not provided",
// 				"FamilyName": "Doe",
// 				"Gender":     "Female",
// 				"YOB":        "Not provided",
// 				"Location":   "Not provided",
// 				"Offerings":  "Service B",
// 			},
// 			readError: nil,
// 			expectedResult: resource.Result{
// 				Content: fmt.Sprintf(
// 					"Name: %s\nGender: %s\nAge: %s\nLocation: %s\nYou provide: %s\n",
// 					"Not provided", "Female", "Not provided", "Not provided", "Service B",
// 				),
// 			},
// 			expectedError: nil,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Create a new instance of MockAccountFileHandler
// 			mockFileHandler := new(mocks.MockAccountFileHandler)

// 			// Set up the mock expectations
// 			mockFileHandler.On("ReadAccountData").Return(tt.accountData, tt.readError)

// 			// Create the Handlers instance with the mock file handler
// 			h := &Handlers{
// 				accountFileHandler: mockFileHandler,
// 			}

// 			// Call the method
// 			result, err := h.GetProfileInfo(context.Background(), "get_profile_info", nil)

// 			// Assert the results
// 			assert.Equal(t, tt.expectedResult, result)
// 			assert.Equal(t, tt.expectedError, err)

// 			// Assert all expectations were met
// 			mockFileHandler.AssertExpectations(t)
// 		})
// 	}
// }
