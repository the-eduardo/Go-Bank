package api

import (
	"database/sql"
	"github.com/gin-gonic/gin"
	db "github.com/the-eduardo/Go-Bank/db/sqlc"
	"net/http"
)

//////// Create Entry Request using existing
type createEntryRequest struct {
	AccountID int64 `json:"account_id" binding:"required,min=1"`
	Amount    int64 `json:"amount" binding:"required,gt=1"`
}

func (server *Server) createEntry(ctx *gin.Context) {
	var req createEntryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	arg := db.CreateEntriesParams{
		AccountID: req.AccountID,
		Amount:    req.Amount,
	}
	entry, err := server.store.CreateEntries(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, entry)
}

////////////////// Get Entry request using EntryID

type getEntry struct {
	EntryID int64 `uri:"id" binding:"required,min=1"`
}

func (server *Server) getEntry(ctx *gin.Context) {
	var req getEntry
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	entry, err := server.store.GetEntries(ctx, req.EntryID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, entry)
}

////////////// List existing entries for an account

type listEntryRequest struct {
	AccountID int64 `uri:"account_id" binding:"required,min=1"`
	PageID    int32 `form:"page_id" binding:"required,min=1"`
	PageSize  int32 `form:"page_size" binding:"required,min=5,max=20"`
}

func (server *Server) listEntries(ctx *gin.Context) {
	var req listEntryRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	arg := db.ListEntriesParams{
		Limit:     req.PageSize,
		Offset:    (req.PageID - 1) * req.PageSize,
		AccountID: req.AccountID,
	}

	entries, err := server.store.ListEntries(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, entries)
}
