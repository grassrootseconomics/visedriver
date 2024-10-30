package utils

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/stretchr/testify/require"

	"git.grassecon.net/urdt/ussd/internal/storage"
	"git.grassecon.net/urdt/ussd/common"
	memdb "git.defalsify.org/vise.git/db/mem"
	dataserviceapi "github.com/grassrootseconomics/ussd-data-service/pkg/api"
)

// InitializeTestDb sets up and returns an in-memory database and store.
func InitializeTestDb(t *testing.T) (context.Context, *common.UserDataStore) {
	ctx := context.Background()

	// Initialize memDb
	db := memdb.NewMemDb()
	err := db.Connect(ctx, "")
	require.NoError(t, err, "Failed to connect to memDb")

	// Create common.UserDataStore with memDb
	store := &common.UserDataStore{Db: db}

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
	holdings := []dataserviceapi.TokenHoldings{
		{ContractAddress: "0xd4c288865Ce", TokenSymbol: "SRF", TokenDecimals: "6", Balance: "100"},
		{ContractAddress: "0x41c188d63Qa", TokenSymbol: "MILO", TokenDecimals: "4", Balance: "200"},
	}

	expectedResult := VoucherMetadata{
		Symbols:   "1:SRF\n2:MILO",
		Balances:  "1:100\n2:200",
		Decimals:  "1:6\n2:4",
		Addresses: "1:0xd4c288865Ce\n2:0x41c188d63Qa",
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
	assert.Equal(t, "SRF", result.TokenSymbol)
	assert.Equal(t, "100", result.Balance)
	assert.Equal(t, "6", result.TokenDecimals)
	assert.Equal(t, "0xd4c288865Ce", result.ContractAddress)
}

func TestStoreTemporaryVoucher(t *testing.T) {
	ctx, store := InitializeTestDb(t)
	sessionId := "session123"

	// Test data
	voucherData := &dataserviceapi.TokenHoldings{
		TokenSymbol:     "SRF",
		Balance:         "200",
		TokenDecimals:   "6",
		ContractAddress: "0xd4c288865Ce0985a481Eef3be02443dF5E2e4Ea9",
	}

	// Execute the function being tested
	err := StoreTemporaryVoucher(ctx, store, sessionId, voucherData)
	require.NoError(t, err)

	// Verify stored data
	expectedEntries := map[common.DataTyp][]byte{
		common.DATA_TEMPORARY_SYM:     []byte("SRF"),
		common.DATA_TEMPORARY_BAL:     []byte("200"),
		common.DATA_TEMPORARY_DECIMAL: []byte("6"),
		common.DATA_TEMPORARY_ADDRESS: []byte("0xd4c288865Ce0985a481Eef3be02443dF5E2e4Ea9"),
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
	tempData := &dataserviceapi.TokenHoldings{
		TokenSymbol:     "SRF",
		Balance:         "200",
		TokenDecimals:   "6",
		ContractAddress: "0xd4c288865Ce0985a481Eef3be02443dF5E2e4Ea9",
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
	newData := &dataserviceapi.TokenHoldings{
		TokenSymbol:     "SRF",
		Balance:         "200",
		TokenDecimals:   "6",
		ContractAddress: "0xd4c288865Ce0985a481Eef3be02443dF5E2e4Ea9",
	}

	// Old temporary data
	tempData := &dataserviceapi.TokenHoldings{
		TokenSymbol:     "OLD",
		Balance:         "100",
		TokenDecimals:   "8",
		ContractAddress: "0xold",
	}
	require.NoError(t, StoreTemporaryVoucher(ctx, store, sessionId, tempData))

	// Execute update
	err := UpdateVoucherData(ctx, store, sessionId, newData)
	require.NoError(t, err)

	// Verify active data was stored correctly
	activeEntries := map[common.DataTyp][]byte{
		common.DATA_ACTIVE_SYM:     []byte(newData.TokenSymbol),
		common.DATA_ACTIVE_BAL:     []byte(newData.Balance),
		common.DATA_ACTIVE_DECIMAL: []byte(newData.TokenDecimals),
		common.DATA_ACTIVE_ADDRESS: []byte(newData.ContractAddress),
	}

	for key, expectedValue := range activeEntries {
		storedValue, err := store.ReadEntry(ctx, sessionId, key)
		require.NoError(t, err)
		require.Equal(t, expectedValue, storedValue, "Active data mismatch for key %v", key)
	}

	// Verify temporary data was cleared
	tempKeys := []common.DataTyp{
		common.DATA_TEMPORARY_SYM,
		common.DATA_TEMPORARY_BAL,
		common.DATA_TEMPORARY_DECIMAL,
		common.DATA_TEMPORARY_ADDRESS,
	}

	for _, key := range tempKeys {
		storedValue, err := store.ReadEntry(ctx, sessionId, key)
		require.NoError(t, err)
		AssertEmptyValue(t, storedValue, "Temporary data not cleared for key %v", key)
	}
}
