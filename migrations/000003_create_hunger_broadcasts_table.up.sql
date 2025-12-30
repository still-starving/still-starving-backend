CREATE TABLE IF NOT EXISTS hunger_broadcasts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    location VARCHAR(255) NOT NULL,
    urgency VARCHAR(20) DEFAULT 'normal' CHECK (urgency IN ('normal', 'urgent')),
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'expired', 'fulfilled')),
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_hunger_broadcasts_user_id ON hunger_broadcasts(user_id);
CREATE INDEX IF NOT EXISTS idx_hunger_broadcasts_status ON hunger_broadcasts(status);
CREATE INDEX IF NOT EXISTS idx_hunger_broadcasts_created_at ON hunger_broadcasts(created_at DESC);
