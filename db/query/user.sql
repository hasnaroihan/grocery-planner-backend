-- name: GetUser :one
SELECT * from users
WHERE id = $1 LIMIT 1;

-- name: ListUsers :many
SELECT * from users
ORDER BY username
LIMIT $1
OFFSET $2;

-- name: CreateUser :one
INSERT INTO users (
    username,
    email,
    password,
    role
) VALUES (
  $1, $2, $3, $4
)
RETURNING *;

-- name: UpdateUser :one
UPDATE users
  set username = $2,
  email = $3
WHERE id = $1
RETURNING *;

-- name: UpdatePassword :exec
UPDATE users
  set password = $2
WHERE email = $1;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;