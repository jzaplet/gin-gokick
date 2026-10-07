-- name: RecordFailure :exec
INSERT INTO auth_failures (scope, network) VALUES (@scope, @network);

-- name: CountFailures :one
SELECT count(*) FROM auth_failures WHERE scope = @scope AND network = @network AND failed_at > now() - @period::interval;

-- name: DeleteOldFailures :exec
DELETE FROM auth_failures WHERE scope = @scope AND failed_at <= now() - @period::interval;
