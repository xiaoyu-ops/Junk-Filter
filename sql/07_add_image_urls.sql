-- Migration: Add image_urls column to content table
-- Stores extracted image URLs from RSS feed items as JSONB array

ALTER TABLE content ADD COLUMN IF NOT EXISTS image_urls JSONB DEFAULT '[]';
