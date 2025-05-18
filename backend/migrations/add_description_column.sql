-- Migration to add description column to dishes table
ALTER TABLE dishes ADD COLUMN IF NOT EXISTS description TEXT; 