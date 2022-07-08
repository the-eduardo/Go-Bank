package db

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/the-eduardo/Go-Bank/util"
	"testing"
	"time"
)

func createRandomTransfer(t *testing.T, account, account2 Account) Transfer {
	arg := CreateTransferParams{
		FromAccountID: account.ID,
		ToAccountID:   account2.ID,
		Amount:        util.RandomBalance(),
	}
	transfer, err := testQueries.CreateTransfer(context.Background(), arg)
	require.NoError(t, err, "CreateTransfer")
	require.NotEmptyf(t, transfer.ID, "transfer.ID is empty")

	require.Equal(t, arg.FromAccountID, transfer.FromAccountID, "transfer.FromAccountID is not equal")
	require.Equal(t, arg.ToAccountID, transfer.ToAccountID, "transfer.ToAccountID is not equal")
	require.Equal(t, arg.Amount, transfer.Amount, "transfer.Amount is not equal")

	require.NotEqualf(t, arg.FromAccountID, transfer.ToAccountID, "AccountID is equal")

	require.NotZero(t, transfer.ID, "transfer.ID is zero")
	require.NotZero(t, transfer.CreatedAt, "transfer.CreatedAt is zero")

	return transfer
}
func TestCreateTransfer(t *testing.T) {
	account := createRandomAccount(t)
	account2 := createRandomAccount(t)
	createRandomTransfer(t, account, account2)
}

func TestGetTransfer(t *testing.T) {
	account := createRandomAccount(t)
	account2 := createRandomAccount(t)
	transfer := createRandomTransfer(t, account, account2)

	transfer2, err := testQueries.GetTransfer(context.Background(), transfer.ID)
	require.NoError(t, err, "GetTransfer")
	require.NotEmptyf(t, transfer2.ID, "transfer.ID is empty")

	require.Equal(t, transfer.ID, transfer2.ID, "transfer.ID is not equal")
	require.Equal(t, transfer.FromAccountID, transfer2.FromAccountID, "transfer.FromAccountID is not equal")
	require.Equal(t, transfer.ToAccountID, transfer2.ToAccountID, "transfer.ToAccountID is not equal")
	require.Equal(t, transfer.Amount, transfer2.Amount, "transfer.Amount is not equal")
	require.WithinDurationf(t, transfer.CreatedAt, transfer2.CreatedAt, time.Second, "transfer.CreatedAt is not equal")
}

func TestListTransfer(t *testing.T) {
	account := createRandomAccount(t)
	account2 := createRandomAccount(t)
	for i := 0; i < 10; i++ {
		createRandomTransfer(t, account, account2)
		createRandomTransfer(t, account2, account)
	}
	arg := ListTransferParams{
		FromAccountID: account.ID,
		ToAccountID:   account2.ID,
		Limit:         5,
		Offset:        5,
	}
	transfers, err := testQueries.ListTransfer(context.Background(), arg)
	require.NoError(t, err, "ListTransfer")
	require.Len(t, transfers, 5, "transfers is not 5")

	for _, transfer := range transfers {
		require.NotEmptyf(t, transfer.ID, "transfer.ID is empty")
		require.True(t, transfer.FromAccountID == account.ID || transfer.FromAccountID == account2.ID, "transfer.FromAccountID is not equal")

	}
}
