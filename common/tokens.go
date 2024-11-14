package common

import (
	"context"
	"math/big"
	"strconv"
)

func ParseAndScaleAmount(storedAmount, activeDecimal []byte) (string, error) {
	// Parse token decimal
	tokenDecimal, err := strconv.Atoi(string(activeDecimal))
	if err != nil {

		return "", err
	}

	// Parse amount
	amount, _, err := big.ParseFloat(string(storedAmount), 10, 0, big.ToZero)
	if err != nil {
		return "", err
	}

	// Scale the amount
	multiplier := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(tokenDecimal)), nil))
	finalAmount := new(big.Float).Mul(amount, multiplier)

	// Convert finalAmount to a string
	finalAmountStr := new(big.Int)
	finalAmount.Int(finalAmountStr)

	return finalAmountStr.String(), nil
}

func ReadTransactionData(ctx context.Context, store DataStore, sessionId string) (map[DataTyp][]byte, error) {
	dataKeys := []DataTyp{
		DATA_TEMPORARY_VALUE,
		DATA_ACTIVE_SYM,
		DATA_AMOUNT,
		DATA_PUBLIC_KEY,
		DATA_RECIPIENT,
		DATA_ACTIVE_DECIMAL,
		DATA_ACTIVE_ADDRESS,
	}

	data := make(map[DataTyp][]byte)
	for _, key := range dataKeys {
		value, err := store.ReadEntry(ctx, sessionId, key)
		if err != nil {
			return nil, err
		}
		data[key] = value
	}
	return data, nil
}
