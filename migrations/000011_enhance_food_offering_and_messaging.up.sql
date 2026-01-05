-- Migration to enhance food offering and messaging features

-- 1. Updates to food_posts
ALTER TABLE food_posts ADD COLUMN IF NOT EXISTS price DECIMAL(10, 2) DEFAULT NULL;
ALTER TABLE food_posts ADD COLUMN IF NOT EXISTS currency VARCHAR(3) DEFAULT 'EUR';

-- 2. Updates to conversations
-- Make food_post_id nullable
ALTER TABLE conversations ALTER COLUMN food_post_id DROP NOT NULL;

-- Add hunger_broadcast_id
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS hunger_broadcast_id UUID REFERENCES hunger_broadcasts(id) ON DELETE CASCADE;

-- Drop old unique index and create a more flexible one
-- This ensures uniqueness for either (post, participants) or (broadcast, participants)
DROP INDEX IF EXISTS idx_conversations_unique;

CREATE UNIQUE INDEX IF NOT EXISTS idx_conversations_unique_post 
ON conversations (food_post_id, LEAST(participant_1_id, participant_2_id), GREATEST(participant_1_id, participant_2_id))
WHERE food_post_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_conversations_unique_broadcast 
ON conversations (hunger_broadcast_id, LEAST(participant_1_id, participant_2_id), GREATEST(participant_1_id, participant_2_id))
WHERE hunger_broadcast_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_conversations_hunger_broadcast ON conversations(hunger_broadcast_id);

-- 3. Updates to messages
ALTER TABLE messages ADD COLUMN IF NOT EXISTS type VARCHAR(20) DEFAULT 'text';
ALTER TABLE messages ADD COLUMN IF NOT EXISTS metadata JSONB DEFAULT NULL;
