package db

import (
	"context"
	"database/sql"
	"github.com/stretchr/testify/require"
	"github.com/the-eduardo/Go-Bank/util"
	"testing"
	"time"
)

func createRandomAccount(t *testing.T) Account {
	user := createRandomUser(t)
	arg := CreateAccountParams{
		Owner:    user.Username,
		Balance:  util.RandomBalance(),
		Currency: util.RandomCurrency(),
	}
	account, err := testQueries.CreateAccount(context.Background(), arg)
	require.NoError(t, err, "CreateAccount")
	require.NotEmptyf(t, account.ID, "account.ID is empty")

	require.Equal(t, arg.Owner, account.Owner, "account.Owner is not equal")
	require.Equal(t, arg.Balance, account.Balance, "account.Balance is not equal")
	require.Equal(t, arg.Currency, account.Currency, "account.Currency is not equal")

	require.NotZero(t, account.ID, "account.ID is zero")
	require.NotZero(t, account.CreatedAt, "account.CreatedAt is zero")

	return account
}
func TestCreateAccount(t *testing.T) {
	createRandomAccount(t)
}
func TestGetAccount(t *testing.T) {
	account1 := createRandomAccount(t)
	account2, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err, "GetAccount")
	require.NotEmptyf(t, account2.ID, "account.ID is empty")

	require.Equal(t, account1.ID, account2.ID, "account.ID is not equal")
	require.Equal(t, account1.Owner, account2.Owner, "account.Owner is not equal")
	require.Equal(t, account1.Balance, account2.Balance, "account.Balance is not equal")
	require.Equal(t, account1.Currency, account2.Currency, "account.Currency is not equal")
	require.WithinDurationf(t, account1.CreatedAt, account2.CreatedAt, time.Second, "account.CreatedAt is not equal")
}

func TestUpdateAccount(t *testing.T) {
	account1 := createRandomAccount(t)

	arg := UpdateAccountParams{
		ID:      account1.ID,
		Balance: util.RandomBalance(),
	}
	account2, err := testQueries.UpdateAccount(context.Background(), arg)
	require.NoError(t, err, "UpdateAccount")
	require.NotEmptyf(t, account2.ID, "account.ID is empty")

	require.Equal(t, account1.ID, account2.ID, "account.ID is not equal")
	require.Equal(t, account1.Owner, account2.Owner, "account.Owner is not equal")
	require.Equal(t, arg.Balance, account2.Balance, "account.Balance is not equal")
	require.Equal(t, account1.Currency, account2.Currency, "account.Currency is not equal")
	require.WithinDurationf(t, account1.CreatedAt, account2.CreatedAt, time.Second, "account.CreatedAt is not equal")
}

func TestDeleteAccount(t *testing.T) {
	account1 := createRandomAccount(t)
	err := testQueries.DeleteAccount(context.Background(), account1.ID)
	require.NoError(t, err, "DeleteAccount")
	account2, err := testQueries.GetAccount(context.Background(), account1.ID)
	require.Error(t, err, "GetAccount")
	require.EqualError(t, err, sql.ErrNoRows.Error(), "account.ID is not equal")
	require.Empty(t, account2.ID, "account.ID is not empty")
}

func TestListAccounts(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomAccount(t)
	}
	arg := ListAccountsParams{
		Limit:  5,
		Offset: 5,
	}
	accounts, err := testQueries.ListAccounts(context.Background(), arg)
	require.NoError(t, err, "ListAccounts")
	require.Len(t, accounts, 5, "accounts.Len is not equal")

	for _, account := range accounts {
		require.NotEmptyf(t, account.ID, "account.ID is empty")
		require.NotEmptyf(t, account.Owner, "account.Owner is empty")
		require.NotZero(t, account.Balance, "account.Balance is zero")
		require.NotZero(t, account.Currency, "account.Currency is zero")
		require.NotZero(t, account.CreatedAt, "account.CreatedAt is zero")
	}
}
