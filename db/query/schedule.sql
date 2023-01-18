-- name: GetSchedule :one
SELECT * from schedules
WHERE id = $1 LIMIT 1;

-- name: ListSchedule :many
SELECT * from schedules
WHERE author = $1
ORDER BY created_at
LIMIT $2
OFFSET $3;

-- name: GetScheduleRecipe :many
SELECT * from schedules_recipes
WHERE schedule_id = $1;

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