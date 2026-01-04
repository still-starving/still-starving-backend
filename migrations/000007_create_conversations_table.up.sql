CREATE TABLE IF NOT EXISTS conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    food_post_id UUID NOT NULL REFERENCES food_posts(id) ON DELETE CASCADE,
    participant_1_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    participant_2_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    last_message_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT different_participants CHECK (participant_1_id != participant_2_id)
);

-- Ensure unique conversation per food post and participant pair
-- Order participants to avoid duplicates (A,B) vs (B,A)
CREATE UNIQUE INDEX IF NOT EXISTS idx_conversations_unique ON conversations(
    food_post_id,
    LEAST(participant_1_id, participant_2_id),
    GREATEST(participant_1_id, participant_2_id)
);

CREATE INDEX IF NOT EXISTS idx_conversations_participant_1 ON conversations(participant_1_id);
CREATE INDEX IF NOT EXISTS idx_conversations_participant_2 ON conversations(participant_2_id);
CREATE INDEX IF NOT EXISTS idx_conversations_food_post ON conversations(food_post_id);
CREATE INDEX IF NOT EXISTS idx_conversations_last_message ON conversations(last_message_at DESC);
