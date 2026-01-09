-- name: GetWatchlist :many
SELECT symbol, created_at FROM watchlists
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: AddToWatchlist :exec
INSERT INTO watchlists (user_id, symbol)
VALUES ($1, $2)
ON CONFLICT (user_id, symbol) DO NOTHING;

-- name: RemoveFromWatchlist :exec
DELETE FROM watchlists
WHERE user_id = $1 AND symbol = $2;
