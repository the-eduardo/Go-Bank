package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"github.com/the-eduardo/Go-Bank/util"
	"testing"
	"time"
)

func createRandomUser(t *testing.T) User {
	hashedPassword, err := util.HashPassword(util.RandomString(6))
	require.NoError(t, err)

	arg := CreateUserParams{
		Username:       util.RandomOwner(),
		HashedPassword: hashedPassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}
	user, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.HashedPassword, user.HashedPassword)
	require.Equal(t, arg.FullName, user.FullName)
	require.Equal(t, arg.Email, user.Email)

	require.NotZero(t, user.CreatedAt)

	return user

}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUser(t *testing.T) {

	user1 := createRandomUser(t)
	require.NotEmpty(t, user1)

	user2, err := testQueries.GetUser(context.Background(), user1.Username)
	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.Username, user2.Username)
	require.Equal(t, user1.FullName, user2.FullName)
	require.Equal(t, user1.Email, user2.Email)
	require.Equal(t, user1.HashedPassword, user2.HashedPassword)
	require.WithinDuration(t, user1.CreatedAt.Time, user2.CreatedAt.Time, time.Second)
	require.WithinDuration(t, user1.PasswordChangedAt.Time, user2.PasswordChangedAt.Time, time.Second)

}

func TestUpdateFullName(t *testing.T) {
	oldName := createRandomUser(t)
	require.NotEmpty(t, oldName)
	newName := util.RandomOwner()
	arg := UpdateUserParams{
		Username: oldName.Username,
		FullName: pgtype.Text{
			String: newName,
			Valid:  true,
		},
	}
	updatedUser, err := testQueries.UpdateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, updatedUser)
	require.Equal(t, oldName.Username, updatedUser.Username)
	require.Equal(t, arg.FullName.String, updatedUser.FullName)
	require.Equal(t, oldName.Email, updatedUser.Email)
	require.Equal(t, oldName.HashedPassword, updatedUser.HashedPassword)
	require.WithinDuration(t, oldName.CreatedAt.Time, updatedUser.CreatedAt.Time, time.Second)
	require.WithinDuration(t, oldName.PasswordChangedAt.Time, updatedUser.PasswordChangedAt.Time, time.Second)
}

//func TestUpdateUsername(t *testing.T) {
//	oldUser := createRandomUser(t)
//	require.NotEmpty(t, oldUser)
//	newUsername := util.RandomString(10)
//	arg := UpdateUserParams{
//		Username: newUsername,
//	}
//	updatedUser, err := testQueries.UpdateUser(context.Background(), arg)
//	require.NoError(t, err)
//	require.NotEmpty(t, updatedUser)
//	require.Equal(t, arg.Username, updatedUser.Username)
//	require.Equal(t, oldUser.FullName, updatedUser.FullName)
//	require.Equal(t, oldUser.Email, updatedUser.Email)
//	require.Equal(t, oldUser.HashedPassword, updatedUser.HashedPassword)
//	require.WithinDuration(t, oldUser.CreatedAt.Time, updatedUser.CreatedAt.Time, time.Second)
//	require.WithinDuration(t, oldUser.PasswordChangedAt.Time, updatedUser.PasswordChangedAt.Time, time.Second)
//}

func TestUpdateUserEmail(t *testing.T) {
	oldUser := createRandomUser(t)
	require.NotEmpty(t, oldUser)
	newMail := util.RandomEmail()
	arg := UpdateUserParams{
		Username: oldUser.Username,
		Email: pgtype.Text{
			String: newMail,
			Valid:  true,
		},
	}
	updatedUser, err := testQueries.UpdateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, updatedUser)
	require.Equal(t, oldUser.Username, updatedUser.Username)
	require.Equal(t, oldUser.FullName, updatedUser.FullName)
	require.Equal(t, arg.Email.String, updatedUser.Email)
	require.Equal(t, oldUser.HashedPassword, updatedUser.HashedPassword)
	require.WithinDuration(t, oldUser.CreatedAt.Time, updatedUser.CreatedAt.Time, time.Second)
	require.WithinDuration(t, oldUser.PasswordChangedAt.Time, updatedUser.PasswordChangedAt.Time, time.Second)
}

func TestUpdateUserPassword(t *testing.T) {
	oldUser := createRandomUser(t)
	require.NotEmpty(t, oldUser)
	newPassword, err := util.HashPassword(util.RandomString(10))
	require.NoError(t, err)
	arg := UpdateUserParams{
		Username: oldUser.Username,
		HashedPassword: pgtype.Text{
			String: newPassword,
			Valid:  true,
		},
	}
	updatedUser, err := testQueries.UpdateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, updatedUser)
	require.Equal(t, oldUser.Username, updatedUser.Username)
	require.Equal(t, oldUser.FullName, updatedUser.FullName)
	require.Equal(t, oldUser.Email, updatedUser.Email)
	require.Equal(t, arg.HashedPassword.String, updatedUser.HashedPassword)
	require.WithinDuration(t, oldUser.CreatedAt.Time, updatedUser.CreatedAt.Time, time.Second)
	require.WithinDuration(t, oldUser.PasswordChangedAt.Time, updatedUser.PasswordChangedAt.Time, time.Second)
}

func TestUpdateAllUserFields(t *testing.T) {
	oldUser := createRandomUser(t)
	require.NotEmpty(t, oldUser)
	newPassword, err := util.HashPassword(util.RandomString(10))
	newFullName := util.RandomOwner()
	newEmail := util.RandomEmail()

	require.NoError(t, err)
	arg := UpdateUserParams{
		Username: oldUser.Username,
		HashedPassword: pgtype.Text{
			String: newPassword,
			Valid:  true,
		},
		FullName: pgtype.Text{
			String: newFullName,
			Valid:  true,
		},
		Email: pgtype.Text{
			String: newEmail,
			Valid:  true,
		},
	}
	updatedUser, err := testQueries.UpdateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, updatedUser)
	require.Equal(t, oldUser.Username, updatedUser.Username)
	require.Equal(t, arg.FullName.String, updatedUser.FullName)
	require.Equal(t, arg.Email.String, updatedUser.Email)
	require.Equal(t, arg.HashedPassword.String, updatedUser.HashedPassword)
	require.WithinDuration(t, oldUser.CreatedAt.Time, updatedUser.CreatedAt.Time, time.Second)
	require.WithinDuration(t, oldUser.PasswordChangedAt.Time, updatedUser.PasswordChangedAt.Time, time.Second)
}
