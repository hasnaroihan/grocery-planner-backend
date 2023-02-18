-- name: CreateUnit :one
INSERT INTO units (
    name
) VALUES (
    $1
) RETURNING *;

-- name: GetUnit :one
SELECT * from units
WHERE id = $1;

-- name: ListUnits :many
SELECT * from units;

-- name: UpdateUnit :one
UPDATE units
    set name = $2
WHERE id = $1
RETURNING *;

-- name: DeleteUnit :exec
DELETE FROM units
WHERE id = $1;