package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hasnaroihan/grocery-planner/util"
	"github.com/stretchr/testify/require"
)

func TestPASETOMaker(t *testing.T) {
	PASETOMaker, err := NewPASETOToken(util.RandomString(32))
	require.NoError(t, err)

	subject,_ := uuid.NewRandom()
	duration := time.Minute
	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)
	audiences := []string{"http://localhost"}

	PASETOToken, err := PASETOMaker.CreateToken(subject, duration, audiences)
	require.NoError(t, err)
	require.NotEmpty(t, PASETOToken)

	payload, err := PASETOMaker.VerifyToken(PASETOToken)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	require.NotZero(t, payload)
	require.Equal(t, subject, payload.Subject)
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expiredAt, payload.ExpiredAt, time.Second)
	require.Equal(t, audiences, payload.Audience)
}

func TestExpiredPASETOToken(t *testing.T){
	pasetoMaker, err := NewPASETOToken(util.RandomString(32))
	require.NoError(t, err)

	subject, err := uuid.NewRandom()
	require.NoError(t, err)

	pasetoToken, err := pasetoMaker.CreateToken(subject, -2*time.Minute, []string{})
	require.NoError(t, err)
	require.NotEmpty(t, pasetoToken)

	payload, err := pasetoMaker.VerifyToken(pasetoToken)
	require.Error(t, err)
	require.Empty(t, payload)
}