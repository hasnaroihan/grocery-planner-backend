package auth

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Payload struct {
	ID			uuid.UUID	`json:"jti"`
	Username	string		`json:"sub"`
	IssuedAt	time.Time	`json:"iat"`
	ExpiredAt	time.Time	`json:"exp"`
	Audience	[]string	`json:"aud"`
	Issuer		string		`jsin:"iss"`
}

func NewPayload(username string, duration time.Duration, audiences []string) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	payload := &Payload{
		ID:			tokenID,
		Username: 	username,
		IssuedAt: 	time.Now(),
		ExpiredAt: 	time.Now().Add(duration),
		Audience:	audiences,
	}

	return payload, nil
}

func (p *Payload) GetNotBefore() (*jwt.NumericDate, error) {
	now := time.Now()
	if now.After(p.ExpiredAt) {
		return nil, jwt.ErrTokenExpired
	}
	return jwt.NewNumericDate(now), nil
}

func (p *Payload) GetAudience() (jwt.ClaimStrings, error) {
	return jwt.ClaimStrings(p.Audience), nil
}

func (p *Payload) GetExpirationTime() (*jwt.NumericDate, error) {
	if time.Time.IsZero(p.ExpiredAt) {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return jwt.NewNumericDate(p.ExpiredAt), nil
}

func (p *Payload) GetIssuedAt() (*jwt.NumericDate, error) {
	if time.Time.IsZero(p.IssuedAt) {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return jwt.NewNumericDate(p.IssuedAt), nil
}

func (p *Payload) GetIssuer() (string, error) {
	return p.Issuer, nil
}

func (p *Payload) GetSubject() (string, error) {
	return p.Username, nil
}