package utils

import (
	"context"
	"encoding/json"

	"git.defalsify.org/vise.git/db"
)

type AccountFileHandler struct {
	//FilePath string
	store db.Db
}

// func NewAccountFileHandler(path string) *AccountFileHandler {
// 	return &AccountFileHandler{FilePath: path}
// }

func NewAccountFileHandler(store db.Db) *AccountFileHandler {
	return &AccountFileHandler{
		store: store,
	}
}

// func (afh *AccountFileHandler) ReadAccountData() (map[string]string, error) {
// 	jsonData, err := os.ReadFile(afh.FilePath)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var accountData map[string]string
// 	err = json.Unmarshal(jsonData, &accountData)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return accountData, nil
// }

func (afh *AccountFileHandler) ReadAccountData(ctx context.Context, sessionId string) (map[string]string, error) {
	var accountData map[string]string
	jsonData, err := ReadEntry(ctx, afh.store, sessionId, DATA_ACCOUNT)
	err = json.Unmarshal(jsonData, &accountData)
	if err != nil {
		return nil, err
	}
	return accountData, nil
}

// func (afh *AccountFileHandler) WriteAccountData(accountData map[string]string) error {
// 	jsonData, err := json.Marshal(accountData)
// 	if err != nil {
// 		return err
// 	}

// 	return os.WriteFile(afh.FilePath, jsonData, 0644)
// }

func (afh *AccountFileHandler) WriteAccountData(ctx context.Context, sessionId string, accountData map[string]string) error {
	_, err := json.Marshal(accountData)
	if err != nil {
		return err
	}

	return nil
}

//	func (afh *AccountFileHandler) EnsureFileExists() error {
//		f, err := os.OpenFile(afh.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
//		if err != nil {
//			return err
//		}
//		return f.Close()
//	}
func (afh *AccountFileHandler) EnsureFileExists() error {
	return nil
}
