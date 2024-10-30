package utils

import (
	"context"
	"fmt"
	"strings"

	"git.grassecon.net/urdt/ussd/internal/storage"
	dataserviceapi "github.com/grassrootseconomics/ussd-data-service/pkg/api"
)

// VoucherMetadata helps organize voucher data fields
type VoucherMetadata struct {
	Symbol  string
	Balance string
	Decimal string
	Address string
}

// ProcessVouchers converts holdings into formatted strings
func ProcessVouchers(holdings []dataserviceapi.TokenHoldings) VoucherMetadata {
	var data VoucherMetadata
	var symbols, balances, decimals, addresses []string

	for i, h := range holdings {
		symbols = append(symbols, fmt.Sprintf("%d:%s", i+1, h.TokenSymbol))
		balances = append(balances, fmt.Sprintf("%d:%s", i+1, h.Balance))
		decimals = append(decimals, fmt.Sprintf("%d:%s", i+1, h.TokenDecimals))
		addresses = append(addresses, fmt.Sprintf("%d:%s", i+1, h.ContractAddress))
	}

	data.Symbol = strings.Join(symbols, "\n")
	data.Balance = strings.Join(balances, "\n")
	data.Decimal = strings.Join(decimals, "\n")
	data.Address = strings.Join(addresses, "\n")

	return data
}

// GetVoucherData retrieves and matches voucher data
func GetVoucherData(ctx context.Context, db storage.PrefixDb, input string) (*dataserviceapi.TokenHoldings, error) {
	keys := []string{"sym", "bal", "deci", "addr"}
	data := make(map[string]string)

	for _, key := range keys {
		value, err := db.Get(ctx, []byte(key))
		if err != nil {
			return nil, fmt.Errorf("failed to get %s: %v", key, err)
		}
		data[key] = string(value)
	}

	symbol, balance, decimal, address := MatchVoucher(input,
		data["sym"],
		data["bal"],
		data["deci"],
		data["addr"])

	if symbol == "" {
		return nil, nil
	}

	return &dataserviceapi.TokenHoldings{
		TokenSymbol:     string(symbol),
		Balance:         string(balance),
		TokenDecimals:   string(decimal),
		ContractAddress: string(address),
	}, nil
}

// MatchVoucher finds the matching voucher symbol, balance, decimals and contract address based on the input.
func MatchVoucher(input, symbols, balances, decimals, addresses string) (symbol, balance, decimal, address string) {
	symList := strings.Split(symbols, "\n")
	balList := strings.Split(balances, "\n")
	decList := strings.Split(decimals, "\n")
	addrList := strings.Split(addresses, "\n")

	for i, sym := range symList {
		parts := strings.SplitN(sym, ":", 2)
		if len(parts) != 2 {
			continue
		}

		if input == parts[0] || strings.EqualFold(input, parts[1]) {
			symbol = parts[1]
			if i < len(balList) {
				balance = strings.SplitN(balList[i], ":", 2)[1]
			}
			if i < len(decList) {
				decimal = strings.SplitN(decList[i], ":", 2)[1]
			}
			if i < len(addrList) {
				address = strings.SplitN(addrList[i], ":", 2)[1]
			}
			break
		}
	}
	return
}

// StoreTemporaryVoucher saves voucher metadata as temporary entries in the DataStore.
func StoreTemporaryVoucher(ctx context.Context, store DataStore, sessionId string, data *dataserviceapi.TokenHoldings) error {
	entries := map[DataTyp][]byte{
		DATA_TEMPORARY_SYM:     []byte(data.TokenSymbol),
		DATA_TEMPORARY_BAL:     []byte(data.Balance),
		DATA_TEMPORARY_DECIMAL: []byte(data.TokenDecimals),
		DATA_TEMPORARY_ADDRESS: []byte(data.ContractAddress),
	}

	for key, value := range entries {
		if err := store.WriteEntry(ctx, sessionId, key, value); err != nil {
			return err
		}
	}
	return nil
}

// GetTemporaryVoucherData retrieves temporary voucher metadata from the DataStore.
func GetTemporaryVoucherData(ctx context.Context, store DataStore, sessionId string) (*dataserviceapi.TokenHoldings, error) {
	keys := []DataTyp{
		DATA_TEMPORARY_SYM,
		DATA_TEMPORARY_BAL,
		DATA_TEMPORARY_DECIMAL,
		DATA_TEMPORARY_ADDRESS,
	}

	data := &dataserviceapi.TokenHoldings{}
	values := make([][]byte, len(keys))

	for i, key := range keys {
		value, err := store.ReadEntry(ctx, sessionId, key)
		if err != nil {
			return nil, err
		}
		values[i] = value
	}

	data.TokenSymbol = string(values[0])
	data.Balance = string(values[1])
	data.TokenDecimals = string(values[2])
	data.ContractAddress = string(values[3])

	return data, nil
}

// UpdateVoucherData sets the active voucher data and clears the temporary voucher data in the DataStore.
func UpdateVoucherData(ctx context.Context, store DataStore, sessionId string, data *dataserviceapi.TokenHoldings) error {
	// Active voucher data entries
	activeEntries := map[DataTyp][]byte{
		DATA_ACTIVE_SYM:     []byte(data.TokenSymbol),
		DATA_ACTIVE_BAL:     []byte(data.Balance),
		DATA_ACTIVE_DECIMAL: []byte(data.TokenDecimals),
		DATA_ACTIVE_ADDRESS: []byte(data.ContractAddress),
	}

	// Clear temporary voucher data entries
	tempEntries := map[DataTyp][]byte{
		DATA_TEMPORARY_SYM:     []byte(""),
		DATA_TEMPORARY_BAL:     []byte(""),
		DATA_TEMPORARY_DECIMAL: []byte(""),
		DATA_TEMPORARY_ADDRESS: []byte(""),
	}

	// Write active data
	for key, value := range activeEntries {
		if err := store.WriteEntry(ctx, sessionId, key, value); err != nil {
			return err
		}
	}

	// Clear temporary data
	for key, value := range tempEntries {
		if err := store.WriteEntry(ctx, sessionId, key, value); err != nil {
			return err
		}
	}

	return nil
}
