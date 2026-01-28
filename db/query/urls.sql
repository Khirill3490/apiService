-- name: SaveURL :one
INSERT INTO urls (alias, url)
VALUES ($1, $2)
RETURNING id, alias;
