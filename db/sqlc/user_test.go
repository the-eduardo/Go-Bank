package db

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/the-eduardo/Go-Bank/util"
	"testing"
	"time"
)

func createRandomUser(t *testing.T) User {
	arg := CreateUserParams{
		Username:       util.RandomOwner(),
		HashedPassword: "secret",
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}

	user, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err, "CreateUser")
	require.NotEmptyf(t, user, "user is empty")

	require.Equal(t, arg.Username, user.Username, "Username is not equal")
	require.Equal(t, arg.Email, user.Email, "Email is not equal")
	require.Equal(t, arg.HashedPassword, user.HashedPassword, "Wrong password")
	require.Equal(t, arg.FullName, user.FullName, "FullName is not equal")

	require.Truef(t, user.PasswordChangedAt.IsZero(), "PasswordChangedAt is not zero")
	require.NotZero(t, user.CreatedAt, "user.CreatedAt is zero")

	return user
}
func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}
func TestGetUser(t *testing.T) {
	user1 := createRandomUser(t)
	user2, err := testQueries.GetUser(context.Background(), user1.Username)
	require.NoError(t, err, "GetUser")
	require.NotEmptyf(t, user2.Username, "Username is empty")

	require.Equal(t, user1.Username, user2.Username, "Username is not equal")
	require.Equal(t, user1.Email, user2.Email, "Email is not equal")
	require.Equal(t, user1.HashedPassword, user2.HashedPassword, "Wrong password")
	require.Equal(t, user1.FullName, user2.FullName, "user.FullName is not equal")
	require.WithinDurationf(t, user1.CreatedAt, user2.CreatedAt, time.Second, "user.CreatedAt is not equal")
	require.WithinDurationf(t, user1.PasswordChangedAt, user2.PasswordChangedAt, time.Second, "user.CreatedAt is not equal")

}
