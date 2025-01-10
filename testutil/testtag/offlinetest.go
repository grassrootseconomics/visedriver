// +build !online

package testtag

import (
	"git.grassecon.net/grassrootseconomics/visedriver/remote"
	accountservice "git.grassecon.net/grassrootseconomics/visedriver/testutil/testservice"
)

var (
	AccountService remote.AccountServiceInterface = &accountservice.TestAccountService{}
)
