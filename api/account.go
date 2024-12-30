package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	db "github.com/techschool/simplebank/db/sqlc"
)

type createAccountRequest struct {
	// client input validation
	Owner    string `json:"owner" binding:"required"`
	Currency string `json:"currency" binding:"required,oneof=USD EUR CAD"` // be careful of usage (no sapce!)
}

// WHY `ctx`? In Gin, every HandlerFunc has `*Context` as input.
func (server *Server) createAccount(ctx *gin.Context) {
	var req createAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil { // bad request
		ctx.JSON(http.StatusBadRequest, errorResponse(err)) // send response
		return
	}

	arg := db.CreateAccountParams{
		Owner:    req.Owner,
		Currency: req.Currency,
		Balance:  0,
	}
	account, err := server.store.CreateAccount(ctx, arg)
	if err != nil { // internal issue (req validated already)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err)) // send response
		return
	}
	ctx.JSON(http.StatusOK, account)
}

// WHAT Binding? The process of automatically mapping incoming HTTP request data (e.g., JSON, query params) to a Go struct.
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

	ctx.JSON(http.StatusOK, account)
}

type listAccountRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

// WHAT Pagination? Divide the records into multiple pages of small size; achieve this by [query params].
// WHY called [query]? Because the question mark in URL.
// [query param] === [URL param]
func (server *Server) listAccount(ctx *gin.Context) {
	var req listAccountRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		// ctx.JSON(http.StatusBadRequest, err) // WHY incorrect?
		// because `ctx.JSON` serializes the `err` object directly -> empty JSON object {}
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.ListAccountsParams{
		Limit:  req.PageSize, // WHAT `Limist`? 限制返回的【总条数】
		Offset: (req.PageID - 1) * req.PageSize,
	}

	accounts, err := server.store.ListAccounts(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
	}
	ctx.JSON(http.StatusOK, accounts)
}
