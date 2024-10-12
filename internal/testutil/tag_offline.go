//go:build !online
// +build !online

package testutil

import (
	"git.grassecon.net/urdt/ussd/internal/handlers/server"
)

var (
	AccountService server.AccountServiceInterface = &server.TestAccountService{}
)
