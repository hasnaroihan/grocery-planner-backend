package auth

import (
	"testing"
	"time"

	"github.com/hasnaroihan/grocery-planner/util"
	"github.com/stretchr/testify/require"
)

func TestPASETOMaker(t *testing.T) {
	PASETOMaker, err := NewPASETOToken(util.RandomString(32))
	require.NoError(t, err)

	username := util.RandomUsername()
	duration := time.Minute
	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)
	audiences := []string{"http://localhost"}

	PASETOToken, err := PASETOMaker.CreateToken(username, duration, audiences)
	require.NoError(t, err)
	require.NotEmpty(t, PASETOToken)

	payload, err := PASETOMaker.VerifyToken(PASETOToken)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	require.NotZero(t, payload)
	require.Equal(t, username, payload.Username)
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expiredAt, payload.ExpiredAt, time.Second)
	require.Equal(t, audiences, payload.Audience)
}