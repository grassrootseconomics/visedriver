package ussd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path"
	"testing"

	"git.defalsify.org/vise.git/db"
	"git.defalsify.org/vise.git/persist"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
	"git.grassecon.net/urdt/ussd/internal/mocks"
	"git.grassecon.net/urdt/ussd/internal/models"
	"git.grassecon.net/urdt/ussd/internal/utils"
	"github.com/alecthomas/assert/v2"
	testdataloader "github.com/peteole/testdata-loader"
	"github.com/stretchr/testify/require"
)

var (
	baseDir   = testdataloader.GetBasePath()
	flagsPath = path.Join(baseDir, "services", "registration", "pp.csv")
)

func TestCreateAccount(t *testing.T) {

	fm, err := NewFlagManager(flagsPath)

	if err != nil {
		t.Logf(err.Error())
	}

	// Create required mocks
	mockDataStore := new(mocks.MockUserDataStore)
	mockCreateAccountService := new(mocks.MockAccountService)
	expectedResult := resource.Result{}
	accountCreatedFlag, err := fm.GetFlag("flag_account_created")

	if err != nil {
		t.Logf(err.Error())
	}
	expectedResult.FlagSet = append(expectedResult.FlagSet, accountCreatedFlag)

	// Define session ID and mock data
	sessionId := "session123"
	typ := utils.DATA_ACCOUNT_CREATED
	fakeError := db.ErrNotFound{}
	// Create context with session ID
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	// Define expected interactions with the mock
	mockDataStore.On("ReadEntry", ctx, sessionId, typ).Return([]byte("123"), fakeError)
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
	data := map[utils.DataTyp]string{
		utils.DATA_TRACKING_ID:  expectedAccountResp.Result.TrackingId,
		utils.DATA_PUBLIC_KEY:   expectedAccountResp.Result.PublicKey,
		utils.DATA_CUSTODIAL_ID: expectedAccountResp.Result.CustodialId.String(),
	}

	for key, value := range data {
		mockDataStore.On("WriteEntry", ctx, sessionId, key, []byte(value)).Return(nil)
	}

	// Create a Handlers instance with the mock data store
	h := &Handlers{
		userdataStore:  mockDataStore,
		accountService: mockCreateAccountService,
		flagManager:    fm.parser,
	}

	// Call the method you want to test
	res, err := h.CreateAccount(ctx, "create_account", []byte("some-input"))

	// Assert that no errors occurred
	assert.NoError(t, err)

	//Assert that the account created flag has been set to the result
	assert.Equal(t, res, expectedResult, "Expected result should be equal to the actual result")

	// Assert that expectations were met
	mockDataStore.AssertExpectations(t)
}

func TestWithPersister(t *testing.T) {
	// Test case: Setting a persister
	h := &Handlers{}
	p := &persist.Persister{}

	result := h.WithPersister(p)

	assert.Equal(t, p, h.pe, "The persister should be set correctly.")
	assert.Equal(t, h, result, "The returned handler should be the same instance.")
}

func TestWithPersister_PanicWhenAlreadySet(t *testing.T) {
	// Test case: Panic on multiple calls
	h := &Handlers{pe: &persist.Persister{}}
	require.Panics(t, func() {
		h.WithPersister(&persist.Persister{})
	}, "Should panic when trying to set a persister again.")
}

func TestSaveFirstname(t *testing.T) {
	// Create a new instance of MockMyDataStore
	mockStore := new(mocks.MockUserDataStore)

	// Define test data
	sessionId := "session123"
	firstName := "John"
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	// Set up the expected behavior of the mock
	mockStore.On("WriteEntry", ctx, sessionId, utils.DATA_FIRST_NAME, []byte(firstName)).Return(nil)

	// Create the Handlers instance with the mock store
	h := &Handlers{
		userdataStore: mockStore,
	}

	// Call the method
	res, err := h.SaveFirstname(ctx, "save_firstname", []byte(firstName))

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, resource.Result{}, res)

	// Assert all expectations were met
	mockStore.AssertExpectations(t)
}

func TestSaveFamilyname(t *testing.T) {
	// Create a new instance of UserDataStore
	mockStore := new(mocks.MockUserDataStore)

	// Define test data
	sessionId := "session123"
	familyName := "Doeee"
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	// Set up the expected behavior of the mock
	mockStore.On("WriteEntry", ctx, sessionId, utils.DATA_FAMILY_NAME, []byte(familyName)).Return(nil)

	// Create the Handlers instance with the mock store
	h := &Handlers{
		userdataStore: mockStore,
	}

	// Call the method
	res, err := h.SaveFamilyname(ctx, "save_familyname", []byte(familyName))

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, resource.Result{}, res)

	// Assert all expectations were met
	mockStore.AssertExpectations(t)
}

func TestSaveTemporaryPIn(t *testing.T) {
	fm, err := NewFlagManager(flagsPath)
	mockStore := new(mocks.MockUserDataStore)
	if err != nil {
		log.Fatal(err)
	}
	flag_incorrect_pin, _ := fm.parser.GetFlag("flag_incorrect_pin")

	// Create the Handlers instance with the mock flag manager
	h := &Handlers{
		flagManager:   fm.parser,
		userdataStore: mockStore,
	}
	sessionId := "session123"
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	// Define test cases
	tests := []struct {
		name           string
		input          []byte
		expectedResult resource.Result
	}{
		{
			name:  "Valid Pin entry",
			input: []byte("1234"),
			expectedResult: resource.Result{
				FlagReset: []uint32{flag_incorrect_pin},
			},
		},
		{
			name:  "Invalid Pin entry",
			input: []byte("12343"),
			expectedResult: resource.Result{
				FlagSet: []uint32{flag_incorrect_pin},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Set up the expected behavior of the mock
			mockStore.On("WriteEntry", ctx, sessionId, utils.DATA_TEMPORARY_PIN, []byte(tt.input)).Return(nil)

			// Call the method
			res, err := h.SaveTemporaryPin(ctx, "save_pin", tt.input)

			if err != nil {
				t.Error(err)
			}

			// Assert that the Result FlagSet has the required flags after language switch
			assert.Equal(t, res, tt.expectedResult, "Flags should be equal to account created")

		})
	}
}

func TestSaveYoB(t *testing.T) {
	// Create a new instance of MockMyDataStore
	mockStore := new(mocks.MockUserDataStore)

	// Define test data
	sessionId := "session123"
	yob := "1980"
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	// Set up the expected behavior of the mock
	mockStore.On("WriteEntry", ctx, sessionId, utils.DATA_YOB, []byte(yob)).Return(nil)

	// Create the Handlers instance with the mock store
	h := &Handlers{
		userdataStore: mockStore,
	}

	// Call the method
	res, err := h.SaveYob(ctx, "save_yob", []byte(yob))

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, resource.Result{}, res)

	// Assert all expectations were met
	mockStore.AssertExpectations(t)
}

func TestSaveLocation(t *testing.T) {
	// Create a new instance of MockMyDataStore
	mockStore := new(mocks.MockUserDataStore)

	// Define test data
	sessionId := "session123"
	yob := "Kilifi"
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	// Set up the expected behavior of the mock
	mockStore.On("WriteEntry", ctx, sessionId, utils.DATA_LOCATION, []byte(yob)).Return(nil)

	// Create the Handlers instance with the mock store
	h := &Handlers{
		userdataStore: mockStore,
	}

	// Call the method
	res, err := h.SaveLocation(ctx, "save_location", []byte(yob))

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, resource.Result{}, res)

	// Assert all expectations were met
	mockStore.AssertExpectations(t)
}

func TestSaveOfferings(t *testing.T) {
	// Create a new instance of MockUserDataStore
	mockStore := new(mocks.MockUserDataStore)

	// Define test data
	sessionId := "session123"
	offerings := "Bananas"
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	// Set up the expected behavior of the mock
	mockStore.On("WriteEntry", ctx, sessionId, utils.DATA_OFFERINGS, []byte(offerings)).Return(nil)

	// Create the Handlers instance with the mock store
	h := &Handlers{
		userdataStore: mockStore,
	}

	// Call the method
	res, err := h.SaveOfferings(ctx, "save_offerings", []byte(offerings))

	// Assert results
	assert.NoError(t, err)
	assert.Equal(t, resource.Result{}, res)

	// Assert all expectations were met
	mockStore.AssertExpectations(t)
}

func TestSaveGender(t *testing.T) {
	// Create a new instance of MockMyDataStore
	mockStore := new(mocks.MockUserDataStore)
	mockState := state.NewState(16)

	// Define the session ID and context
	sessionId := "session123"
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	// Define test cases
	tests := []struct {
		name            string
		input           []byte
		expectedGender  string
		expectCall      bool
		executingSymbol string
	}{
		{
			name:            "Valid Male Input",
			input:           []byte("1"),
			expectedGender:  "male",
			executingSymbol: "set_male",
			expectCall:      true,
		},
		{
			name:            "Valid Female Input",
			input:           []byte("2"),
			expectedGender:  "female",
			executingSymbol: "set_female",
			expectCall:      true,
		},
		{
			name:            "Valid Unspecified Input",
			input:           []byte("3"),
			executingSymbol: "set_unspecified",
			expectedGender:  "unspecified",
			expectCall:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up expectations for the mock database
			if tt.expectCall {
				expectedKey := utils.DATA_GENDER
				mockStore.On("WriteEntry", ctx, sessionId, expectedKey, []byte(tt.expectedGender)).Return(nil)
			} else {
				mockStore.On("WriteEntry", ctx, sessionId, utils.DATA_GENDER, []byte(tt.expectedGender)).Return(nil)
			}
			mockState.ExecPath = append(mockState.ExecPath, tt.executingSymbol)
			// Create the Handlers instance with the mock store
			h := &Handlers{
				userdataStore: mockStore,
				st:            mockState,
			}

			// Call the method
			_, err := h.SaveGender(ctx, "save_gender", tt.input)

			// Assert no error
			assert.NoError(t, err)

			// Verify expectations
			if tt.expectCall {
				mockStore.AssertCalled(t, "WriteEntry", ctx, sessionId, utils.DATA_GENDER, []byte(tt.expectedGender))
			} else {
				mockStore.AssertNotCalled(t, "WriteEntry", ctx, sessionId, utils.DATA_GENDER, []byte(tt.expectedGender))
			}
		})
	}
}

func TestCheckIdentifier(t *testing.T) {
	// Create a new instance of MockMyDataStore
	mockStore := new(mocks.MockUserDataStore)

	// Define the session ID and context
	sessionId := "session123"
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	// Define test cases
	tests := []struct {
		name            string
		mockPublicKey   []byte
		mockErr         error
		expectedContent string
		expectError     bool
	}{
		{
			name:            "Saved public Key",
			mockPublicKey:   []byte("0xa8363"),
			mockErr:         nil,
			expectedContent: "0xa8363",
			expectError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up expectations for the mock database
			mockStore.On("ReadEntry", ctx, sessionId, utils.DATA_PUBLIC_KEY).Return(tt.mockPublicKey, tt.mockErr)

			// Create the Handlers instance with the mock store
			h := &Handlers{
				userdataStore: mockStore,
			}

			// Call the method
			res, err := h.CheckIdentifier(ctx, "check_identifier", nil)

			// Assert results

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedContent, res.Content)

			// Verify expectations
			mockStore.AssertExpectations(t)
		})
	}
}

func TestGetSender(t *testing.T) {
	mockStore := new(mocks.MockUserDataStore)

	// Define test data
	sessionId := "session123"
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)
	publicKey := "0xcasgatweksalw1018221"

	// Set up the expected behavior of the mock
	mockStore.On("ReadEntry", ctx, sessionId, utils.DATA_PUBLIC_KEY).Return([]byte(publicKey), nil)

	// Create the Handlers instance with the mock store
	h := &Handlers{
		userdataStore: mockStore,
	}

	// Call the method
	res, _ := h.GetSender(ctx, "max_amount", []byte("check_balance"))

	//Assert that the public key from readentry operation  is what was set as the result content.
	assert.Equal(t, publicKey, res.Content)

}

func TestGetAmount(t *testing.T) {
	mockStore := new(mocks.MockUserDataStore)

	// Define test data
	sessionId := "session123"
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)
	Amount := "0.03CELO"

	// Set up the expected behavior of the mock
	mockStore.On("ReadEntry", ctx, sessionId, utils.DATA_AMOUNT).Return([]byte(Amount), nil)

	// Create the Handlers instance with the mock store
	h := &Handlers{
		userdataStore: mockStore,
	}

	// Call the method
	res, _ := h.GetAmount(ctx, "get_amount", []byte("Getting amount..."))

	//Assert that the retrieved amount is what was set as the content
	assert.Equal(t, Amount, res.Content)

}

func TestGetRecipient(t *testing.T) {
	mockStore := new(mocks.MockUserDataStore)

	// Define test data
	sessionId := "session123"
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)
	recepient := "0xcasgatweksalw1018221"

	// Set up the expected behavior of the mock
	mockStore.On("ReadEntry", ctx, sessionId, utils.DATA_RECIPIENT).Return([]byte(recepient), nil)

	// Create the Handlers instance with the mock store
	h := &Handlers{
		userdataStore: mockStore,
	}

	// Call the method
	res, _ := h.GetRecipient(ctx, "get_recipient", []byte("Getting recipient..."))

	//Assert that the retrieved recepient is what was set as the content
	assert.Equal(t, recepient, res.Content)

}

func TestGetFlag(t *testing.T) {
	fm, err := NewFlagManager(flagsPath)
	expectedFlag := uint32(9)

	if err != nil {
		t.Logf(err.Error())
	}
	flag, err := fm.GetFlag("flag_account_created")

	if err != nil {
		t.Logf(err.Error())
	}

	assert.Equal(t, uint32(flag), expectedFlag, "Flags should be equal to account created")
}

func TestSetLanguage(t *testing.T) {
	// Create a new instance of the Flag Manager
	fm, err := NewFlagManager(flagsPath)

	if err != nil {
		log.Fatal(err)
	}
	// Define test cases
	tests := []struct {
		name           string
		execPath       []string
		expectedResult resource.Result
	}{
		{
			name:     "Set Default Language (English)",
			execPath: []string{"set_eng"},
			expectedResult: resource.Result{
				FlagSet: []uint32{state.FLAG_LANG, 8},
				Content: "eng",
			},
		},
		{
			name:     "Set Swahili Language",
			execPath: []string{"set_swa"},
			expectedResult: resource.Result{
				FlagSet: []uint32{state.FLAG_LANG, 8},
				Content: "swa",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockState := state.NewState(16)
			// Set the ExecPath
			mockState.ExecPath = tt.execPath

			// Create the Handlers instance with the mock flag manager
			h := &Handlers{
				flagManager: fm.parser,
				st:          mockState,
			}

			// Call the method
			res, err := h.SetLanguage(context.Background(), "set_language", nil)

			if err != nil {
				t.Error(err)
			}

			// Assert that the Result FlagSet has the required flags after language switch
			assert.Equal(t, res, tt.expectedResult, "Result should match expected result")

		})
	}
}
func TestResetAllowUpdate(t *testing.T) {
	fm, err := NewFlagManager(flagsPath)

	flag_allow_update, _ := fm.parser.GetFlag("flag_allow_update")

	if err != nil {
		log.Fatal(err)
	}
	// Define test cases
	tests := []struct {
		name           string
		input          []byte
		expectedResult resource.Result
	}{
		{
			name:  "Resets allow update",
			input: []byte(""),
			expectedResult: resource.Result{
				FlagReset: []uint32{flag_allow_update},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Create the Handlers instance with the mock flag manager
			h := &Handlers{
				flagManager: fm.parser,
			}

			// Call the method
			res, err := h.ResetAllowUpdate(context.Background(), "reset_allow update", tt.input)

			if err != nil {
				t.Error(err)
			}
			// Assert that the Result FlagSet has the required flags after language switch
			assert.Equal(t, res, tt.expectedResult, "Flags should be equal to account created")

		})
	}
}

func TestResetAccountAuthorized(t *testing.T) {
	fm, err := NewFlagManager(flagsPath)

	flag_account_authorized, _ := fm.parser.GetFlag("flag_account_authorized")

	if err != nil {
		log.Fatal(err)
	}
	// Define test cases
	tests := []struct {
		name           string
		input          []byte
		expectedResult resource.Result
	}{
		{
			name:  "Resets account authorized",
			input: []byte(""),
			expectedResult: resource.Result{
				FlagReset: []uint32{flag_account_authorized},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Create the Handlers instance with the mock flag manager
			h := &Handlers{
				flagManager: fm.parser,
			}

			// Call the method
			res, err := h.ResetAccountAuthorized(context.Background(), "reset_account_authorized", tt.input)

			if err != nil {
				t.Error(err)
			}
			// Assert that the Result FlagSet has the required flags after language switch
			assert.Equal(t, res, tt.expectedResult, "Result should contain flag(s) that have been reset")

		})
	}
}

func TestIncorrectPinReset(t *testing.T) {
	fm, err := NewFlagManager(flagsPath)

	flag_incorrect_pin, _ := fm.parser.GetFlag("flag_incorrect_pin")

	if err != nil {
		log.Fatal(err)
	}
	// Define test cases
	tests := []struct {
		name           string
		input          []byte
		expectedResult resource.Result
	}{
		{
			name:  "Test incorrect pin reset",
			input: []byte(""),
			expectedResult: resource.Result{
				FlagReset: []uint32{flag_incorrect_pin},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Create the Handlers instance with the mock flag manager
			h := &Handlers{
				flagManager: fm.parser,
			}

			// Call the method
			res, err := h.ResetIncorrectPin(context.Background(), "reset_incorrect_pin", tt.input)

			if err != nil {
				t.Error(err)
			}
			// Assert that the Result FlagSet has the required flags after language switch
			assert.Equal(t, res, tt.expectedResult, "Result should contain flag(s) that have been reset")

		})
	}
}

func TestResetIncorrectYob(t *testing.T) {
	fm, err := NewFlagManager(flagsPath)

	flag_incorrect_date_format, _ := fm.parser.GetFlag("flag_incorrect_date_format")

	if err != nil {
		log.Fatal(err)
	}
	// Define test cases
	tests := []struct {
		name           string
		input          []byte
		expectedResult resource.Result
	}{
		{
			name:  "Test incorrect yob reset",
			input: []byte(""),
			expectedResult: resource.Result{
				FlagReset: []uint32{flag_incorrect_date_format},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Create the Handlers instance with the mock flag manager
			h := &Handlers{
				flagManager: fm.parser,
			}

			// Call the method
			res, err := h.ResetIncorrectYob(context.Background(), "reset_incorrect_yob", tt.input)

			if err != nil {
				t.Error(err)
			}
			// Assert that the Result FlagSet has the required flags after language switch
			assert.Equal(t, res, tt.expectedResult, "Result should contain flag(s) that have been reset")

		})
	}
}

func TestAuthorize(t *testing.T) {

	fm, err := NewFlagManager(flagsPath)

	if err != nil {
		t.Logf(err.Error())
	}

	// Create required mocks
	mockDataStore := new(mocks.MockUserDataStore)
	mockCreateAccountService := new(mocks.MockAccountService)
	//expectedResult := resource.Result{}
	mockState := state.NewState(16)
	flag_incorrect_pin, _ := fm.GetFlag("flag_incorrect_pin")
	flag_account_authorized, _ := fm.GetFlag("flag_account_authorized")
	flag_allow_update, _ := fm.GetFlag("flag_allow_update")
	//Assuming 1234 is the correct account pin
	accountPIN := "1234"

	// Define session ID and mock data
	sessionId := "session123"
	typ := utils.DATA_ACCOUNT_PIN

	h := &Handlers{
		userdataStore:  mockDataStore,
		accountService: mockCreateAccountService,
		flagManager:    fm.parser,
		st:             mockState,
	}

	tests := []struct {
		name           string
		input          []byte
		expectedResult resource.Result
	}{
		{
			name:  "Test with correct pin",
			input: []byte("1234"),
			expectedResult: resource.Result{
				FlagReset: []uint32{flag_incorrect_pin},
				FlagSet:   []uint32{flag_allow_update, flag_account_authorized},
			},
		},
		{
			name:  "Test with incorrect pin",
			input: []byte("1235"),
			expectedResult: resource.Result{
				FlagReset: []uint32{flag_account_authorized},
				FlagSet:   []uint32{flag_incorrect_pin},
			},
		},
		{
			name:           "Test with pin that is not a 4 digit",
			input:          []byte("1235aqds"),
			expectedResult: resource.Result{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Create context with session ID
			ctx := context.WithValue(context.Background(), "SessionId", sessionId)

			// Define expected interactions with the mock
			mockDataStore.On("ReadEntry", ctx, sessionId, typ).Return([]byte(accountPIN), nil)

			// Create a Handlers instance with the mock data store

			// Call the method under test
			res, err := h.Authorize(ctx, "authorize", []byte(tt.input))

			// Assert that no errors occurred
			assert.NoError(t, err)

			//Assert that the account created flag has been set to the result
			assert.Equal(t, res, tt.expectedResult, "Expected result should be equal to the actual result")

			// Assert that expectations were met
			mockDataStore.AssertExpectations(t)

		})
	}

}

func TestVerifyYob(t *testing.T) {
	fm, err := NewFlagManager(flagsPath)

	if err != nil {
		t.Logf(err.Error())
	}

	sessionId := "session123"
	// Create required mocks
	mockDataStore := new(mocks.MockUserDataStore)
	mockCreateAccountService := new(mocks.MockAccountService)
	mockState := state.NewState(16)
	flag_incorrect_date_format, _ := fm.parser.GetFlag("flag_incorrect_date_format")
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	h := &Handlers{
		userdataStore:  mockDataStore,
		accountService: mockCreateAccountService,
		flagManager:    fm.parser,
		st:             mockState,
	}

	tests := []struct {
		name           string
		input          []byte
		expectedResult resource.Result
	}{
		{
			name:  "Test with correct yob",
			input: []byte("1980"),
			expectedResult: resource.Result{
				FlagReset: []uint32{flag_incorrect_date_format},
			},
		},
		{
			name:  "Test with incorrect yob",
			input: []byte("sgahaha"),
			expectedResult: resource.Result{
				FlagSet: []uint32{flag_incorrect_date_format},
			},
		},
		{
			name:  "Test with numeric but less 4 digits",
			input: []byte("123"),
			expectedResult: resource.Result{
				FlagSet: []uint32{flag_incorrect_date_format},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Call the method under test
			res, err := h.VerifyYob(ctx, "verify_yob", []byte(tt.input))

			// Assert that no errors occurred
			assert.NoError(t, err)

			//Assert that the account created flag has been set to the result
			assert.Equal(t, res, tt.expectedResult, "Expected result should be equal to the actual result")

			// Assert that expectations were met
			mockDataStore.AssertExpectations(t)

		})
	}
}

func TestVerifyCreatePin(t *testing.T) {
	fm, err := NewFlagManager(flagsPath)

	if err != nil {
		t.Logf(err.Error())
	}

	sessionId := "session123"
	// Create required mocks
	mockDataStore := new(mocks.MockUserDataStore)
	mockCreateAccountService := new(mocks.MockAccountService)
	mockState := state.NewState(16)

	flag_valid_pin, _ := fm.parser.GetFlag("flag_valid_pin")
	flag_pin_mismatch, _ := fm.parser.GetFlag("flag_pin_mismatch")
	flag_pin_set, _ := fm.parser.GetFlag("flag_pin_set")
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)
	//Assuming this was the first set PIN to verify against
	firstSetPin := "1234"

	h := &Handlers{
		userdataStore:  mockDataStore,
		accountService: mockCreateAccountService,
		flagManager:    fm.parser,
		st:             mockState,
	}

	tests := []struct {
		name           string
		input          []byte
		expectedResult resource.Result
	}{
		{
			name:  "Test with correct PIN confirmation",
			input: []byte("1234"),
			expectedResult: resource.Result{
				FlagSet:   []uint32{flag_valid_pin, flag_pin_set},
				FlagReset: []uint32{flag_pin_mismatch},
			},
		},
		{
			name:  "Test with PIN that does not match first ",
			input: []byte("1324"),
			expectedResult: resource.Result{
				FlagSet: []uint32{flag_pin_mismatch},
			},
		},
	}

	typ := utils.DATA_TEMPORARY_PIN

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Define expected interactions with the mock
			mockDataStore.On("ReadEntry", ctx, sessionId, typ).Return([]byte(firstSetPin), nil)

			// Set up the expected behavior of the mock
			mockDataStore.On("WriteEntry", ctx, sessionId, utils.DATA_ACCOUNT_PIN, []byte(firstSetPin)).Return(nil)

			// Call the method under test
			res, err := h.VerifyCreatePin(ctx, "verify_create_pin", []byte(tt.input))

			// Assert that no errors occurred
			assert.NoError(t, err)

			//Assert that the account created flag has been set to the result
			assert.Equal(t, res, tt.expectedResult, "Expected result should be equal to the actual result")

			// Assert that expectations were met
			mockDataStore.AssertExpectations(t)

		})
	}
}

func TestCheckAccountStatus(t *testing.T) {
	fm, err := NewFlagManager(flagsPath)

	if err != nil {
		t.Logf(err.Error())
	}
	mockDataStore := new(mocks.MockUserDataStore)
	mockCreateAccountService := new(mocks.MockAccountService)

	sessionId := "session123"
	flag_account_success, _ := fm.GetFlag("flag_account_success")
	flag_account_pending, _ := fm.GetFlag("flag_account_pending")

	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	h := &Handlers{
		userdataStore:  mockDataStore,
		accountService: mockCreateAccountService,
		flagManager:    fm.parser,
	}
	tests := []struct {
		name           string
		input          []byte
		status         string
		expectedResult resource.Result
	}{
		{
			name:   "Test when account status is Success",
			input:  []byte("TrackingId1234"),
			status: "SUCCESS",
			expectedResult: resource.Result{
				FlagSet:   []uint32{flag_account_success},
				FlagReset: []uint32{flag_account_pending},
			},
		},
	}

	typ := utils.DATA_TRACKING_ID
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Define expected interactions with the mock
			mockDataStore.On("ReadEntry", ctx, sessionId, typ).Return(tt.input, nil)

			mockCreateAccountService.On("CheckAccountStatus", string(tt.input)).Return(tt.status, nil)
			mockDataStore.On("WriteEntry", ctx, sessionId, utils.DATA_ACCOUNT_STATUS, []byte(tt.status)).Return(nil)

			// Call the method under test
			res, _ := h.CheckAccountStatus(ctx, "check_status", tt.input)

			// Assert that no errors occurred
			assert.NoError(t, err)

			//Assert that the account created flag has been set to the result
			assert.Equal(t, res, tt.expectedResult, "Expected result should be equal to the actual result")

			// Assert that expectations were met
			mockDataStore.AssertExpectations(t)

		})
	}

}

func TestTransactionReset(t *testing.T) {
	fm, err := NewFlagManager(flagsPath)

	if err != nil {
		t.Logf(err.Error())
	}
	flag_invalid_recipient, _ := fm.GetFlag("flag_invalid_recipient")
	flag_invalid_recipient_with_invite, _ := fm.GetFlag("flag_invalid_recipient_with_invite")

	mockDataStore := new(mocks.MockUserDataStore)
	mockCreateAccountService := new(mocks.MockAccountService)

	sessionId := "session123"

	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	h := &Handlers{
		userdataStore:  mockDataStore,
		accountService: mockCreateAccountService,
		flagManager:    fm.parser,
	}
	tests := []struct {
		name           string
		input          []byte
		status         string
		expectedResult resource.Result
	}{
		{
			name: "Test transaction reset for amount and recipient",
			expectedResult: resource.Result{
				FlagReset: []uint32{flag_invalid_recipient, flag_invalid_recipient_with_invite},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDataStore.On("WriteEntry", ctx, sessionId, utils.DATA_AMOUNT, []byte("")).Return(nil)
			mockDataStore.On("WriteEntry", ctx, sessionId, utils.DATA_RECIPIENT, []byte("")).Return(nil)

			// Call the method under test
			res, _ := h.TransactionReset(ctx, "transaction_reset", tt.input)

			// Assert that no errors occurred
			assert.NoError(t, err)

			//Assert that the account created flag has been set to the result
			assert.Equal(t, res, tt.expectedResult, "Expected result should be equal to the actual result")

			// Assert that expectations were met
			mockDataStore.AssertExpectations(t)

		})
	}
}

func TestResetInvalidAmount(t *testing.T) {
	sessionId := "session123"

	fm, err := NewFlagManager(flagsPath)

	if err != nil {
		t.Logf(err.Error())
	}

	flag_invalid_amount, _ := fm.parser.GetFlag("flag_invalid_amount")

	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	mockDataStore := new(mocks.MockUserDataStore)
	mockCreateAccountService := new(mocks.MockAccountService)

	h := &Handlers{
		userdataStore:  mockDataStore,
		accountService: mockCreateAccountService,
		flagManager:    fm.parser,
	}

	tests := []struct {
		name           string
		input          []byte
		status         string
		expectedResult resource.Result
	}{
		{
			name: "Test amount reset",
			expectedResult: resource.Result{
				FlagReset: []uint32{flag_invalid_amount},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDataStore.On("WriteEntry", ctx, sessionId, utils.DATA_AMOUNT, []byte("")).Return(nil)

			// Call the method under test
			res, _ := h.ResetTransactionAmount(ctx, "transaction_reset_amount", tt.input)

			// Assert that no errors occurred
			assert.NoError(t, err)

			//Assert that the account created flag has been set to the result
			assert.Equal(t, res, tt.expectedResult, "Expected result should be equal to the actual result")

			// Assert that expectations were met
			mockDataStore.AssertExpectations(t)

		})
	}

}

func TestInitiateTransaction(t *testing.T) {
	sessionId := "session123"

	fm, err := NewFlagManager(flagsPath)

	if err != nil {
		t.Logf(err.Error())
	}
	account_authorized_flag, err := fm.parser.GetFlag("flag_account_authorized")

	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	mockDataStore := new(mocks.MockUserDataStore)
	mockCreateAccountService := new(mocks.MockAccountService)

	h := &Handlers{
		userdataStore:  mockDataStore,
		accountService: mockCreateAccountService,
		flagManager:    fm.parser,
	}

	tests := []struct {
		name           string
		input          []byte
		PublicKey      []byte
		Recipient      []byte
		Amount         []byte
		status         string
		expectedResult resource.Result
	}{
		{
			name:      "Test amount reset",
			PublicKey: []byte("0x1241527192"),
			Amount:    []byte("0.002CELO"),
			Recipient: []byte("0x12415ass27192"),
			expectedResult: resource.Result{
				FlagReset: []uint32{account_authorized_flag},
				Content:   "Your request has been sent. 0x12415ass27192 will receive 0.002CELO from 0x1241527192.",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Define expected interactions with the mock
			mockDataStore.On("ReadEntry", ctx, sessionId, utils.DATA_PUBLIC_KEY).Return(tt.PublicKey, nil)
			mockDataStore.On("ReadEntry", ctx, sessionId, utils.DATA_AMOUNT).Return(tt.Amount, nil)
			mockDataStore.On("ReadEntry", ctx, sessionId, utils.DATA_RECIPIENT).Return(tt.Recipient, nil)
			//mockDataStore.On("WriteEntry", ctx, sessionId, utils.DATA_AMOUNT, []byte("")).Return(nil)

			// Call the method under test
			res, _ := h.InitiateTransaction(ctx, "transaction_reset_amount", tt.input)

			// Assert that no errors occurred
			assert.NoError(t, err)

			//Assert that the account created flag has been set to the result
			assert.Equal(t, res, tt.expectedResult, "Expected result should be equal to the actual result")

			// Assert that expectations were met
			mockDataStore.AssertExpectations(t)

		})
	}

}

func TestQuit(t *testing.T) {
	fm, err := NewFlagManager(flagsPath)

	if err != nil {
		t.Logf(err.Error())
	}
	flag_account_authorized, _ := fm.parser.GetFlag("flag_account_authorized")

	mockDataStore := new(mocks.MockUserDataStore)
	mockCreateAccountService := new(mocks.MockAccountService)

	sessionId := "session123"

	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	h := &Handlers{
		userdataStore:  mockDataStore,
		accountService: mockCreateAccountService,
		flagManager:    fm.parser,
	}
	tests := []struct {
		name           string
		input          []byte
		status         string
		expectedResult resource.Result
	}{
		{
			name: "Test quit message",
			expectedResult: resource.Result{
				FlagReset: []uint32{flag_account_authorized},
				Content:   "Thank you for using Sarafu. Goodbye!",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Call the method under test
			res, _ := h.Quit(ctx, "test_quit", tt.input)

			// Assert that no errors occurred
			assert.NoError(t, err)

			//Assert that the account created flag has been set to the result
			assert.Equal(t, res, tt.expectedResult, "Expected result should be equal to the actual result")

			// Assert that expectations were met
			mockDataStore.AssertExpectations(t)

		})
	}
}
func TestIsValidPIN(t *testing.T) {
	tests := []struct {
		name     string
		pin      string
		expected bool
	}{
		{
			name:     "Valid PIN with 4 digits",
			pin:      "1234",
			expected: true,
		},
		{
			name:     "Valid PIN with leading zeros",
			pin:      "0001",
			expected: true,
		},
		{
			name:     "Invalid PIN with less than 4 digits",
			pin:      "123",
			expected: false,
		},
		{
			name:     "Invalid PIN with more than 4 digits",
			pin:      "12345",
			expected: false,
		},
		{
			name:     "Invalid PIN with letters",
			pin:      "abcd",
			expected: false,
		},
		{
			name:     "Invalid PIN with special characters",
			pin:      "12@#",
			expected: false,
		},
		{
			name:     "Empty PIN",
			pin:      "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := isValidPIN(tt.pin)
			if actual != tt.expected {
				t.Errorf("isValidPIN(%q) = %v; expected %v", tt.pin, actual, tt.expected)
			}
		})
	}
}

func TestQuitWithBalance(t *testing.T) {
	fm, err := NewFlagManager(flagsPath)

	if err != nil {
		t.Logf(err.Error())
	}
	flag_account_authorized, _ := fm.parser.GetFlag("flag_account_authorized")

	mockDataStore := new(mocks.MockUserDataStore)
	mockCreateAccountService := new(mocks.MockAccountService)

	sessionId := "session123"

	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	h := &Handlers{
		userdataStore:  mockDataStore,
		accountService: mockCreateAccountService,
		flagManager:    fm.parser,
	}
	tests := []struct {
		name           string
		input          []byte
		publicKey      []byte
		balance        string
		expectedResult resource.Result
	}{
		{
			name:      "Test quit with balance",
			balance:   "0.02CELO",
			publicKey: []byte("0xrqeqrequuq"),
			expectedResult: resource.Result{
				FlagReset: []uint32{flag_account_authorized},
				Content:   fmt.Sprintf("Your account balance is %s", "0.02CELO"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockDataStore.On("ReadEntry", ctx, sessionId, utils.DATA_PUBLIC_KEY).Return(tt.publicKey, nil)
			mockCreateAccountService.On("CheckBalance", string(tt.publicKey)).Return(tt.balance, nil)

			// Call the method under test
			res, _ := h.QuitWithBalance(ctx, "test_quit_with_balance", tt.input)

			// Assert that no errors occurred
			assert.NoError(t, err)

			//Assert that the account created flag has been set to the result
			assert.Equal(t, res, tt.expectedResult, "Expected result should be equal to the actual result")

			// Assert that expectations were met
			mockDataStore.AssertExpectations(t)

		})
	}
}

func TestValidateAmount(t *testing.T) {
	fm, err := NewFlagManager(flagsPath)

	if err != nil {
		t.Logf(err.Error())
	}
	flag_invalid_amount, _ := fm.parser.GetFlag("flag_invalid_amount")
	mockDataStore := new(mocks.MockUserDataStore)
	mockCreateAccountService := new(mocks.MockAccountService)

	sessionId := "session123"

	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	h := &Handlers{
		userdataStore:  mockDataStore,
		accountService: mockCreateAccountService,
		flagManager:    fm.parser,
	}
	tests := []struct {
		name           string
		input          []byte
		publicKey      []byte
		activeSym      []byte
		activeBal      []byte
		balance        string
		expectedResult resource.Result
	}{
		{
			name:      "Test with valid amount and active symbol",
			input:     []byte("0.001"),
			publicKey: []byte("0xrqeqrequuq"),
			activeSym: []byte("CELO"),
			activeBal: []byte("0.003"),
			expectedResult: resource.Result{
				Content: "0.001",
			},
		},
		{
			name:      "Test with amount larger than active balance",
			input:     []byte("0.02"),
			publicKey: []byte("0xrqeqrequuq"),
			activeSym: []byte("CELO"),
			activeBal: []byte("0.003"),
			expectedResult: resource.Result{
				FlagSet: []uint32{flag_invalid_amount},
				Content: "0.02",
			},
		},
		{
			name:      "Test with invalid amount format",
			input:     []byte("0.02ms"),
			publicKey: []byte("0xrqeqrequuq"),
			balance:   "0.003 CELO",
			expectedResult: resource.Result{
				FlagSet: []uint32{flag_invalid_amount},
				Content: "0.02ms",
			},
		},
		{
			name:      "Test fallback to current balance without active symbol",
			input:     []byte("0.001"),
			publicKey: []byte("0xrqeqrequuq"),
			balance:   "0.003 CELO",
			expectedResult: resource.Result{
				Content: "0.001",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock behavior for public key retrieval
			mockDataStore.On("ReadEntry", ctx, sessionId, utils.DATA_PUBLIC_KEY).Return(tt.publicKey, nil)

			// Mock behavior for active symbol and balance retrieval (if present)
			if len(tt.activeSym) > 0 {
				mockDataStore.On("ReadEntry", ctx, sessionId, utils.DATA_ACTIVE_SYM).Return(tt.activeSym, nil)
				mockDataStore.On("ReadEntry", ctx, sessionId, utils.DATA_ACTIVE_BAL).Return(tt.activeBal, nil)
			} else {
				mockDataStore.On("ReadEntry", ctx, sessionId, utils.DATA_ACTIVE_SYM).Return(nil, fmt.Errorf("not found"))
				mockCreateAccountService.On("CheckBalance", string(tt.publicKey)).Return(tt.balance, nil)
			}

			// Mock behavior for storing the amount (if valid)
			mockDataStore.On("WriteEntry", ctx, sessionId, utils.DATA_AMOUNT, tt.input).Return(nil).Maybe()

			// Call the method under test
			res, _ := h.ValidateAmount(ctx, "test_validate_amount", tt.input)

			// Assert no errors occurred
			assert.NoError(t, err)

			// Assert the result matches the expected result
			assert.Equal(t, tt.expectedResult, res, "Expected result should match actual result")

			// Assert all expectations were met
			mockDataStore.AssertExpectations(t)
		})
	}
}

func TestValidateRecipient(t *testing.T) {
	fm, err := NewFlagManager(flagsPath)

	flag_invalid_recipient, _ := fm.parser.GetFlag("flag_invalid_recipient")
	mockDataStore := new(mocks.MockUserDataStore)

	sessionId := "session123"

	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	if err != nil {
		log.Fatal(err)
	}
	// Define test cases
	tests := []struct {
		name           string
		input          []byte
		expectedResult resource.Result
	}{
		{
			name:  "Test with invalid recepient",
			input: []byte("000"),
			expectedResult: resource.Result{
				FlagSet: []uint32{flag_invalid_recipient},
				Content: "000",
			},
		},
		{
			name:           "Test with valid recepient",
			input:          []byte("0705X2"),
			expectedResult: resource.Result{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockDataStore.On("WriteEntry", ctx, sessionId, utils.DATA_RECIPIENT, tt.input).Return(nil)

			// Create the Handlers instance with the mock flag manager
			h := &Handlers{
				flagManager:   fm.parser,
				userdataStore: mockDataStore,
			}

			// Call the method
			res, err := h.ValidateRecipient(ctx, "validate_recepient", tt.input)

			if err != nil {
				t.Error(err)
			}
			// Assert that the Result FlagSet has the required flags after language switch
			assert.Equal(t, res, tt.expectedResult, "Result should contain flag(s) that have been reset")

		})
	}
}

func TestCheckBalance(t *testing.T) {
	tests := []struct {
		name           string
		sessionId      string
		publicKey      string
		activeSym      string
		activeBal      string
		expectedResult resource.Result
		expectError    bool
	}{
		{
			name:           "User with active sym",
			sessionId:      "session456",
			publicKey:      "0X98765432109",
			activeSym:      "ETH",
			activeBal:      "1.5",
			expectedResult: resource.Result{Content: "1.5 ETH"},
			expectError:    false,
		},
		{
			name:           "User without active sym",
			sessionId:      "session123",
			publicKey:      "0X13242618721",
			activeSym:      "",
			activeBal:      "",
			expectedResult: resource.Result{Content: "0.003 CELO"},
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDataStore := new(mocks.MockUserDataStore)
			mockAccountService := new(mocks.MockAccountService)
			ctx := context.WithValue(context.Background(), "SessionId", tt.sessionId)

			h := &Handlers{
				userdataStore:  mockDataStore,
				accountService: mockAccountService,
			}

			// Mock calls for public key
			mockDataStore.On("ReadEntry", ctx, tt.sessionId, utils.DATA_PUBLIC_KEY).Return([]byte(tt.publicKey), nil)

			if tt.activeSym == "" {
				// Mock for user without active sym
				mockDataStore.On("ReadEntry", ctx, tt.sessionId, utils.DATA_ACTIVE_SYM).Return([]byte{}, db.ErrNotFound{})
				mockAccountService.On("CheckBalance", tt.publicKey).Return("0.003 CELO", nil)
			} else {
				// Mock for user with active sym
				mockDataStore.On("ReadEntry", ctx, tt.sessionId, utils.DATA_ACTIVE_SYM).Return([]byte(tt.activeSym), nil)
				mockDataStore.On("ReadEntry", ctx, tt.sessionId, utils.DATA_ACTIVE_BAL).Return([]byte(tt.activeBal), nil)
			}

			res, err := h.CheckBalance(ctx, "check_balance", []byte("123456"))

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, res, "Result should match expected output")
			}

			mockDataStore.AssertExpectations(t)
			mockAccountService.AssertExpectations(t)
		})
	}
}

func TestGetProfile(t *testing.T) {

	sessionId := "session123"

	mockDataStore := new(mocks.MockUserDataStore)
	mockCreateAccountService := new(mocks.MockAccountService)

	h := &Handlers{
		userdataStore:  mockDataStore,
		accountService: mockCreateAccountService,
	}
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	tests := []struct {
		name        string
		keys        []utils.DataTyp
		profileInfo []string
		result      resource.Result
	}{
		{
			name:        "Test with full profile information",
			keys:        []utils.DataTyp{utils.DATA_FAMILY_NAME, utils.DATA_FIRST_NAME, utils.DATA_GENDER, utils.DATA_OFFERINGS, utils.DATA_LOCATION, utils.DATA_YOB},
			profileInfo: []string{"Doee", "John", "Male", "Bananas", "Kilifi", "1976"},
			result: resource.Result{
				Content: fmt.Sprintf(
					"Name: %s\nGender: %s\nAge: %s\nLocation: %s\nYou provide: %s\n",
					"John Doee", "Male", "48", "Kilifi", "Bananas",
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for index, key := range tt.keys {
				mockDataStore.On("ReadEntry", ctx, sessionId, key).Return([]byte(tt.profileInfo[index]), nil)
			}
			res, _ := h.GetProfileInfo(ctx, "get_profile_info", []byte(""))

			// Assert that expectations were met
			mockDataStore.AssertExpectations(t)

			//Assert that the result set to content is what was expected
			assert.Equal(t, res, tt.result, "Result should contain profile information served back to user")

		})
	}
}

func TestVerifyNewPin(t *testing.T) {
	sessionId := "session123"

	fm, _ := NewFlagManager(flagsPath)

	flag_valid_pin, _ := fm.parser.GetFlag("flag_valid_pin")
	mockDataStore := new(mocks.MockUserDataStore)
	mockCreateAccountService := new(mocks.MockAccountService)
	h := &Handlers{
		userdataStore:  mockDataStore,
		flagManager:    fm.parser,
		accountService: mockCreateAccountService,
	}
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	tests := []struct {
		name           string
		input          []byte
		expectedResult resource.Result
	}{
		{
			name:  "Test with valid pin",
			input: []byte("1234"),
			expectedResult: resource.Result{
				FlagSet: []uint32{flag_valid_pin},
			},
		},
		{
			name:  "Test with invalid pin",
			input: []byte("123"),
			expectedResult: resource.Result{
				FlagReset: []uint32{flag_valid_pin},
			},
		},
		{
			name:  "Test with invalid pin",
			input: []byte("12345"),
			expectedResult: resource.Result{
				FlagReset: []uint32{flag_valid_pin},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			//Call the function under test
			res, _ := h.VerifyNewPin(ctx, "verify_new_pin", tt.input)

			// Assert that expectations were met
			mockDataStore.AssertExpectations(t)

			//Assert that the result set to content is what was expected
			assert.Equal(t, res, tt.expectedResult, "Result should contain flags set according to user input")

		})
	}

}

func TestConfirmPin(t *testing.T) {
	sessionId := "session123"

	fm, _ := NewFlagManager(flagsPath)
	flag_pin_mismatch, _ := fm.parser.GetFlag("flag_pin_mismatch")
	mockDataStore := new(mocks.MockUserDataStore)
	mockCreateAccountService := new(mocks.MockAccountService)
	h := &Handlers{
		userdataStore:  mockDataStore,
		flagManager:    fm.parser,
		accountService: mockCreateAccountService,
	}
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	tests := []struct {
		name           string
		input          []byte
		temporarypin   []byte
		expectedResult resource.Result
	}{
		{
			name:         "Test with correct pin confirmation",
			input:        []byte("1234"),
			temporarypin: []byte("1234"),
			expectedResult: resource.Result{
				FlagReset: []uint32{flag_pin_mismatch},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up the expected behavior of the mock
			mockDataStore.On("WriteEntry", ctx, sessionId, utils.DATA_ACCOUNT_PIN, []byte(tt.temporarypin)).Return(nil)

			mockDataStore.On("ReadEntry", ctx, sessionId, utils.DATA_TEMPORARY_PIN).Return(tt.temporarypin, nil)

			//Call the function under test
			res, _ := h.ConfirmPinChange(ctx, "confirm_pin_change", tt.temporarypin)

			// Assert that expectations were met
			mockDataStore.AssertExpectations(t)

			//Assert that the result set to content is what was expected
			assert.Equal(t, res, tt.expectedResult, "Result should contain flags set according to user input")

		})
	}
}

func TestSetVoucher(t *testing.T) {
	mockDataStore := new(mocks.MockUserDataStore)

	sessionId := "session123"
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	temporarySym := []byte("tempSym")
	temporaryBal := []byte("tempBal")

	// Set expectations for the mock data store
	mockDataStore.On("ReadEntry", ctx, sessionId, utils.DATA_TEMPORARY_SYM).Return(temporarySym, nil)
	mockDataStore.On("ReadEntry", ctx, sessionId, utils.DATA_TEMPORARY_BAL).Return(temporaryBal, nil)
	mockDataStore.On("WriteEntry", ctx, sessionId, utils.DATA_ACTIVE_SYM, temporarySym).Return(nil)
	mockDataStore.On("WriteEntry", ctx, sessionId, utils.DATA_ACTIVE_BAL, temporaryBal).Return(nil)
	mockDataStore.On("WriteEntry", ctx, sessionId, utils.DATA_TEMPORARY_SYM, []byte("")).Return(nil)
	mockDataStore.On("WriteEntry", ctx, sessionId, utils.DATA_TEMPORARY_BAL, []byte("")).Return(nil)

	h := &Handlers{
		userdataStore: mockDataStore,
	}

	// Call the method under test
	res, err := h.SetVoucher(ctx, "someSym", []byte{})

	// Assert that no errors occurred
	assert.NoError(t, err)

	// Assert that the result content is correct
	assert.Equal(t, string(temporarySym), res.Content)

	// Assert that expectations were met
	mockDataStore.AssertExpectations(t)
}
