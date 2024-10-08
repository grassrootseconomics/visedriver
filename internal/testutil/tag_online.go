// +build online

package testutil

const OnlineTestEnabled = true




var AccountService server.AccountServiceInterface


func init() {
    AccountService = &server.AccountService{}
}