package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/Oliver-Zen/simplebank/token"
	"github.com/gin-gonic/gin"
)

// extract the authorization header from the reqest
const (
	authorizationHeaderKey = "authorization"
	authorizationTypeBearer = "bearer"
	authorizationPayloadKey = "authorization_payload"
)

// authMiddleware validates the Authorization header in the HTTP request.
// It ensures the header exists, has a valid "Bearer" token format, and verifies the token.
// If the token is valid, the payload is stored in the context for downstream handlers.
// Otherwise, it aborts the request with a 401 Unauthorized status.
func authMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	// return an anonymous middleware func
	return func(ctx *gin.Context) {
		
		// Extract Authorization Header, 2 stpes
		// step (1): check if [Authorization] header exists
		authorizationHeader := ctx.GetHeader(authorizationHeaderKey)
		if len(authorizationHeader) == 0 {
			// client doesn't provide this header
			err := errors.New("authorization header is not provided")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err)) // 401 unauthorized status
			return
		}

		// step (2): check if [Authorization] header has a valid format
		// split the authorization header by space
		// header example: "Bearer: v2.<--token-->"
		// neccessay because server may support many authorization shcemas 
		// e.g., Bearer Token, OAuth, AWS signature, etc.
		fields := strings.Fields(authorizationHeader)
		if len(fields) < 2 {
			err := errors.New("invalid authorization header format")
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err)) // 401 unauthorized status
			return
		}
		
		// Validate Authorization Type
		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authorizationTypeBearer {
			err := fmt.Errorf("unsupported authorization type %s: ", authorizationType)
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err)) // 401 unauthorized status
			return
		}
		
		// parse & verfiy the token
		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken) // VerifyToken returns a pointer to token.Payload
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err)) // 401 unauthorized status
			return
		}
		
		// store `payload` in the context, before passing it to next handler
		ctx.Set(authorizationPayloadKey, payload)
		
		// forward the request to the next handler
		ctx.Next()
	}
}
