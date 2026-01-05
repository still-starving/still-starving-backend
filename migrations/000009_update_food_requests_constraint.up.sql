-- Drop the old unique constraint
ALTER TABLE food_requests DROP CONSTRAINT IF EXISTS food_requests_food_post_id_user_id_key;

-- Add a partial unique index that only applies to pending requests
CREATE UNIQUE INDEX IF NOT EXISTS idx_food_requests_unique_pending 
ON food_requests(food_post_id, user_id) 
WHERE status = 'pending';
