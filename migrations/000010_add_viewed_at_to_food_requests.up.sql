ALTER TABLE food_requests 
ADD COLUMN viewed_at TIMESTAMP NULL;

-- Add index for performance
CREATE INDEX IF NOT EXISTS idx_food_requests_user_viewed 
ON food_requests(user_id, viewed_at) 
WHERE viewed_at IS NULL;
