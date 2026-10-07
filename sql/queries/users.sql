-- name: CreateUser :one
INSERT INTO users(id, created_at, updated_at, email, hashed_password)
VALUES(
  $1,
  $2,
  $3,
  $4,
  $5
) 
RETURNING *;

-- name: DeleteUsers :exec
DELETE FROM users;

-- name: CheckUser :one
SELECT * FROM users WHERE email = $1;

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES(
  $1,
  $2,
  $3,
  $4,
  $5,
  $6
)RETURNING *;

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens WHERE token = $1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens 
SET revoked_at = $1, updated_at = $2
WHERE token = $3;

-- name: UpdateUserData :one
UPDATE users 
SET hashed_password = $1, updated_at = $2, email = $3
WHERE id = $4
RETURNING *;

-- name: GetUserData :one
SELECT * FROM users WHERE id = $1;

-- name: UserUpgradeRed :one
UPDATE users
SET is_chirpy_red = true, updated_at = $2
WHERE id = $1
RETURNING *;
