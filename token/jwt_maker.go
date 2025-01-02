package token

import (
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const minSecretKeysize = 32

// JWTMaker is a JSON Web Token maker. It implements the `Maker` interface.
type JWTMaker struct {
	secretKey string // use symmetric algorithm
}

// NewJWTMaker creates a JWTMaker
func NewJWTMaker(secretKey string) (Maker, error) {
	if len(secretKey) < minSecretKeysize {
		return nil, fmt.Errorf("invalid key size: must be at least %d characters", minSecretKeysize)
	}
	return &JWTMaker{secretKey}, nil
}

// CreateToken creates a new token for a specific username and duration
func (maker *JWTMaker) CreateToken(username string, duration time.Duration) (string, error) {
	//  JWT = Header + Payload (Claim) + Signature

	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, payload) // `HS256` -> Header
	return jwtToken.SignedString([]byte(maker.secretKey))          // -> Signature
	// WHY convert `maker.secretKey` to byte slice?
	// Because signing process requires a byte slice as input for cryptographic algorithms.
}

// VerifyToken checks if the token is valid or not
func (maker *JWTMaker) VerifyToken(token string) (*Payload, error) {
	// `keyFunc` retrieves the key for verifying the token's signature.
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		// Check if the signing method is valid (i.e, whether the algorithm matches or not)
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		// HMAC: a generic method supporting all HMAC algorithms (HS256, HS384, HS512, etc.) for validation.
		// Verification can use SigningMethodHMAC to allow flexibility, but signing must use a specific method like HS256.
		if !ok {
			return nil, ErrInvalidToken
			// return nil, jwt.ErrTokenInvalidClaims
		}
		return []byte(maker.secretKey), nil
	}

	// Parse and validate the token
	// WHY `&Payload{}` instead of `Payload`?
	// `&Payload{}` passes a pointer.
	// `Payload` passes a value copy.
	// `ParseWithClaims` writes the parsed token data into the provided <jwt.Cliams>.
	jwtToken, err := jwt.ParseWithClaims(token, &Payload{}, keyFunc)
	if err != nil {
		// 2 possible scenarios: (1) invalid token & (2) expired
		// to figure out which error
		// Unwrap the original error

		// FIXME:
		// fmt.Printf("Error: %v\n", err)
		// fmt.Printf("Cause: %v\n", errors.Unwrap(err)) // Output: Cause: <nil>
		// WHY nil? (ChatGPT:) Because newError (<- ParseWithClaims)
		// Cannot wrap: current `newError` implementation
		// Can wrap: `wrapped = fmt.Errorf("%w: %w", wrapped, e)`

		// Match error message directly for expiration
		if strings.Contains(err.Error(), jwt.ErrTokenExpired.Error()) {
			return nil, ErrExpiredToken
			// return nil, jwt.ErrTokenExpired
		}
		return nil, ErrInvalidToken
		// return nil, jwt.ErrTokenInvalidClaims
	}

	payload, ok := jwtToken.Claims.(*Payload)
	if !ok {
		return nil, ErrInvalidToken
		// return nil, jwt.ErrTokenInvalidClaims
	}

	// no errors so far, token is valid
	return payload, nil
}
