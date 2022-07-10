package db

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testDB)
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)
	fmt.Println(">> before:", account1.Balance, account2.Balance)
	// run n concurrent transfer transactions
	n := 5
	amount := int64(10)
	errs := make(chan error)
	results := make(chan TransferTxResult)
	for i := 0; i < n; i++ {
		//txName := fmt.Sprintf("tx %d", i+1)
		go func() {
			//ctx := context.WithValue(context.Background(), txKey, txName)
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
	existed := make(map[int]bool)

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

		// check accounts
		fromAccount := result.FromAccount
		require.NotEmptyf(t, fromAccount, "From account is empty")
		require.Equal(t, account1.ID, fromAccount.ID, "fromAccount.ID is not equal")

		toAccount := result.ToAccount
		require.NotEmptyf(t, toAccount, "To account is empty")
		require.Equal(t, account2.ID, toAccount.ID, "toAccount.ID is not equal")
		// check balance
		fmt.Println(">> tx:", fromAccount.Balance, toAccount.Balance)
		diff1 := account1.Balance - fromAccount.Balance
		diff2 := account2.Balance - toAccount.Balance
		require.NotEqual(t, diff1, diff2, "amount is not equal")
		require.True(t, diff1 > 0, "amount is not positive")
		require.True(t, diff1%amount == 0, "amount is not multiple of 10")

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= n, "k is not positive")
		require.NotContains(t, existed, k)
		existed[k] = true
	}
	// check final balance
	updatedAccount1, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err, "GetAccount")
	updatedAccount2, err := testQueries.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err, "GetAccount")

	fmt.Println(">> after:", updatedAccount1.Balance, updatedAccount2.Balance)
	require.Equal(t, account1.Balance-int64(n)*amount, updatedAccount1.Balance, "Final balance is not equal")
	require.Equal(t, account2.Balance+int64(n)*amount, updatedAccount2.Balance, "Final balance is not equal")
}
