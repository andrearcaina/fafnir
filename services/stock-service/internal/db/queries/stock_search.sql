-- name: GetStockMetadataBySymbol :one
SELECT * FROM stock_metadata WHERE symbol = $1;

-- name: SearchStockMetadataByName :many
SELECT * FROM stock_metadata WHERE name ILIKE '%' || $1 || '%' LIMIT $2 OFFSET $3;