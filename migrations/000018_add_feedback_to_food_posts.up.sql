-- Add feedback and claim tracking to food_posts
ALTER TABLE food_posts ADD COLUMN IF NOT EXISTS claimed_by_user_id UUID REFERENCES users(id);
ALTER TABLE food_posts ADD COLUMN IF NOT EXISTS rating INT CHECK (rating >= 1 AND rating <= 5);
ALTER TABLE food_posts ADD COLUMN IF NOT EXISTS review TEXT;
ALTER TABLE food_posts ADD COLUMN IF NOT EXISTS reviewed_at TIMESTAMP;

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_food_posts_claimed_by_user_id ON food_posts(claimed_by_user_id);
