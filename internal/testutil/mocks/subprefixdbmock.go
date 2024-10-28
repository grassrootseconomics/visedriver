package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockSubPrefixDb struct {
	mock.Mock
}

func (m *MockSubPrefixDb) Get(ctx context.Context, key []byte) ([]byte, error) {
	args := m.Called(ctx, key)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockSubPrefixDb) Put(ctx context.Context, key, val []byte) error {
	args := m.Called(ctx, key, val)
	return args.Error(0)
}
