package db

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/techschool/simplebank/util"
)

func createRandomEntry(t *testing.T, account Account) Entry {
	arg := CreateEntriesParams{
		AccountID: account.ID,
		Amount:    util.RandomBalance(),
	}

	entry, err := testQueries.CreateEntries(context.Background(), arg)
	require.NoError(t, err, "CreateEntries")
	require.NotEmpty(t, entry, "entry.ID is empty")

	require.Equal(t, arg.AccountID, entry.AccountID, "entry.AccountID is not equal")
	require.Equal(t, arg.Amount, entry.Amount, "entry.Amount is not equal")

	require.NotZero(t, entry.ID, "entry.ID is zero")
	require.NotZero(t, entry.CreatedAt, "entry.CreatedAt is zero")

	return entry
}

func TestCreateEntry(t *testing.T) {
	account := createRandomAccount(t)
	createRandomEntry(t, account)
}

func TestGetEntry(t *testing.T) {
	account := createRandomAccount(t)
	entry1 := createRandomEntry(t, account)
	entry2, err := testQueries.GetEntries(context.Background(), entry1.ID)
	require.NoError(t, err, "GetEntries")
	require.NotEmpty(t, entry2, "entry.ID is empty")

	require.Equal(t, entry1.ID, entry2.ID, "entry.ID is not equal")
	require.Equal(t, entry1.AccountID, entry2.AccountID, "entry.AccountID is not equal")
	require.Equal(t, entry1.Amount, entry2.Amount, "entry.Amount is not equal")
	require.WithinDuration(t, entry1.CreatedAt, entry2.CreatedAt, time.Second, "entry.CreatedAt is not equal")
}

func TestListEntries(t *testing.T) {
	account := createRandomAccount(t)
	for i := 0; i < 10; i++ {
		createRandomEntry(t, account)
	}

	arg := ListEntriesParams{
		AccountID: account.ID,
		Limit:     5,
		Offset:    5,
	}

	entries, err := testQueries.ListEntries(context.Background(), arg)
	require.NoError(t, err, "ListEntries")
	require.Len(t, entries, 5, "entries length is not 5")

	for _, entry := range entries {
		require.NotEmpty(t, entry, "entry.ID is empty")
		require.Equal(t, arg.AccountID, entry.AccountID, "entry.AccountID is not equal")
	}
}
