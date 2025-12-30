CREATE TABLE IF NOT EXISTS hunger_offers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hunger_broadcast_id UUID NOT NULL REFERENCES hunger_broadcasts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(hunger_broadcast_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_hunger_offers_broadcast_id ON hunger_offers(hunger_broadcast_id);
CREATE INDEX IF NOT EXISTS idx_hunger_offers_user_id ON hunger_offers(user_id);
