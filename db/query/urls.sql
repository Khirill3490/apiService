-- name: SaveURL :one
INSERT INTO urls (alias, url)
VALUES ($1, $2)
RETURNING id, alias;

-- name: GetURLByAlias :one
SELECT url
FROM urls
WHERE alias = $1;

-- name: DeleteURLByAlias :execrows
DELETE FROM urls
WHERE alias = $1;
