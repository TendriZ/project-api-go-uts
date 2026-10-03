CREATE TABLE IF NOT EXISTS users (
    id         SERIAL PRIMARY KEY,
    username   VARCHAR(30)  NOT NULL,
    email      VARCHAR(120) NOT NULL,
    password   VARCHAR(255) NOT NULL,
    is_active  BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS users_username_key ON users (LOWER(username));
CREATE UNIQUE INDEX IF NOT EXISTS users_email_key    ON users (LOWER(email));
