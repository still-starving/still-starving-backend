DROP INDEX IF EXISTS idx_food_requests_user_viewed;

ALTER TABLE food_requests 
DROP COLUMN IF EXISTS viewed_at;
