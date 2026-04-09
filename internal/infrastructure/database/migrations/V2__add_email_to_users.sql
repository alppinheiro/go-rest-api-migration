-- V2__add_email_to_users.sql
-- Add an email column to users. Keep nullable to avoid migration issues.

ALTER TABLE users
ADD COLUMN IF NOT EXISTS email TEXT;
