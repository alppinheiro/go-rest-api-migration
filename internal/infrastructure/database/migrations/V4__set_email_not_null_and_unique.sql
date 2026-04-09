-- V4__set_email_not_null_and_unique.sql
-- Ensure email is NOT NULL and unique. For existing rows with NULL email,
-- populate them with a unique placeholder derived from the user's id.

-- Populate NULL emails with a unique placeholder using the user's id
UPDATE users SET email = (id::text || '@placeholder.invalid') WHERE email IS NULL;

-- Set column NOT NULL
ALTER TABLE users ALTER COLUMN email SET NOT NULL;

-- Create unique index on email
CREATE UNIQUE INDEX IF NOT EXISTS ux_users_email ON users(email);
