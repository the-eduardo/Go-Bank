package db

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	// run n concurrent transfer transactions
	n := 5
	amount := int64(10)
	errs := make(chan error)
	results := make(chan TransferTxResult)
	for i := 0; i < n; i++ {
		go func() {
			result, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			})
			errs <- err
			results <- result
		}()
	}
	//check results
	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err, "TransferTx")
		result := <-results
		require.NotEmptyf(t, result, "transfer is empty")
		// transfer
		transfer := result.Transfer
		require.NotEmptyf(t, transfer.ID, "transfer.ID is empty")
		require.Equal(t, account1.ID, transfer.FromAccountID, "transfer.FromAccountID is not equal")
		require.Equal(t, account2.ID, transfer.ToAccountID, "transfer.ToAccountID is not equal")
		require.Equal(t, amount, transfer.Amount, "transfer.Amount is not equal")
		require.NotEqualf(t, account1.ID, transfer.ToAccountID, "AccountID is equal")
		require.NotZero(t, transfer.ID, "transfer.ID is zero")
		require.NotZero(t, transfer.CreatedAt, "transfer.CreatedAt is zero")

		_, err = store.GetTransfer(context.Background(), transfer.ID)
		require.NoError(t, err, "GetTransfer")

		// check entries
		fromEntry := result.FromEntry
		require.NotEmptyf(t, fromEntry, "entries is empty")
		require.Equal(t, account1.ID, fromEntry.AccountID, "fromEntry.AccountID is not equal")
		require.Equal(t, -amount, fromEntry.Amount, "fromEntry.Amount is not equal")
		require.NotZero(t, fromEntry.ID, "fromEntry.ID is zero")
		require.NotZero(t, fromEntry.CreatedAt, "fromEntry.CreatedAt is zero")

		_, err = store.GetEntries(context.Background(), fromEntry.ID)
		require.NoError(t, err, "GetEntry")

		toEntry := result.ToEntry
		require.NotEmptyf(t, toEntry, "entries is empty")
		require.Equal(t, account2.ID, toEntry.AccountID, "toEntry.AccountID is not equal")
		require.Equal(t, amount, toEntry.Amount, "toEntry.Amount is not equal")
		require.NotZero(t, toEntry.ID, "toEntry.ID is zero")
		require.NotZero(t, toEntry.CreatedAt, "toEntry.CreatedAt is zero")

		_, err = store.GetEntries(context.Background(), toEntry.ID)
		require.NoError(t, err, "GetEntry")

		// TODO: check balance
	}
}
