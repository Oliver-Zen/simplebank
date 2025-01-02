package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	db "github.com/Oliver-Zen/simplebank/db/sqlc"
	"github.com/Oliver-Zen/simplebank/token"
	"github.com/gin-gonic/gin"
)

type transferRequest struct {
	FromAccountID int64 `json:"from_account_id" binding:"required,min=1"` // client input validation
	ToAccountID   int64 `json:"to_account_id" binding:"required,min=1"`
	Amount        int64 `json:"amount" binding:"required,gt=0"`
	// Currency      string `json:"currency" binding:"required,oneof=USD EUR CAD"` // be careful of usage (no sapce!)
	Currency string `json:"currency" binding:"required,currency"` // be careful of usage (no sapce!)
}

// WHY `ctx`? In Gin, every HandlerFunc has `*Context` as input.
// Authorization Rule for Transfer Money API: A logged-in user can only send money from his/her own account
func (server *Server) createTransfer(ctx *gin.Context) {
	var req transferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil { // bad request
		ctx.JSON(http.StatusBadRequest, errorResponse(err)) // send response
		return
	}

	fromAccount, valid := server.validAccount(ctx, req.FromAccountID, req.Currency)
	if !valid {
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if fromAccount.Owner != authPayload.Username {
		err := errors.New("from_account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	_, valid = server.validAccount(ctx, req.ToAccountID, req.Currency)
	if !valid {
		return
	}

	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}
	result, err := server.store.TransferTx(ctx, arg)
	if err != nil { // internal issue (req validated already)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, result)
}

// `validAccount` is a custom params validator.
// `validAccount` checks if an account with an specific ID exists and its currency matches the input currency.
func (server *Server) validAccount(ctx *gin.Context, accountID int64, currency string) (db.Account, bool) {
	account, err := server.store.GetAccount(ctx, accountID)
	if err != nil {
		if err == sql.ErrNoRows { // `ID` doesn't exist
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return account, false
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err)) // internal error
		return account, false
	}

	if account.Currency != currency { // currency mistach, bad request
		err := fmt.Errorf("account [%d] currency mismatch: %s vs %s", accountID, account.Currency, currency)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return account, false
	}
	return account, true
}
