package mocks

import (
	"git.grassecon.net/urdt/ussd/internal/models"
	"github.com/grassrootseconomics/eth-custodial/pkg/api"
	"github.com/stretchr/testify/mock"
)

// MockAccountService implements AccountServiceInterface for testing
type MockAccountService struct {
	mock.Mock
}

func (m *MockAccountService) CreateAccount() (*api.OKResponse, *api.ErrResponse) {
	args := m.Called()
	okResponse, ok := args.Get(0).(*api.OKResponse)
	errResponse, err := args.Get(1).(*api.ErrResponse)

	if ok {
		return okResponse, nil
	}

	if err {
		return nil, errResponse
	}
	return nil, nil
}

func (m *MockAccountService) CheckBalance(publicKey string) (*models.BalanceResponse, error) {
	args := m.Called(publicKey)
	return args.Get(0).(*models.BalanceResponse), args.Error(1)
}

func (m *MockAccountService) CheckAccountStatus(trackingId string) (*models.TrackStatusResponse, error) {
	args := m.Called(trackingId)
	return args.Get(0).(*models.TrackStatusResponse), args.Error(1)
}

func (m *MockAccountService) TrackAccountStatus(publicKey string) (*api.OKResponse, *api.ErrResponse) {
	args := m.Called(publicKey)
	okResponse, ok := args.Get(0).(*api.OKResponse)
	errResponse, err := args.Get(1).(*api.ErrResponse)
	if ok {
		return okResponse, nil
	}
	if err {
		return nil, errResponse
	}
	return nil, nil
}
