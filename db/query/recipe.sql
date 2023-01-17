-- name: GetRecipe :one
SELECT * from recipes
WHERE id = $1 LIMIT 1;

-- name: GetRecipeIngredients :many
SELECT * from recipes_ingredients
WHERE recipe_id = $1;

-- name: ListRecipes :many
SELECT * from recipes
ORDER BY modified_at
LIMIT $1
OFFSET $2;

-- name: CreateRecipe :one
INSERT INTO recipes (
    name,
    author,
    portion,
    steps
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: CreateRecipeIngredient :one
INSERT INTO recipes_ingredients (
    ingredient_id,
    recipe_id,
    amount,
    unit_id
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: UpdateRecipe :exec
UPDATE recipes
    set name = $2,
    portion = $3,
    steps = $4,
    modified_at = $5
WHERE id = $1;

-- name: UpdateRecipeIngredient :exec
UPDATE recipes_ingredients
    set amount = $3,
    unit_id = $4
WHERE recipe_id = $1 AND ingredient_id = $2;

-- name: DeleteRecipe :exec
DELETE FROM recipes
WHERE id = $1;

-- name: DeleteRecipeIngredient :exec
DELETE FROM recipes_ingredients
WHERE recipe_id = $1 AND ingredient_id = $2;