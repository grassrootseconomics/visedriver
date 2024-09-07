package mocks

import (
	"context"

	"git.defalsify.org/vise.git/lang"
	"github.com/stretchr/testify/mock"
)

type MockDb struct {
	mock.Mock
}

func (m *MockDb) SetPrefix(prefix uint8) {
	m.Called(prefix)
}

func (m *MockDb) Prefix() uint8 {
	args := m.Called()
	return args.Get(0).(uint8)
}

func (m *MockDb) Safe() bool {
	args := m.Called()
	return args.Get(0).(bool)
}

func (m *MockDb) SetLanguage(language *lang.Language) {
	m.Called(language)
}

func (m *MockDb) SetLock(uint8, bool) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDb) Connect(ctx context.Context, connectionStr string) error {
	args := m.Called(ctx, connectionStr)
	return args.Error(0)
}

func (m *MockDb) SetSession(sessionId string) {
	m.Called(sessionId)
}

func (m *MockDb) Put(ctx context.Context, key, value []byte) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

func (m *MockDb) Get(ctx context.Context, key []byte) ([]byte, error) {
	args := m.Called(ctx, key)
	return nil, args.Error(0)
}

func (m *MockDb) Close() error {
	args := m.Called(nil)
	return args.Error(0)
}
