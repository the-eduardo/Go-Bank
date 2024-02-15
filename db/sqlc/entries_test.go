package db

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/the-eduardo/Go-Bank/util"
	"testing"
	"time"
)

func createNewEntry(account Account, t *testing.T) Entry {
	arg := NewEntryParams{
		AccountID: account.ID,
		Amount:    util.RandomMoney(),
	}
	entry, err := testQueries.NewEntry(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, entry)

	require.Equal(t, arg.AccountID, entry.AccountID)
	require.Equal(t, arg.Amount, entry.Amount)

	require.NotZero(t, entry.ID)
	require.NotZero(t, entry.CreatedAt)

	return entry
}

func TestNewEntry(t *testing.T) {
	account := createRandomAccount(t)
	createNewEntry(account, t)
}

func TestGetEntry(t *testing.T) {
	account := createRandomAccount(t)
	entry1 := createNewEntry(account, t)
	entry2, err := testQueries.GetEntry(context.Background(), entry1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, entry2)
	require.Equal(t, entry1.ID, entry2.ID)
	require.Equal(t, entry1.AccountID, entry2.AccountID)
	require.Equal(t, entry1.Amount, entry2.Amount)
	require.WithinDuration(t, entry1.CreatedAt.Time, entry2.CreatedAt.Time, time.Second)
}

func TestListEntries(t *testing.T) {
	account := createRandomAccount(t)

	n := 10
	for i := 0; i < n; i++ {
		createNewEntry(account, t)
	}
	arg := ListEntriesParams{
		AccountID: account.ID,
		Limit:     5,
		Offset:    5,
	}
	entries, err := testQueries.ListEntries(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, entries, 5)
	for _, entry := range entries {
		require.NotEmpty(t, entry)
	}
}
