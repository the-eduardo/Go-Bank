package api

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	db "github.com/the-eduardo/Go-Bank/db/sqlc"
	"net/http"
)

//////// Create Transfer Request using existing TransferTxParams
type createTransferRequest struct {
	FromAccountID int64 `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64 `json:"to_account_id" binding:"required,min=1"`
	Amount        int64 `json:"amount" binding:"required,gt=0"`
}

func (server *Server) createTransfer(ctx *gin.Context) {
	var req createTransferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
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

////////////////// Get transfer request using TransferID

type getTransferRequest struct {
	TransferID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getTransfer(ctx *gin.Context) {
	var req getTransferRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	transfer, err := server.store.GetTransfer(ctx, req.TransferID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, transfer)
}

////////////// List existing transfers
// TODO: implement "list all transfers"

type listTransferRequest struct {
	PageID        int32 `form:"page_id" binding:"required,min=1"`
	PageSize      int32 `form:"page_size" binding:"required,min=5,max=20"`
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
}

func (server *Server) listTransfer(ctx *gin.Context) {
	var req listTransferRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	arg := db.ListTransferParams{
		Limit:         req.PageSize,
		Offset:        (req.PageID - 1) * req.PageSize,
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
	}

	transfer, err := server.store.ListTransfer(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, transfer)
}
