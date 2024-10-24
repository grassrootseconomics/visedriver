package config

import "git.grassecon.net/urdt/ussd/initializers"

var (
	CreateAccountURL string
	TrackStatusURL   string
	BalanceURL       string
    TrackURL         string
)

// LoadConfig initializes the configuration values after environment variables are loaded.
func LoadConfig() {
	CreateAccountURL = initializers.GetEnv("CREATE_ACCOUNT_URL", "http://localhost:5003/api/v2/account/creates")
	TrackStatusURL = initializers.GetEnv("TRACK_STATUS_URL", "https://custodial.sarafu.africa/api/track/")
	BalanceURL = initializers.GetEnv("BALANCE_URL", "https://custodial.sarafu.africa/api/account/status/")
    TrackURL = initializers.GetEnv("TRACK_URL", "http://localhost:5003/api/v2/account/status")
}
