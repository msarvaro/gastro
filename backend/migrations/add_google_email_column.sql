-- Migration: Add google_email column to users table for Google OAuth support

ALTER TABLE users ADD COLUMN IF NOT EXISTS google_email VARCHAR(255);

-- Create unique index on google_email for faster lookups and prevent duplicates
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_google_email ON users(google_email) WHERE google_email IS NOT NULL;

-- Add comment for documentation
COMMENT ON COLUMN users.google_email IS 'Google OAuth email address for authentication'; 