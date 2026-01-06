-- Add spice_level and ingredients to food_posts
ALTER TABLE food_posts ADD COLUMN spice_level VARCHAR(20) DEFAULT 'no_spicy';
ALTER TABLE food_posts ADD COLUMN ingredients VARCHAR(1000);

-- Add CHECK constraint for spice_level
ALTER TABLE food_posts ADD CONSTRAINT chk_spice_level CHECK (spice_level IN ('no_spicy', 'medium_spicy', 'spicy', 'very_spicy'));
