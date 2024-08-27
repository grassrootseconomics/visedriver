package utils

import (
	"encoding/json"
	"os"
)

type AccountFileHandler struct {
	FilePath string
}

func NewAccountFileHandler(path string) *AccountFileHandler {
	return &AccountFileHandler{FilePath: path}
}

func (afh *AccountFileHandler) ReadAccountData() (map[string]string, error) {
	jsonData, err := os.ReadFile(afh.FilePath)
	if err != nil {
		return nil, err
	}

	var accountData map[string]string
	err = json.Unmarshal(jsonData, &accountData)
	if err != nil {
		return nil, err
	}

	return accountData, nil
}

func (afh *AccountFileHandler) WriteAccountData(accountData map[string]string) error {
	jsonData, err := json.Marshal(accountData)
	if err != nil {
		return err
	}

	return os.WriteFile(afh.FilePath, jsonData, 0644)
}

func (afh *AccountFileHandler) EnsureFileExists() error {
	f, err := os.OpenFile(afh.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	return f.Close()
}
