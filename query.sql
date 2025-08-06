-- USER --

-- InsertUser inserts a new user
-- name: InsertUser :one
INSERT INTO users (username, email, password_hash)
VALUES ($1, $2, $3)
RETURNING *;

-- SelectUserWhereId finds a user by id
-- name: SelectUserWhereId :one
SELECT * FROM users WHERE id = $1 LIMIT 1;

-- SelectUserWhereEmail finds a user by email
-- name: SelectUserWhereEmail :one
SELECT * FROM users WHERE email = $1 LIMIT 1;

-- SelectUserWhereUsername finds a user by username
-- name: SelectUserWhereUsername :one
SELECT * FROM users WHERE username = $1 LIMIT 1;

-- REFRESH TOKEN --

-- InsertRefreshToken inserts a new refresh token
-- name: InsertRefreshToken :one
INSERT INTO refresh_tokens (user_id, issued_at, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- SelectRefreshTokenByUUID finds a refresh token by uuid
-- name: SelectRefreshTokenByUUID :one
SELECT * FROM refresh_tokens WHERE uuid = $1 LIMIT 1;

-- DeleteRefreshTokenWhereUUID deletes a refresh token by uuid
-- name: DeleteRefreshTokenWhereUUID :exec
DELETE FROM refresh_tokens WHERE uuid = $1;

-- DeleteRefreshTokensWhereExpired deletes al refresh tokens that have expired
-- name: DeleteRefreshTokensWhereExpired :one
WITH deleted AS (
  DELETE FROM refresh_tokens
  WHERE NOW() > expires_at
  RETURNING *
)
SELECT COUNT(*) AS deleted_count
FROM deleted;

-- ROLE --

-- InsertRole creates a new role
-- name: InsertRole :one
INSERT INTO roles (name) VALUES ($1) RETURNING *;

-- SelectRoleWhereName finds a role by its name
-- name: SelectRoleWhereName :one
SELECT * FROM roles WHERE name = $1 LIMIT 1;

-- SelectRolesWhereUserID finds all roles that a user has through user_roles
-- name: SelectRolesWhereUserID :many
SELECT r.*
FROM users u
JOIN user_roles ur ON ur.user_id = u.id
JOIN roles r       ON r.id = ur.role_id
WHERE u.id = $1;

-- USER ROLE --

-- InsertUserRole creates an association between user and roles
-- name: InsertUserRole :one
INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2) RETURNING *;
