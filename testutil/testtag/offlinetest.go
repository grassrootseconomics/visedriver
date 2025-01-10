// +build !online

package testtag

import (
	"git.grassecon.net/grassrootseconomics/visedriver/remote"
	accountservice "git.grassecon.net/grassrootseconomics/visedriver/internal/testutil/testservice"
)

var (
	AccountService remote.AccountServiceInterface = &accountservice.TestAccountService{}
)
