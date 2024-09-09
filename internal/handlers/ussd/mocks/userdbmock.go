package mocks

import (
	"context"

	"git.defalsify.org/vise.git/lang"
	"git.grassecon.net/urdt/ussd/internal/utils"
	"github.com/stretchr/testify/mock"
)

type MockUserDataStore struct {
	mock.Mock
}

func (m *MockUserDataStore) SetPrefix(prefix uint8) {
	m.Called(prefix)
}

func (m *MockUserDataStore) SetSession(sessionId string) {
	m.Called(sessionId)
}

func (m *MockUserDataStore) Get(ctx context.Context, key []byte) ([]byte, error) {
	args := m.Called(ctx, key)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockUserDataStore) ReadEntry(ctx context.Context, sessionId string, typ utils.DataTyp) ([]byte, error) {
	args := m.Called(ctx, sessionId, typ)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockUserDataStore) WriteEntry(ctx context.Context, sessionId string, typ utils.DataTyp, value []byte) error {
	args := m.Called(ctx, sessionId, typ, value)
	return args.Error(0)
}

func (m *MockUserDataStore) Prefix() uint8 {
	args := m.Called()
	return args.Get(0).(uint8)
}

func (m *MockUserDataStore) Safe() bool {
	args := m.Called()
	return args.Get(0).(bool)
}

func (m *MockUserDataStore) SetLanguage(language *lang.Language) {
	m.Called(language)
}

func (m *MockUserDataStore) SetLock(uint8, bool) error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockUserDataStore) Connect(ctx context.Context, connectionStr string) error {
	args := m.Called(ctx, connectionStr)
	return args.Error(0)
}

func (m *MockUserDataStore) Put(ctx context.Context, key, value []byte) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}
func (m *MockUserDataStore) Close() error {
	args := m.Called(nil)
	return args.Error(0)
}
