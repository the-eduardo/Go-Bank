// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0
// source: transfers.sql

package db

import (
	"context"
)

const createNewTransfer = `-- name: CreateNewTransfer :one
INSERT INTO transfers (
    from_account_id,
    to_account_id,
    amount
) VALUES (
    $1, $2, $3
) RETURNING id, from_account_id, to_account_id, amount, created_at
`

type CreateNewTransferParams struct {
	FromAccountID int64 `json:"from_account_id"`
	ToAccountID   int64 `json:"to_account_id"`
	Amount        int64 `json:"amount"`
}

// noinspection SqlResolveForFile
func (q *Queries) CreateNewTransfer(ctx context.Context, arg CreateNewTransferParams) (Transfer, error) {
	row := q.db.QueryRow(ctx, createNewTransfer, arg.FromAccountID, arg.ToAccountID, arg.Amount)
	var i Transfer
	err := row.Scan(
		&i.ID,
		&i.FromAccountID,
		&i.ToAccountID,
		&i.Amount,
		&i.CreatedAt,
	)
	return i, err
}

const getTransferById = `-- name: GetTransferById :one
SELECT id, from_account_id, to_account_id, amount, created_at FROM transfers
WHERE id = $1
`

// GetTransferById returns a single transfer by ID
func (q *Queries) GetTransferById(ctx context.Context, id int64) (Transfer, error) {
	row := q.db.QueryRow(ctx, getTransferById, id)
	var i Transfer
	err := row.Scan(
		&i.ID,
		&i.FromAccountID,
		&i.ToAccountID,
		&i.Amount,
		&i.CreatedAt,
	)
	return i, err
}

const listTransfersByAccountId = `-- name: ListTransfersByAccountId :many
SELECT id, from_account_id, to_account_id, amount, created_at FROM transfers
WHERE from_account_id = $1 OR to_account_id = $1
ORDER BY created_at DESC
LIMIT $2
OFFSET $3
`

type ListTransfersByAccountIdParams struct {
	FromAccountID int64 `json:"from_account_id"`
	Limit         int64 `json:"limit"`
	Offset        int64 `json:"offset"`
}

// ListTransfersByAccountId returns a list of transfers for a given account ID
func (q *Queries) ListTransfersByAccountId(ctx context.Context, arg ListTransfersByAccountIdParams) ([]Transfer, error) {
	rows, err := q.db.Query(ctx, listTransfersByAccountId, arg.FromAccountID, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Transfer{}
	for rows.Next() {
		var i Transfer
		if err := rows.Scan(
			&i.ID,
			&i.FromAccountID,
			&i.ToAccountID,
			&i.Amount,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
