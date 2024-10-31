package mocks

import (
	"context"

	"git.defalsify.org/vise.git/db"
	"git.grassecon.net/urdt/ussd/internal/utils"
	"github.com/stretchr/testify/mock"
)

type MockUserDataStore struct {
	db.Db
	mock.Mock
}

func (m *MockUserDataStore) ReadEntry(ctx context.Context, sessionId string, typ utils.DataTyp) ([]byte, error) {
	args := m.Called(ctx, sessionId, typ)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockUserDataStore) WriteEntry(ctx context.Context, sessionId string, typ utils.DataTyp, value []byte) error {
	args := m.Called(ctx, sessionId, typ, value)
	return args.Error(0)
}
