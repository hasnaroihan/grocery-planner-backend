package auth

import (
	"fmt"
	"time"

	"github.com/aead/chacha20poly1305"
	"github.com/google/uuid"
	"github.com/o1egl/paseto"
)

type PASETOMaker struct {
	paseto *paseto.V2
	symmetricKey []byte
}

func NewPASETOToken(symmetricKey string) (TokenMaker, error) {
	if len(symmetricKey) < chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size: must be equal or more to %d characters", chacha20poly1305.KeySize)
	}

	return &PASETOMaker{paseto: paseto.NewV2(), symmetricKey: []byte(symmetricKey)}, nil
}

func (p *PASETOMaker) CreateToken(subject uuid.UUID, duration time.Duration, audiences []string) (string, error) {
	payload, err := NewPayload(subject, duration, audiences)
	if err != nil {
		return "", err
	}

	PASETOToken, err := p.paseto.Encrypt(p.symmetricKey, payload, nil)
	if err != nil {
		return "", err
	}
	return PASETOToken, nil
}

func (p *PASETOMaker) VerifyToken(token string) (*Payload, error) {
	payload := &Payload{}
	err := p.paseto.Decrypt(token, p.symmetricKey, payload, nil)
	if err != nil {
		return nil, err
	}

	err = payload.Valid()
	if err != nil {
		return nil, err
	}

	return payload, nil
}