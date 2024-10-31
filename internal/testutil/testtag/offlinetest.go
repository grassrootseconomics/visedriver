// +build !online

package testtag

import (
	"git.grassecon.net/urdt/ussd/internal/handlers/server"
	accountservice "git.grassecon.net/urdt/ussd/internal/testutil/testservice"
)

var (
	AccountService server.AccountServiceInterface = &accountservice.TestAccountService{}
)
