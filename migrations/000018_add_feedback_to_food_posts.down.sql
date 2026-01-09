-- Remove feedback and claim tracking from food_posts
DROP INDEX IF EXISTS idx_food_posts_claimed_by_user_id;
ALTER TABLE food_posts DROP COLUMN IF EXISTS reviewed_at;
ALTER TABLE food_posts DROP COLUMN IF EXISTS review;
ALTER TABLE food_posts DROP COLUMN IF EXISTS rating;
ALTER TABLE food_posts DROP COLUMN IF EXISTS claimed_by_user_id;
