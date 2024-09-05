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
	err = json.Unmarshal(jsonData, &accountData)
	if err != nil {
		return nil, err
	}

	return accountData, nil
}

func (afh *AccountFileHandler) WriteAccountData(ctx context.Context, sessionId string, accountData map[string]string) error {
	b, err := json.Marshal(accountData)
	if err != nil {
		return err
	}
	return WriteEntry(ctx, afh.store, sessionId, DATA_ACCOUNT, b)
}
