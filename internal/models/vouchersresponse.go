package models

// VoucherHolding represents a single voucher holding
type VoucherHolding struct {
	Symbol  string `json:"symbol"`
	Address string `json:"address"`
}