-- Remove cooked_at from food_posts
ALTER TABLE food_posts DROP COLUMN IF EXISTS cooked_at;
