package auth

import (
	"time"

	"github.com/google/uuid"
)

type TokenMaker interface {
	CreateToken(subject uuid.UUID, duration time.Duration, audiences []string) (string, error)
	VerifyToken(token string) (*Payload, error)
}