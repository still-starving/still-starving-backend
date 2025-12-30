CREATE TABLE IF NOT EXISTS food_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    food_post_id UUID NOT NULL REFERENCES food_posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message TEXT,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'rejected')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(food_post_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_food_requests_food_post_id ON food_requests(food_post_id);
CREATE INDEX IF NOT EXISTS idx_food_requests_user_id ON food_requests(user_id);
CREATE INDEX IF NOT EXISTS idx_food_requests_status ON food_requests(status);
