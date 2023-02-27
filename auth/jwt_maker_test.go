package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hasnaroihan/grocery-planner/util"
	"github.com/stretchr/testify/require"
)

func TestJWTMaker(t *testing.T) {
	jwtMaker, err := NewJWTToken(util.RandomString(32))
	require.NoError(t, err)
	
	username := util.RandomUsername()
	duration := time.Minute
	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)
	audiences := []string{"http://localhost"}

	jwtToken, err := jwtMaker.CreateToken(username, duration, audiences)
	require.NoError(t, err)
	require.NotEmpty(t, jwtToken)

	payload, err := jwtMaker.VerifyToken(jwtToken)
	require.NoError(t, err)
	require.NotEmpty(t, payload)

	require.NotZero(t, payload)
	require.Equal(t, username, payload.Username)
	require.WithinDuration(t, issuedAt, payload.IssuedAt, time.Second)
	require.WithinDuration(t, expiredAt, payload.ExpiredAt, time.Second)
	require.Equal(t, audiences, payload.Audience)
}

func TestExpiredJWTToken(t *testing.T){
	jwtMaker, err := NewJWTToken(util.RandomString(32))
	require.NoError(t, err)

	jwtToken, err := jwtMaker.CreateToken(util.RandomUsername(), -2*time.Minute, []string{})
	require.NoError(t, err)
	require.NotEmpty(t, jwtToken)

	payload, err := jwtMaker.VerifyToken(jwtToken)
	require.Error(t, err)
	require.Empty(t, payload)
}

func TestInvalidJWTTokenAlgNone(t *testing.T) {
	payload, err := NewPayload(util.RandomUsername(), time.Minute, []string{})
	require.NoError(t, err)

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodNone, payload)
	token, err := jwtToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	jwtMaker, err := NewJWTToken(util.RandomString(32))
	require.NoError(t, err)

	payload, err = jwtMaker.VerifyToken(token)
	require.Error(t, err)
	require.Empty(t, payload)
}