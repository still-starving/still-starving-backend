CREATE EXTENSION IF NOT EXISTS postgis;

ALTER TABLE food_posts ADD COLUMN latitude DOUBLE PRECISION;
ALTER TABLE food_posts ADD COLUMN longitude DOUBLE PRECISION;
ALTER TABLE food_posts ADD COLUMN location_geo GEOGRAPHY(POINT, 4326);

CREATE INDEX idx_food_posts_location_geo ON food_posts USING GIST (location_geo);

-- Do the same for hunger broadcasts
ALTER TABLE hunger_broadcasts ADD COLUMN latitude DOUBLE PRECISION;
ALTER TABLE hunger_broadcasts ADD COLUMN longitude DOUBLE PRECISION;
ALTER TABLE hunger_broadcasts ADD COLUMN location_geo GEOGRAPHY(POINT, 4326);

CREATE INDEX idx_hunger_broadcasts_location_geo ON hunger_broadcasts USING GIST (location_geo);
