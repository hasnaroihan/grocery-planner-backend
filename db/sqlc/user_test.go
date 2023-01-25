package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/hasnaroihan/grocery-planner/util"
	"github.com/stretchr/testify/require"
)

func CreateRandomUser(t *testing.T) User {
	arg := CreateUserParams{
		Username: util.RandomUsername(),
		Email: util.RandomEmail(),
		Password: "testpassword", // Real password will be hashed
		Role: util.RandomRole(),
	}
	user, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.Email, user.Email)
	require.Equal(t, arg.Password, user.Password)
	require.Equal(t, arg.Role, user.Role)

	require.NotZero(t, user.ID)
	require.NotZero(t, user.CreatedAt)

	return user
}
func TestCreateUser(t *testing.T) {
	arg := CreateUserParams{
		Username: util.RandomUsername(),
		Email: util.RandomEmail(),
		Password: "testpassword", // Real password will be hashed
		Role: util.RandomRole(),
	}

	user, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.Email, user.Email)
	require.Equal(t, arg.Password, user.Password)
	require.Equal(t, arg.Role, user.Role)

	require.NotZero(t, user.ID)
	require.NotZero(t, user.CreatedAt)
}

func TestGetUser(t *testing.T) {
	// Create User
	userNew := CreateRandomUser(t)
	user, err := testQueries.GetUser(
		context.Background(),
		userNew.ID,
	)

	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, userNew.ID, user.ID)
	require.Equal(t, userNew.Username, user.Username)
	require.Equal(t, userNew.Email, user.Email)
	require.Equal(t, userNew.Password, user.Password)
	require.Equal(t, userNew.Role, user.Role)
	require.WithinDuration(t, userNew.CreatedAt, user.CreatedAt, time.Second)

	require.Equal(t, userNew.VerifiedAt, user.VerifiedAt)
}

func TestListUsers(t *testing.T) {
	for i := 0; i < 10; i++ {
		CreateRandomUser(t)
	}

	arg := ListUsersParams {
		Limit: 5,
		Offset: 5,
	}
	users, err := testQueries.ListUsers(
		context.Background(),
		arg,
	)
	require.NoError(t, err)
	require.NotEmpty(t, users)
	require.Len(t, users, int(arg.Limit))

	for _,row := range users {
		require.NotEmpty(t, row)
	}
}

func TestUpdateUser(t *testing.T) {
	userNew := CreateRandomUser(t)

	arg := UpdateUserParams {
		ID: userNew.ID,
		Username: util.RandomUsername(),
		Email: userNew.Email,
		VerifiedAt: sql.NullTime{
			Time: time.Now(),
			Valid: true,
		},
	}

	user, err := testQueries.UpdateUser(
		context.Background(),
		arg,
	)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.ID, user.ID)
	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.Email, user.Email)
	require.WithinDuration(t, arg.VerifiedAt.Time, user.VerifiedAt.Time, time.Second)
}

func TestUpdatePassword(t *testing.T) {
	userNew := CreateRandomUser(t)

	arg := UpdatePasswordParams {
		Email: userNew.Email,
		Password: "newpassword",
	}

	user, err := testQueries.UpdatePassword(
		context.Background(),
		arg,
	)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Email, user.Email)
	require.Equal(t, arg.Password, user.Password)
}

func TestDeleteUser(t *testing.T) {
	userNew := CreateRandomUser(t)
	err := testQueries.DeleteUser(
		context.Background(),
		userNew.ID,
	)
	require.NoError(t, err)

	// Test read deleted user
	user, err := testQueries.GetUser(
		context.Background(),
		userNew.ID,
	)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, user)
}