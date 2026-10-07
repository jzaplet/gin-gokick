-- name: CreateSession :one
INSERT INTO sessions (token_hash, user_id, expires_at, user_agent, last_seen_ip) VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: SessionByTokenHash :one
SELECT * FROM sessions WHERE token_hash = $1 AND expires_at > now();

-- name: TouchSession :exec
UPDATE sessions SET last_seen_at = now(), last_seen_ip = coalesce(sqlc.narg('last_seen_ip')::inet, last_seen_ip) WHERE id = @id;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE token_hash = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions WHERE expires_at <= now();

-- name: ListUserSessions :many
SELECT id, user_agent, created_at, last_seen_at, last_seen_ip, token_hash = @current_hash AS is_current
FROM sessions
WHERE user_id = @user_id AND expires_at > now() AND (@ip::text = '' OR strpos(host(last_seen_ip), lower(@ip::text)) > 0)
ORDER BY
    CASE WHEN @sort_by::text = 'createdAt' AND @descending::bool = false THEN created_at END,
    CASE WHEN @sort_by::text = 'createdAt' AND @descending::bool THEN created_at END DESC,
    CASE WHEN @sort_by::text = 'lastSeenAt' AND @descending::bool = false THEN last_seen_at END,
    CASE WHEN @sort_by::text = 'lastSeenAt' AND @descending::bool THEN last_seen_at END DESC,
    id
LIMIT @row_limit::bigint OFFSET @row_offset::bigint;

-- name: CountUserSessions :one
SELECT count(*) AS total, count(*) FILTER (WHERE token_hash <> @current_hash) AS others
FROM sessions
WHERE user_id = @user_id AND expires_at > now() AND (@ip::text = '' OR strpos(host(last_seen_ip), lower(@ip::text)) > 0);

-- name: EndUserSessions :execrows
DELETE FROM sessions
WHERE user_id = @user_id AND id = ANY(@ids::uuid[]) AND token_hash <> @current_hash AND expires_at > now();

-- name: EndOtherUserSessions :execrows
DELETE FROM sessions
WHERE user_id = @user_id AND token_hash <> @current_hash AND expires_at > now() AND (@ip::text = '' OR strpos(host(last_seen_ip), lower(@ip::text)) > 0);
