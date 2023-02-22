-- name: GetUser :one
SELECT * from users
WHERE id = $1 LIMIT 1;

-- name: GetLogin :one
SELECT id, username, password from users
WHERE username = $1 LIMIT 1
FOR SHARE;

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

-- name: UpdateVerified :one
UPDATE users
  set verified_at = $2
WHERE id = $1
RETURNING *;

-- name: UpdatePassword :one
UPDATE users
  set password = $2
WHERE email = $1
RETURNING email, password;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;