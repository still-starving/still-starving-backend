DROP INDEX IF EXISTS idx_food_posts_location_geo;
ALTER TABLE food_posts DROP COLUMN IF EXISTS location_geo;
ALTER TABLE food_posts DROP COLUMN IF EXISTS latitude;
ALTER TABLE food_posts DROP COLUMN IF EXISTS longitude;

DROP INDEX IF EXISTS idx_hunger_broadcasts_location_geo;
ALTER TABLE hunger_broadcasts DROP COLUMN IF EXISTS location_geo;
ALTER TABLE hunger_broadcasts DROP COLUMN IF EXISTS latitude;
ALTER TABLE hunger_broadcasts DROP COLUMN IF EXISTS longitude;

-- We don't drop the extension in case other tables utilize it in the future, 
-- or if it was installed by a superuser and we lack permissions to drop it safely without cascading.
-- DROP EXTENSION IF EXISTS postgis;
