package utils

import (
	"context"
	"testing"

	"git.grassecon.net/urdt/ussd/internal/storage"
	"github.com/alecthomas/assert/v2"
	"github.com/stretchr/testify/require"

	memdb "git.defalsify.org/vise.git/db/mem"
)

// InitializeTestDb sets up and returns an in-memory database and store.
func InitializeTestDb(t *testing.T) (context.Context, *UserDataStore) {
	ctx := context.Background()

	// Initialize memDb
	db := memdb.NewMemDb()
	err := db.Connect(ctx, "")
	require.NoError(t, err, "Failed to connect to memDb")

	// Create UserDataStore with memDb
	store := &UserDataStore{Db: db}

	t.Cleanup(func() {
		db.Close() // Ensure the DB is closed after each test
	})

	return ctx, store
}

// AssertEmptyValue checks if a value is empty/nil/zero
func AssertEmptyValue(t *testing.T, value []byte, msgAndArgs ...interface{}) {
	assert.Equal(t, len(value), 0, msgAndArgs...)
}

func TestMatchVoucher(t *testing.T) {
	symbols := "1:SRF\n2:MILO"
	balances := "1:100\n2:200"
	decimals := "1:6\n2:4"
	addresses := "1:0xd4c288865Ce\n2:0x41c188d63Qa"

	// Test for valid voucher
	symbol, balance, decimal, address := MatchVoucher("2", symbols, balances, decimals, addresses)

	// Assertions for valid voucher
	assert.Equal(t, "MILO", symbol)
	assert.Equal(t, "200", balance)
	assert.Equal(t, "4", decimal)
	assert.Equal(t, "0x41c188d63Qa", address)

	// Test for non-existent voucher
	symbol, balance, decimal, address = MatchVoucher("3", symbols, balances, decimals, addresses)

	// Assertions for non-match
	assert.Equal(t, "", symbol)
	assert.Equal(t, "", balance)
	assert.Equal(t, "", decimal)
	assert.Equal(t, "", address)
}

func TestProcessVouchers(t *testing.T) {
	holdings := []struct {
		ContractAddress string `json:"contractAddress"`
		TokenSymbol     string `json:"tokenSymbol"`
		TokenDecimals   string `json:"tokenDecimals"`
		Balance         string `json:"balance"`
	}{
		{ContractAddress: "0xd4c288865Ce", TokenSymbol: "SRF", TokenDecimals: "6", Balance: "100"},
		{ContractAddress: "0x41c188d63Qa", TokenSymbol: "MILO", TokenDecimals: "4", Balance: "200"},
	}

	expectedResult := VoucherMetadata{
		Symbol:  "1:SRF\n2:MILO",
		Balance: "1:100\n2:200",
		Decimal: "1:6\n2:4",
		Address: "1:0xd4c288865Ce\n2:0x41c188d63Qa",
	}

	result := ProcessVouchers(holdings)

	assert.Equal(t, expectedResult, result)
}

func TestGetVoucherData(t *testing.T) {
	ctx := context.Background()

	db := memdb.NewMemDb()
	err := db.Connect(ctx, "")
	if err != nil {
		t.Fatal(err)
	}
	spdb := storage.NewSubPrefixDb(db, []byte("vouchers"))

	// Test voucher data
	mockData := map[string][]byte{
		"sym":  []byte("1:SRF\n2:MILO"),
		"bal":  []byte("1:100\n2:200"),
		"deci": []byte("1:6\n2:4"),
		"addr": []byte("1:0xd4c288865Ce\n2:0x41c188d63Qa"),
	}

	// Put the data
	for key, value := range mockData {
		err = spdb.Put(ctx, []byte(key), []byte(value))
		if err != nil {
			t.Fatal(err)
		}
	}

	result, err := GetVoucherData(ctx, spdb, "1")

	assert.NoError(t, err)
	assert.Equal(t, "SRF", result.Symbol)
	assert.Equal(t, "100", result.Balance)
	assert.Equal(t, "6", result.Decimal)
	assert.Equal(t, "0xd4c288865Ce", result.Address)
}

func TestStoreTemporaryVoucher(t *testing.T) {
	ctx, store := InitializeTestDb(t)
	sessionId := "session123"

	// Test data
	voucherData := &VoucherMetadata{
		Symbol:  "SRF",
		Balance: "200",
		Decimal: "6",
		Address: "0xd4c288865Ce0985a481Eef3be02443dF5E2e4Ea9",
	}

	// Execute the function being tested
	err := StoreTemporaryVoucher(ctx, store, sessionId, voucherData)
	require.NoError(t, err)

	// Verify stored data
	expectedEntries := map[DataTyp][]byte{
		DATA_TEMPORARY_SYM:     []byte("SRF"),
		DATA_TEMPORARY_BAL:     []byte("200"),
		DATA_TEMPORARY_DECIMAL: []byte("6"),
		DATA_TEMPORARY_ADDRESS: []byte("0xd4c288865Ce0985a481Eef3be02443dF5E2e4Ea9"),
	}

	for key, expectedValue := range expectedEntries {
		storedValue, err := store.ReadEntry(ctx, sessionId, key)
		require.NoError(t, err)
		require.Equal(t, expectedValue, storedValue, "Mismatch for key %v", key)
	}
}

func TestGetTemporaryVoucherData(t *testing.T) {
	ctx, store := InitializeTestDb(t)
	sessionId := "session123"

	// Test voucher data
	tempData := &VoucherMetadata{
		Symbol:  "SRF",
		Balance: "200",
		Decimal: "6",
		Address: "0xd4c288865Ce0985a481Eef3be02443dF5E2e4Ea9",
	}

	// Store the data
	err := StoreTemporaryVoucher(ctx, store, sessionId, tempData)
	require.NoError(t, err)

	// Execute the function being tested
	data, err := GetTemporaryVoucherData(ctx, store, sessionId)
	require.NoError(t, err)
	require.Equal(t, tempData, data)
}

func TestUpdateVoucherData(t *testing.T) {
	ctx, store := InitializeTestDb(t)
	sessionId := "session123"

	// New voucher data
	newData := &VoucherMetadata{
		Symbol:  "SRF",
		Balance: "200",
		Decimal: "6",
		Address: "0xd4c288865Ce0985a481Eef3be02443dF5E2e4Ea9",
	}

	// Old temporary data
	tempData := &VoucherMetadata{
		Symbol:  "OLD",
		Balance: "100",
		Decimal: "8",
		Address: "0xold",
	}
	require.NoError(t, StoreTemporaryVoucher(ctx, store, sessionId, tempData))

	// Execute update
	err := UpdateVoucherData(ctx, store, sessionId, newData)
	require.NoError(t, err)

	// Verify active data was stored correctly
	activeEntries := map[DataTyp][]byte{
		DATA_ACTIVE_SYM:     []byte(newData.Symbol),
		DATA_ACTIVE_BAL:     []byte(newData.Balance),
		DATA_ACTIVE_DECIMAL: []byte(newData.Decimal),
		DATA_ACTIVE_ADDRESS: []byte(newData.Address),
	}

	for key, expectedValue := range activeEntries {
		storedValue, err := store.ReadEntry(ctx, sessionId, key)
		require.NoError(t, err)
		require.Equal(t, expectedValue, storedValue, "Active data mismatch for key %v", key)
	}

	// Verify temporary data was cleared
	tempKeys := []DataTyp{
		DATA_TEMPORARY_SYM,
		DATA_TEMPORARY_BAL,
		DATA_TEMPORARY_DECIMAL,
		DATA_TEMPORARY_ADDRESS,
	}

	for _, key := range tempKeys {
		storedValue, err := store.ReadEntry(ctx, sessionId, key)
		require.NoError(t, err)
		AssertEmptyValue(t, storedValue, "Temporary data not cleared for key %v", key)
	}
}
