package api

import (
	"database/sql"
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

// to avoid sending sensitive data (e.g., hashedPassword, access_token, etc.) back to client
type userResponse struct {
	Username          string    `json:"username"` // `<tag_content>`: tag for struct field, providing metadata that can be used by libraries
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

func newUserResponse(user db.User) userResponse {
	return userResponse{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}
}

// `createUser` handles HTTP requests for creating a new user in the RESTful API.
func (server *Server) createUser(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil { // bad request
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	// HOW `CreateUser(gomock.Any(), gomock.Any())` weaken the test? (2)
	// hashedPassword, err = util.HashPassword("xyz") // an invalid password
	// expect fail (but pass before using custom gomock matcher)
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

	// HOW `CreateUser(gomock.Any(), gomock.Any())` weaken the test? (1)
	// arg = db.CreateUserParams{} // expect fail (but pass before using custom gomock matcher)

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

	res := newUserResponse(user)

	ctx.JSON(http.StatusOK, res)
}

type loginUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginUserResponse struct {
	AccessToken string       `json:"access_token" binding:"required,alphanum"`
	User        userResponse `json:"user" binding:"required,min=6"`
}

func (server *Server) loginUser(ctx *gin.Context) {
	var req loginUserRequest
	// WHY cannot ctx.ShouldBindJSON(req)?
	// see docs
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// `ShouldBindJSON` populates the `Username` and `Email` fields in `req` based on the JSON keys and values.
		ctx.JSON(http.StatusBadRequest, errorResponse(err)) 
		return
	}

	user, err := server.store.GetUser(ctx, req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// check password
	err = util.CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	accessToken, err := server.tokenMaker.CreateToken(
		user.Username,
		server.config.AccessTokenDuration,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	res := loginUserResponse{
		AccessToken: accessToken,
		User:        newUserResponse(user),
	}
	ctx.JSON(http.StatusOK, res)
	// return // redundent return statement
}
