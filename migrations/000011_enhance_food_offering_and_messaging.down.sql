-- Revert enhancements to food offering and messaging

-- 3. Revert messages updates
ALTER TABLE messages DROP COLUMN IF EXISTS metadata;
ALTER TABLE messages DROP COLUMN IF EXISTS type;

-- 2. Revert conversations updates
DROP INDEX IF EXISTS idx_conversations_hunger_broadcast;
DROP INDEX IF EXISTS idx_conversations_unique_broadcast;
DROP INDEX IF EXISTS idx_conversations_unique_post;

CREATE UNIQUE INDEX IF NOT EXISTS idx_conversations_unique ON conversations(
    food_post_id,
    LEAST(participant_1_id, participant_2_id),
    GREATEST(participant_1_id, participant_2_id)
);

ALTER TABLE conversations DROP COLUMN IF EXISTS hunger_broadcast_id;
ALTER TABLE conversations ALTER COLUMN food_post_id SET NOT NULL;

-- 1. Revert food_posts updates
ALTER TABLE food_posts DROP COLUMN IF EXISTS currency;
ALTER TABLE food_posts DROP COLUMN IF EXISTS price;
