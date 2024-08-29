package mocks

import (
	"git.grassecon.net/urdt/ussd/internal/models"
	"github.com/stretchr/testify/mock"
)

type MockAccountFileHandler struct {
	mock.Mock
}

func (m *MockAccountFileHandler) EnsureFileExists() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockAccountFileHandler) ReadAccountData() (map[string]string, error) {
	args := m.Called()
	return args.Get(0).(map[string]string), args.Error(1)
}

func (m *MockAccountFileHandler) WriteAccountData(data map[string]string) error {
	args := m.Called(data)
	return args.Error(0)
}

type MockAccountService struct {
	mock.Mock
}

func (m *MockAccountService) CreateAccount() (*models.AccountResponse, error) {
	args := m.Called()
	return args.Get(0).(*models.AccountResponse), args.Error(1)
}

func (m *MockAccountService) CheckAccountStatus(TrackingId string) (string, error) {
	args := m.Called()
	return args.Get(0).(string), args.Error(1)
}

func (m *MockAccountService) CheckBalance(PublicKey string) (string, error) {
	args := m.Called()
	return args.Get(0).(string), args.Error(1)
}
