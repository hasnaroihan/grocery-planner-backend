-- name: CreateIngredient :one
INSERT INTO ingredients (
    name, default_unit
) VALUES (
    $1, $2
) RETURNING *;

-- name: GetIngredient :one
SELECT * from ingredients
WHERE id = $1;

-- name: ListIngredients :many
SELECT * from ingredients
ORDER BY name;

-- name: SearchIngredients :many
SELECT id, name, default_unit from ingredients
WHERE name LIKE $1;

-- name: UpdateIngredient :one
UPDATE ingredients
    set name = $2,
    default_unit = $3
WHERE id = $1
RETURNING *;

-- name: DeleteIngredient :exec
DELETE FROM ingredients
WHERE id = $1;