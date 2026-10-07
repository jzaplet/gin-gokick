-- name: CreateUser :one
INSERT INTO users (email, password_hash, locale) VALUES (lower(@email), @password_hash, @locale) RETURNING *;

-- name: UserByEmail :one
SELECT * FROM users WHERE email = lower(@email);

-- name: UpdatePasswordHash :exec
UPDATE users SET password_hash = @password_hash WHERE id = @id;

-- name: DeleteUser :exec
DELETE FROM users WHERE id = @id;

-- name: EmailByID :one
SELECT email FROM users WHERE id = @id;

-- name: UpdateLocale :exec
UPDATE users SET locale = @locale WHERE id = @id;
