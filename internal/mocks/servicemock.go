package mocks

import (
	"git.grassecon.net/urdt/ussd/internal/models"
	"github.com/stretchr/testify/mock"
)

// MockAccountService implements AccountServiceInterface for testing
type MockAccountService struct {
	mock.Mock
}

func (m *MockAccountService) CreateAccount() (*models.AccountResponse, error) {
	args := m.Called()
	return args.Get(0).(*models.AccountResponse), args.Error(1)
}

func (m *MockAccountService) CheckBalance(publicKey string) (string, error) {
	args := m.Called(publicKey)
	return args.String(0), args.Error(1)
}

func (m *MockAccountService) CheckAccountStatus(trackingId string) (string, error) {
	args := m.Called(trackingId)
	return args.String(0), args.Error(1)
}

func (m *MockAccountService) FetchVouchers(publicKey string) (*models.VoucherHoldingResponse, error) {
	args := m.Called()
	return args.Get(0).(*models.VoucherHoldingResponse), args.Error(1)
}
