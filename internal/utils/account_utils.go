package utils

import (
	"context"
	"encoding/json"

	"git.defalsify.org/vise.git/db"
)

type AccountFileHandler struct {
	store db.Db
}

func NewAccountFileHandler(store db.Db) *AccountFileHandler {
	return &AccountFileHandler{
		store: store,
	}
}

func (afh *AccountFileHandler) ReadAccountData(ctx context.Context, sessionId string) (map[string]string, error) {
	var accountData map[string]string
	jsonData, err := ReadEntry(ctx, afh.store, sessionId, DATA_ACCOUNT)
	if err != nil {
		return nil,err
	}
	err = json.Unmarshal(jsonData, &accountData)
	if err != nil {
		return nil, err
	}
	return accountData, nil
}

func (afh *AccountFileHandler) WriteAccountData(ctx context.Context, sessionId string, accountData map[string]string) error {
	_, err := json.Marshal(accountData)
	if err != nil {
		return err
	}

	return nil
}

func (afh *AccountFileHandler) EnsureFileExists() error {
	return nil
}
