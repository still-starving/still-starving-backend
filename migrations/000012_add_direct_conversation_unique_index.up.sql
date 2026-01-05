-- Add unique index for direct coordination conversations (both contexts NULL)
-- Uses LEAST/GREATEST to ensure (UserA, UserB) is same as (UserB, UserA)
CREATE UNIQUE INDEX IF NOT EXISTS idx_conversations_unique_direct
ON conversations (LEAST(participant_1_id, participant_2_id), GREATEST(participant_1_id, participant_2_id))
WHERE food_post_id IS NULL AND hunger_broadcast_id IS NULL;
