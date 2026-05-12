-- +goose Up
-- +goose StatementBegin
ALTER TABLE sessions ADD COLUMN shared INTEGER NOT NULL DEFAULT 0 CHECK (shared IN (0, 1));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE sessions DROP COLUMN shared;
-- +goose StatementEnd
