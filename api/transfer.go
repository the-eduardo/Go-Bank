package api

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	db "github.com/the-eduardo/Go-Bank/db/sqlc"
	"net/http"
)

type CreateTransferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required,currency"`
}

func (server *Server) createTransfer(ctx *gin.Context) {
	var req CreateTransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Check if the accounts exist and if the currency matches
	if !accountValidator(server, ctx, req.FromAccountID, req.Currency, true) {
		//ctx.JSON(http.StatusNotFound, errorResponse(errors.New("from account not found")))
		return
	}
	if !accountValidator(server, ctx, req.ToAccountID, req.Currency, true) {
		//ctx.JSON(http.StatusNotFound, errorResponse(errors.New("to account not found")))
		return
	}

	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}
	transfer, err := server.store.TransferTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, transfer)
}

type getTransferRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getTransfer(ctx *gin.Context) {
	var req getTransferRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	transfer, err := server.store.GetTransferById(ctx, req.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, transfer)
}

type ListTransferRequest struct {
	FromAccountID int64 `form:"from_account_id" binding:"required,min=1"`
	PageID        int64 `form:"page_id" binding:"required,min=1"`
	PageSize      int64 `form:"page_size" binding:"required,min=5,max=10"`
}

func (server *Server) listTransfers(ctx *gin.Context) {
	var req ListTransferRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// Check if the accounts exist
	if !accountValidator(server, ctx, req.FromAccountID, "", false) {
		ctx.JSON(http.StatusNotFound, errorResponse(errors.New("account not found")))
		return
	}

	arg := db.ListTransfersByAccountIdParams{
		FromAccountID: req.FromAccountID,
		Limit:         req.PageSize,
		Offset:        (req.PageID - 1) * req.PageSize,
	}

	transfer, err := server.store.ListTransfersByAccountId(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, transfer)
}
