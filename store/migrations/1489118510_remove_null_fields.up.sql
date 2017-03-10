-- user table
DELETE FROM "user" WHERE created IS NULL;
ALTER TABLE "user" ALTER COLUMN created SET NOT NULL;

DELETE FROM "user" WHERE modified IS NULL;
ALTER TABLE "user" ALTER COLUMN modified SET NOT NULL;

DELETE FROM "user" WHERE username IS NULL;
ALTER TABLE "user" ALTER COLUMN username SET NOT NULL;

DELETE FROM "user" WHERE password_hash IS NULL;
ALTER TABLE "user" ALTER COLUMN password_hash SET NOT NULL;

ALTER TABLE "user" DROP CONSTRAINT user_streamer_id_key;
UPDATE "user" set streamer_id = 0 WHERE streamer_id IS NULL;
ALTER TABLE "user" ALTER COLUMN streamer_id SET NOT NULL;
ALTER TABLE "user" ALTER COLUMN streamer_id SET DEFAULT 0;

ALTER TABLE "user" DROP CONSTRAINT user_streamer_username_key;
UPDATE "user" set streamer_username = '' WHERE streamer_username IS NULL;
ALTER TABLE "user" ALTER COLUMN streamer_username SET NOT NULL;
ALTER TABLE "user" ALTER COLUMN streamer_username SET DEFAULT '';

UPDATE "user" set streamer_oauth_data = '' WHERE streamer_oauth_data IS NULL;
ALTER TABLE "user" ALTER COLUMN streamer_oauth_data SET NOT NULL;
ALTER TABLE "user" ALTER COLUMN streamer_oauth_data SET DEFAULT '';

ALTER TABLE "user" DROP CONSTRAINT user_bot_id_key;
UPDATE "user" set bot_id = 0 WHERE bot_id IS NULL;
ALTER TABLE "user" ALTER COLUMN bot_id SET NOT NULL;
ALTER TABLE "user" ALTER COLUMN bot_id SET DEFAULT 0;

ALTER TABLE "user" DROP CONSTRAINT user_bot_username_key;
UPDATE "user" set bot_username = '' WHERE bot_username IS NULL;
ALTER TABLE "user" ALTER COLUMN bot_username SET NOT NULL;
ALTER TABLE "user" ALTER COLUMN bot_username SET DEFAULT '';

UPDATE "user" set bot_oauth_data = '' WHERE bot_oauth_data IS NULL;
ALTER TABLE "user" ALTER COLUMN bot_oauth_data SET NOT NULL;
ALTER TABLE "user" ALTER COLUMN bot_oauth_data SET DEFAULT '';

-- nonce table
DELETE FROM nonce WHERE created IS NULL;
ALTER TABLE nonce ALTER COLUMN created SET NOT NULL;

DELETE FROM nonce WHERE modified IS NULL;
ALTER TABLE nonce ALTER COLUMN modified SET NOT NULL;

DELETE FROM nonce WHERE nonce IS NULL;
ALTER TABLE nonce ALTER COLUMN nonce SET NOT NULL;

-- message table
DELETE FROM message WHERE created IS NULL;
ALTER TABLE message ALTER COLUMN created SET NOT NULL;

DELETE FROM message WHERE modified IS NULL;
ALTER TABLE message ALTER COLUMN modified SET NOT NULL;

DELETE FROM message WHERE source IS NULL;
ALTER TABLE message ALTER COLUMN source SET NOT NULL;

DELETE FROM message WHERE message IS NULL;
ALTER TABLE message ALTER COLUMN message SET NOT NULL;

UPDATE message set twitch_owner_id = 0 WHERE twitch_owner_id IS NULL;
ALTER TABLE message ALTER COLUMN twitch_owner_id SET NOT NULL;
ALTER TABLE message ALTER COLUMN twitch_owner_id SET DEFAULT 0;

UPDATE message set discord_owner_id = '' WHERE discord_owner_id IS NULL;
ALTER TABLE message ALTER COLUMN discord_owner_id SET NOT NULL;
ALTER TABLE message ALTER COLUMN discord_owner_id SET DEFAULT '';
