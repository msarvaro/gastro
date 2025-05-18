-- Add description column to dishes table
ALTER TABLE dishes ADD COLUMN description TEXT DEFAULT '';

-- Add indexes to improve query performance
CREATE INDEX IF NOT EXISTS idx_dishes_description ON dishes(description); 