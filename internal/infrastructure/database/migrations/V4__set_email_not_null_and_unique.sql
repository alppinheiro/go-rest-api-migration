UPDATE users SET email = (id::text || '@placeholder.invalid') WHERE email IS NULL;

ALTER TABLE users ALTER COLUMN email SET NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS ux_users_email ON users(email);
