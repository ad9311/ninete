-- +goose Up
CREATE TABLE refresh_tokens (
  id INTEGER PRIMARY KEY NOT NULL,
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  uuid BLOB NOT NULL,
  issued_at INTEGER NOT NULL,
  expires_at INTEGER NOT NULL,
  revoked INTEGER NOT NULL DEFAULT 0
);

PRAGMA user_version = 2;

-- +goose Down
DROP TABLE IF EXISTS refresh_tokens;

PRAGMA user_version = 1;
