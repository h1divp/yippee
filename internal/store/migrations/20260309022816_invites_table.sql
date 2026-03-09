-- +goose Up
-- +goose StatementBegin
CREATE TABLE invites (
    id INTEGER PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    created_by INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    used_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
    used_at DATETIME,
    expires_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS invites;
-- +goose StatementEnd
