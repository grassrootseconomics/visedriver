package models

type VoucherDataResult struct {
	TokenName     string `json:"tokenName"`
	TokenSymbol   string `json:"tokenSymbol"`
	TokenDecimals string `json:"tokenDecimals"`
	SinkAddress   string `json:"sinkAddress"`
}
