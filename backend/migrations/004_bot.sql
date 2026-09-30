CREATE TABLE bot_users (
    user_id    TEXT PRIMARY KEY,
    first_name TEXT NOT NULL DEFAULT '',
    started_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE tasks ADD COLUMN reminded_at TIMESTAMPTZ;
ALTER TABLE tasks ADD COLUMN remind_after TIMESTAMPTZ;
