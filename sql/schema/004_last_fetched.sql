-- +goose Up
-- Add a nullable timestamp column to track when a feed was last fetched
ALTER TABLE feeds ADD COLUMN last_fetched_at TIMESTAMP NULL;

-- +goose Down
-- Remove the column if rolling back
ALTER TABLE feeds DROP COLUMN last_fetched_at;