-- name: CreateUnit :one
INSERT INTO units (
    name
) VALUES (
    $1
) RETURNING *;

-- name: ListUnit :many
SELECT * from units;

-- name: UpdateUnit :exec
UPDATE units
    set name = $2
WHERE id = $1;

-- name: DeleteUnit :exec
DELETE FROM units
WHERE id = $1;