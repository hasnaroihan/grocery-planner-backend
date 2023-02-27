package auth

import "time"

type TokenMaker interface {
	CreateToken(username string, duration time.Duration, audiences []string) (string, error)
	VerifyToken(token string) (*Payload, error)
}