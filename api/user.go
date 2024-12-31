package api

import (
	"net/http"
	"time"

	db "github.com/Oliver-Zen/simplebank/db/sqlc"
	"github.com/Oliver-Zen/simplebank/util"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type createUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

// to avoid sending `hashedPassword` back to client
type createUserResponse struct {
	Username          string    `json:"username"` // `<tag_content>`: tag for struct field, providing metadata that can be used by libraries
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

// `createUser` handles HTTP requests for creating a new user in the RESTful API.
func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil { // bad request
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err)) // internal error
		return
	}

	arg := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashedPassword,
		FullName:       req.FullName,
		Email:          req.Email,
	}
	user, err := server.store.CreateUser(ctx, arg)
	if err != nil { // internal issue (req validated already)
		// handle DB error(s)
		if pqErr, ok := err.(*pq.Error); ok {
			// `err.(*pq.Error)`: Attempts to convert `err` (an error interface) into a *pq.Error.
			// `ok`: A boolean indicating whether the type assertion was successful.
			switch pqErr.Code.Name() {
			case "unique_violation": // check table structure to determine constraint(s)
				// WHAT is 403 Forbidden? Server understands the request but refuses to authorize it
				ctx.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err)) // internal error
		return
	}

	res := &createUserResponse{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}

	ctx.JSON(http.StatusOK, res)
}
