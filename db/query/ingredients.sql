-- name: CreateIngredient :one
INSERT INTO ingredients (
    name, default_unit
) VALUES (
    $1, $2
) RETURNING *;

-- name: GetIngredient :one
SELECT * from ingredients
WHERE id = $1;

-- name: ListIngredient :many
SELECT * from ingredients
ORDER BY name;

-- name: SearchIngredient :many
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

-- name: ListGrocery :many
SELECT DISTINCT ingredients.id, ingredients.name
FROM ingredients
INNER JOIN recipes_ingredients
ON ingredients.id = recipes_ingedients.ingredient_id
INNER JOIN schedules_recipes
ON recipes_ingredients.recipe_id = schedules_recipes.recipe_id
WHERE schedules_recipes.schedule_id = $1
ORDER BY ingredients.name;