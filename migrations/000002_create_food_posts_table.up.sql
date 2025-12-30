CREATE TABLE IF NOT EXISTS food_posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    quantity VARCHAR(100) NOT NULL,
    location VARCHAR(255) NOT NULL,
    expiry_date TIMESTAMP NOT NULL,
    image_url TEXT,
    status VARCHAR(20) DEFAULT 'available' CHECK (status IN ('available', 'claimed', 'expired')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_food_posts_user_id ON food_posts(user_id);
CREATE INDEX IF NOT EXISTS idx_food_posts_status ON food_posts(status);
CREATE INDEX IF NOT EXISTS idx_food_posts_created_at ON food_posts(created_at DESC);
