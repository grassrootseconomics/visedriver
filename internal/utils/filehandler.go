package utils

import (
	"encoding/json"
	"os"
)


type AccountFileHandlerInterface interface {
    EnsureFileExists() error
    ReadAccountData() (map[string]string, error)
    WriteAccountData(data map[string]string) error
}



type AccountFileHandler2 struct {
    FilePath string
}

// Implement the methods required by AccountFileHandlerInterface.
func (afh *AccountFileHandler2) EnsureFileExists() error {
	f, err := os.OpenFile(afh.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	return f.Close()
}
func (afh *AccountFileHandler2) ReadAccountData() (map[string]string, error) { 
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
func (afh *AccountFileHandler2) WriteAccountData(data map[string]string) error { 
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return os.WriteFile(afh.FilePath, jsonData, 0644)
}