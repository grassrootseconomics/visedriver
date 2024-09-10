package ussd

import (
	"context"
	"encoding/json"
	"log"
	"path"
	"testing"

	"git.defalsify.org/vise.git/db"
	"git.defalsify.org/vise.git/resource"
	"git.defalsify.org/vise.git/state"
	"git.grassecon.net/urdt/ussd/internal/handlers/ussd/mocks"
	"git.grassecon.net/urdt/ussd/internal/models"
	"git.grassecon.net/urdt/ussd/internal/utils"
	"github.com/alecthomas/assert/v2"
	testdataloader "github.com/peteole/testdata-loader"
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

	// Define the session ID and context
	sessionId := "session123"
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

	// Define test cases
	tests := []struct {
		name           string
		input          []byte
		expectedGender string
		expectCall     bool
	}{
		{
			name:           "Valid Male Input",
			input:          []byte("1"),
			expectedGender: "Male",
			expectCall:     true,
		},
		{
			name:           "Valid Female Input",
			input:          []byte("2"),
			expectedGender: "Female",
			expectCall:     true,
		},
		{
			name:           "Valid Unspecified Input",
			input:          []byte("3"),
			expectedGender: "Unspecified",
			expectCall:     true,
		},
		{
			name:           "Empty Input",
			input:          []byte(""),
			expectedGender: "",
			expectCall:     false,
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

			// Create the Handlers instance with the mock store
			h := &Handlers{
				userdataStore: mockStore,
			}

			// Call the method
			_, err := h.SaveGender(ctx, "someSym", tt.input)

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

func TestMaxAmount(t *testing.T) {
	mockStore := new(mocks.MockUserDataStore)
	mockCreateAccountService := new(mocks.MockAccountService)

	// Define test data
	sessionId := "session123"
	ctx := context.WithValue(context.Background(), "SessionId", sessionId)
	publicKey := "0xcasgatweksalw1018221"
	expectedBalance := "0.003CELO"

	// Set up the expected behavior of the mock
	mockStore.On("ReadEntry", ctx, sessionId, utils.DATA_PUBLIC_KEY).Return([]byte(publicKey), nil)
	mockCreateAccountService.On("CheckBalance", publicKey).Return(expectedBalance, nil)

	// Create the Handlers instance with the mock store
	h := &Handlers{
		userdataStore:  mockStore,
		accountService: mockCreateAccountService,
	}

	// Call the method
	res, _ := h.MaxAmount(ctx, "max_amount", []byte("check_balance"))

	//Assert that the balance that was set as the result content is what was returned by  Check Balance
	assert.Equal(t, expectedBalance, res.Content)

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
		name                string
		input               []byte
		expectedFlags       []uint32
		expectedResult      resource.Result
		flagManagerResponse uint32
		flagManagerError    error
	}{
		{
			name:          "English language",
			input:         []byte("0"),
			expectedFlags: []uint32{state.FLAG_LANG, 123},
			expectedResult: resource.Result{
				FlagSet: []uint32{state.FLAG_LANG, 8},
				Content: "eng",
			},
			flagManagerResponse: 123,
			flagManagerError:    nil,
		},
		{
			name:          "Swahili language",
			input:         []byte("1"),
			expectedFlags: []uint32{state.FLAG_LANG, 123},
			expectedResult: resource.Result{
				FlagSet: []uint32{state.FLAG_LANG, 8},
				Content: "swa",
			},
			flagManagerResponse: 123,
			flagManagerError:    nil,
		},
		{
			name:          "Unhandled Input",
			input:         []byte("3"),
			expectedFlags: []uint32{123},
			expectedResult: resource.Result{
				FlagSet: []uint32{8},
			},
			flagManagerResponse: 123,
			flagManagerError:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Create the Handlers instance with the mock flag manager
			h := &Handlers{
				flagManager: fm.parser,
			}

			// Call the method
			res, err := h.SetLanguage(context.Background(), "set_language", tt.input)

			if err != nil {
				t.Error(err)
			}

			// Assert that the Result FlagSet has the required flags after language switch
			assert.Equal(t, res, tt.expectedResult, "Flags should be equal to account created")

		})
	}
}

func TestSetResetSingleEdit(t *testing.T) {
	fm, err := NewFlagManager(flagsPath)

	flag_allow_update, _ := fm.parser.GetFlag("flag_allow_update")
	flag_single_edit, _ := fm.parser.GetFlag("flag_single_edit")

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
			name:  "Set single Edit",
			input: []byte("2"),
			expectedResult: resource.Result{
				FlagSet:   []uint32{flag_single_edit},
				FlagReset: []uint32{flag_allow_update},
			},
		},
		{
			name:  "Set single Edit",
			input: []byte("3"),
			expectedResult: resource.Result{
				FlagSet:   []uint32{flag_single_edit},
				FlagReset: []uint32{flag_allow_update},
			},
		},
		{
			name:  "Set single edit",
			input: []byte("4"),
			expectedResult: resource.Result{
				FlagReset: []uint32{flag_allow_update},
				FlagSet:   []uint32{flag_single_edit},
			},
		},
		{
			name:  "No single edit set",
			input: []byte("1"),
			expectedResult: resource.Result{
				FlagReset: []uint32{flag_single_edit},
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
			res, err := h.SetResetSingleEdit(context.Background(), "set_reset_single_edit", tt.input)

			if err != nil {
				t.Error(err)
			}
			// Assert that the Result FlagSet has the required flags after language switch
			assert.Equal(t, res, tt.expectedResult, "Flags should match reset edit")

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

func TestIncorrectYob(t *testing.T) {
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

func TestVerifyPin(t *testing.T) {
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
		// {
		// 	name:  "Test with numeric but less 4 digits",
		// 	input: []byte("123"),
		// 	expectedResult: resource.Result{
		// 		FlagSet: []uint32{flag_incorrect_date_format},
		// 	},
		// },
	}

	typ := utils.DATA_ACCOUNT_PIN

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Define expected interactions with the mock
			mockDataStore.On("ReadEntry", ctx, sessionId, typ).Return([]byte(firstSetPin), nil)

			// Call the method under test
			res, err := h.VerifyPin(ctx, "verify_pin", []byte(tt.input))

			// Assert that no errors occurred
			assert.NoError(t, err)

			//Assert that the account created flag has been set to the result
			assert.Equal(t, res, tt.expectedResult, "Expected result should be equal to the actual result")

			// Assert that expectations were met
			mockDataStore.AssertExpectations(t)

		})
	}
}
