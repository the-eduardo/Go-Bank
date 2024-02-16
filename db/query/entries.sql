-- noinspection SqlResolveForFile
-- NewEntry Does not add the amount of money. Use AddAccountBalance instead
-- name: NewEntry :one
INSERT INTO entries (
account_id,
amount
) VALUES (
 $1, $2
) RETURNING *;

-- GetEntry returns the entry with an entry ID
-- name: GetEntry :one
SELECT * FROM entries
WHERE id = $1;

-- ListEntries returns a list of entries for the given account ID
-- name: ListEntries :many
SELECT * FROM entries
WHERE account_id = $1
ORDER BY id
LIMIT $2
OFFSET $3;