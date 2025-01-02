package api

import (
	"database/sql"
	"errors"
	"net/http"

	db "github.com/Oliver-Zen/simplebank/db/sqlc"
	"github.com/Oliver-Zen/simplebank/token"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

type createAccountRequest struct {
	Currency string `json:"currency" binding:"required,currency"` // be careful of usage (no sapce!)

	// A User should only be able to create Account for him/herself
	// Owner string `json:"owner" binding:"required"` // client input validation
	// Currency string `json:"currency" binding:"required,oneof=USD EUR CAD"` // be careful of usage (no sapce!)
}

// WHY `ctx`? In Gin, every HandlerFunc has `*Context` as input.
// Authorization Rule for Create Account API: A logged-in user can only create an account for him/herself.
func (server *Server) createAccount(ctx *gin.Context) {
	var req createAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil { // bad request
		ctx.JSON(http.StatusBadRequest, errorResponse(err)) // send response
		return
	}

	// get the username stored in header
	// MustGet return a general interface (any, aka. interface{})
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	arg := db.CreateAccountParams{
		// Owner:    req.Owner,
		Owner:    authPayload.Username,
		Currency: req.Currency,
		Balance:  0,
	}
	account, err := server.store.CreateAccount(ctx, arg)
	if err != nil { // internal issue (req validated already)
		// handle DB error(s)
		if pqErr, ok := err.(*pq.Error); ok {
			// `err.(*pq.Error)`: Attempts to convert `err` (an error interface) into a *pq.Error.
			// `ok`: A boolean indicating whether the type assertion was successful.
			// log.Println(pqErr.Code.Name())
			switch pqErr.Code.Name() {
			case "foreign_key_violation", "unique_violation":
				// WHAT is 403 Forbidden? Server understands the request but refuses to authorize it
				ctx.JSON(http.StatusForbidden, errorResponse(err)) // send response
				return
			}
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err)) // send response
		return
	}
	ctx.JSON(http.StatusOK, account)
}

// WHAT is Binding?
// The process of automatically mapping incoming HTTP request data (e.g., JSON, query params) to a Go struct.
// Authorization Rule for Get Account API: A logged-in user can only get accounts that he/she owns.
type getAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

// WHAT URI param? param embedded directly in the path.
func (server *Server) getAccount(ctx *gin.Context) {
	var req getAccountRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	account, err := server.store.GetAccount(ctx, req.ID)
	if err != nil { // 2 possible scenarios:
		// 1) `ID` doesn't exist
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		// 2) internal error
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// account := db.Account{} // to understand mockDB: `Times(1)`
	// account = db.Account{}  // to understand mockDB: `requireBodyMatchAccount`

	// authorization
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if account.Owner != authPayload.Username {
		err := errors.New("account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, account)
}

type listAccountRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

// WHAT Pagination? Divide the records into multiple pages of small size; achieve this by [query params].
// WHY called [query]? Because the question mark in URL.
// [query param] === [URL param]
// Authorization Rule for List Account API: A logged-in user can only list accounts that he/she owns.
func (server *Server) listAccount(ctx *gin.Context) {
	var req listAccountRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		// ctx.JSON(http.StatusBadRequest, err) // WHY incorrect?
		// because `ctx.JSON` serializes the `err` object directly -> empty JSON object {}
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// authorization
	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	arg := db.ListAccountsParams{
		Owner:  authPayload.Username,
		Limit:  req.PageSize, // WHAT `Limist`? 限制返回的【总条数】
		Offset: (req.PageID - 1) * req.PageSize,
	}

	accounts, err := server.store.ListAccounts(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
	}
	ctx.JSON(http.StatusOK, accounts)
}
