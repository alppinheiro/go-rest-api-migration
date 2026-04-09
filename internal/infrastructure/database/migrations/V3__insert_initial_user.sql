-- V3__insert_initial_user.sql
-- Insert an initial user for testing/bootstrapping

INSERT INTO users (id, name, created_at, email)
VALUES ('00000000-0000-0000-0000-000000000001', 'Initial User', NOW(), 'initial@example.com');
