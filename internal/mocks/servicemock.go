package mocks

import (
	"context"

	"git.grassecon.net/urdt/ussd/internal/models"
	"github.com/stretchr/testify/mock"
)

// MockAccountService implements AccountServiceInterface for testing
type MockAccountService struct {
	mock.Mock
}

func (m *MockAccountService) CreateAccount(ctx context.Context) (*models.AccountResponse, error) {
	args := m.Called()
	return args.Get(0).(*models.AccountResponse), args.Error(1)
}

func (m *MockAccountService) CheckBalance(publicKey string,ctx context.Context) (*models.BalanceResponse, error) {
	args := m.Called(publicKey)
	return args.Get(0).(*models.BalanceResponse), args.Error(1)
}

func (m *MockAccountService) CheckAccountStatus(trackingId string,ctx context.Context) (*models.TrackStatusResponse, error) {
	args := m.Called(trackingId)
	return args.Get(0).(*models.TrackStatusResponse), args.Error(1)
}