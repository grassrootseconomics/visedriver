package config

import "git.grassecon.net/urdt/ussd/initializers"

var (
	CreateAccountURL string
	TrackStatusURL   string
	BalanceURL       string
)

// LoadConfig initializes the configuration values after environment variables are loaded.
func LoadConfig() {
	CreateAccountURL = initializers.GetEnv("CREATE_ACCOUNT_URL", "https://custodial.sarafu.africa/api/account/create")
	TrackStatusURL = initializers.GetEnv("TRACK_STATUS_URL", "https://custodial.sarafu.africa/api/track/")
	BalanceURL = initializers.GetEnv("BALANCE_URL", "https://custodial.sarafu.africa/api/account/status/")
}
