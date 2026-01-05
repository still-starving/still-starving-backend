-- Revert to the original unique constraint
DROP INDEX IF EXISTS idx_food_requests_unique_pending;

ALTER TABLE food_requests 
ADD CONSTRAINT food_requests_food_post_id_user_id_key 
UNIQUE(food_post_id, user_id);
