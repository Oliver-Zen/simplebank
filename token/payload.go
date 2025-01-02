package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Different types of error returned by the VerifyToken function
var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrExpiredToken = errors.New("token has expired")
)

// Payload contains the payload data of the token.
type Payload struct {
	// normally these 3 fields are enough
	Username  string    `json:"username"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`

	// add `ID` to invalidate some specific tokens in case they are leaked
	ID uuid.UUID `json:"id"`
}

// GetAudience implements jwt.Claims.
func (payload *Payload) GetAudience() (jwt.ClaimStrings, error) {
	// panic("unimplemented")
	return nil, nil // No audience used; return nil
}

// GetExpirationTime implements jwt.Claims.
func (payload *Payload) GetExpirationTime() (*jwt.NumericDate, error) {
	// panic("unimplemented")
	return jwt.NewNumericDate(payload.ExpiredAt), nil
}

// GetIssuedAt implements jwt.Claims.
func (payload *Payload) GetIssuedAt() (*jwt.NumericDate, error) {
	// panic("unimplemented")
	return jwt.NewNumericDate(payload.IssuedAt), nil
}

// GetIssuer implements jwt.Claims.
func (payload *Payload) GetIssuer() (string, error) {
	// panic("unimplemented")
	return "", nil // No issuer; return empty string
}

// GetNotBefore implements jwt.Claims.
func (payload *Payload) GetNotBefore() (*jwt.NumericDate, error) {
	// panic("unimplemented")
	return nil, nil // NotBefore not used; return nil
}

// GetSubject implements jwt.Claims.
func (payload *Payload) GetSubject() (string, error) {
	// panic("unimplemented")
	return "", nil // No subject; return empty string
}

// NewPayload creates a new token payload with a specific username and duration.
func NewPayload(username string, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	payload := &Payload{
		ID:        tokenID,
		Username:  username,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}
	return payload, err
}

// Valid checks if the token payload is valid or not.
// Specficially, it checks whether the token is expeired or not.
func (payload *Payload) Valid() error {
	if time.Now().After(payload.ExpiredAt) {
		return ErrExpiredToken
		// return jwt.ErrTokenExpired
	}
	return nil
}
