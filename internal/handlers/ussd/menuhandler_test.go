package ussd

import (
	"context"
	"testing"

	"git.defalsify.org/vise.git/resource"
	"git.grassecon.net/urdt/ussd/internal/handlers/ussd/mocks"
	"git.grassecon.net/urdt/ussd/internal/utils"
	"github.com/alecthomas/assert/v2"
)

// func TestCreateAccount(t *testing.T) {
// 	// Create a new instance of MockMyDataStore
// 	mockDataStore := new(mocks.MockMyDataStore)
// 	mockCreateAccountService := new(mocks.MockAccountService)

// 	// Define session ID and mock data
// 	sessionId := "session123"
// 	typ := utils.DATA_ACCOUNT_CREATED
// 	fakeError := db.ErrNotFound{}
// 	// Create context with session ID
// 	ctx := context.WithValue(context.Background(), "SessionId", sessionId)

// 	// Define expected interactions with the mock
// 	mockDataStore.On("ReadEntry", ctx, sessionId, typ).Return([]byte("123"), fakeError)
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
// 	data := map[utils.DataTyp]string{
// 		utils.DATA_TRACKING_ID:  expectedAccountResp.Result.TrackingId,
// 		utils.DATA_PUBLIC_KEY:   expectedAccountResp.Result.PublicKey,
// 		utils.DATA_CUSTODIAL_ID: expectedAccountResp.Result.CustodialId.String(),
// 	}

// 	for key, value := range data {
// 		//err := utils.WriteEntry(ctx, h.userdataStore, sessionId, key, []byte(value))
// 		mockDataStore.On("WriteEntry", ctx, sessionId, key, []byte(value)).Return(nil)
// 	}
// 	//mockDataStore.On("WriteEntry", mock.Anything, sessionId, mock.Anything, mock.Anything).Return(nil)

// 	// Create a Handlers instance with the mock data store
// 	h := &Handlers{
// 		userdataStore:  mockDataStore,
// 		accountService: mockCreateAccountService,
// 	}

// 	// Call the method you want to test
// 	_, err := h.CreateAccount(ctx, "some-symbol", []byte("some-input"))

// 	// Assert that no errors occurred
// 	assert.NoError(t, err)

// 	// Assert that expectations were met
// 	mockDataStore.AssertExpectations(t)
// }

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
	// Create a new instance of MockMyDataStore
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
