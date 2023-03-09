-- name: GetRecipe :one
SELECT * from recipes
WHERE id = $1 LIMIT 1;

-- name: GetRecipeIngredients :many
SELECT ri.recipe_id, ri.ingredient_id, i.name, ri. amount, ri.unit_id
from recipes_ingredients as ri
INNER JOIN ingredients as i
ON ri.ingredient_id = i.id
WHERE recipe_id = $1
FOR SHARE;

-- name: ListRecipes :many
SELECT * from recipes
ORDER BY modified_at
LIMIT $1
OFFSET $2;

-- name: ListRecipesUser :many
SELECT * from recipes
WHERE author = $1
ORDER BY modified_at
LIMIT $2
OFFSET $3;

-- name: SearchRecipe :many
SELECT id, name, author, modified_at from recipes
WHERE name LIKE $1
LIMIT $2
OFFSET $3;

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

-- name: UpdateRecipe :one
UPDATE recipes
    set name = $2,
    portion = $3,
    steps = $4,
    modified_at = (now() at time zone 'utc')
WHERE id = $1
RETURNING *;

-- name: UpdateRecipeIngredient :one
UPDATE recipes_ingredients
    set amount = $3,
    unit_id = $4
WHERE recipe_id = $1 AND ingredient_id = $2
RETURNING *;

-- name: DeleteRecipe :exec
DELETE FROM recipes
WHERE id = $1;

-- name: DeleteRecipeIngredient :exec
DELETE FROM recipes_ingredients
WHERE recipe_id = $1 AND ingredient_id = $2;