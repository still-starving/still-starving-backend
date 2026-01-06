-- Remove spice_level and ingredients from food_posts
ALTER TABLE food_posts DROP COLUMN IF EXISTS spice_level;
ALTER TABLE food_posts DROP COLUMN IF EXISTS ingredients;
