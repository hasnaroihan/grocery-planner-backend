-- name: GetSchedule :one
SELECT * from schedules
WHERE id = $1 LIMIT 1;

-- name: ListSchedules :many
SELECT * from schedules
ORDER BY created_at
LIMIT $1
OFFSET $2;

-- name: GetScheduleRecipe :many
SELECT sr.schedule_id, sr.recipe_id, r.name, sr.portion
from schedules_recipes as sr 
INNER JOIN recipes as r
ON sr.reciped_id = r.recipe_id
WHERE sr.schedule_id = $1;

-- name: CreateSchedule :one
INSERT INTO schedules (
    author
) VALUES (
    $1
) RETURNING *;

-- name: CreateScheduleRecipe :one
INSERT INTO schedules_recipes (
    schedule_id,
    recipe_id,
    portion
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: DeleteSchedule :exec
DELETE FROM schedules
WHERE id = $1;

-- name: DeleteScheduleRecipe :exec
DELETE FROM schedules_recipes
WHERE schedule_id = $1 AND recipe_id = $2;

-- name: ListGroceries :many
SELECT i.id, i.name
FROM schedules_recipes AS sr
INNER JOIN recipes_ingredients AS ri
ON sr.recipe_id = ri.recipe_id
INNER JOIN ingredients AS i
ON ri.ingredient_id = i.id
WHERE sr.schedule_id = $1
GROUP BY i.id
ORDER BY i.name;