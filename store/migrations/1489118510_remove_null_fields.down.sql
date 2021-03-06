-- user table
ALTER TABLE "user" ALTER COLUMN created DROP NOT NULL;

ALTER TABLE "user" ALTER COLUMN modified DROP NOT NULL;

ALTER TABLE "user" ALTER COLUMN username DROP NOT NULL;

ALTER TABLE "user" ALTER COLUMN password_hash DROP NOT NULL;

ALTER TABLE "user" ALTER COLUMN streamer_id DROP NOT NULL;
ALTER TABLE "user" ALTER COLUMN streamer_id DROP DEFAULT;

ALTER TABLE "user" ALTER COLUMN streamer_username DROP NOT NULL;
ALTER TABLE "user" ALTER COLUMN streamer_username DROP DEFAULT;

ALTER TABLE "user" ALTER COLUMN streamer_oauth_data DROP NOT NULL;
ALTER TABLE "user" ALTER COLUMN streamer_oauth_data DROP DEFAULT;

ALTER TABLE "user" ALTER COLUMN bot_id DROP NOT NULL;
ALTER TABLE "user" ALTER COLUMN bot_id DROP DEFAULT;

ALTER TABLE "user" ALTER COLUMN bot_username DROP NOT NULL;
ALTER TABLE "user" ALTER COLUMN bot_username DROP DEFAULT;

ALTER TABLE "user" ALTER COLUMN bot_oauth_data DROP NOT NULL;
ALTER TABLE "user" ALTER COLUMN bot_oauth_data DROP DEFAULT;

-- nonce table
ALTER TABLE nonce ALTER COLUMN created DROP NOT NULL;

ALTER TABLE nonce ALTER COLUMN modified DROP NOT NULL;

ALTER TABLE nonce ALTER COLUMN nonce DROP NOT NULL;

-- message table
ALTER TABLE message ALTER COLUMN created DROP NOT NULL;

ALTER TABLE message ALTER COLUMN modified DROP NOT NULL;

ALTER TABLE message ALTER COLUMN source DROP NOT NULL;

ALTER TABLE message ALTER COLUMN message DROP NOT NULL;

ALTER TABLE message ALTER COLUMN twitch_owner_id DROP NOT NULL;
ALTER TABLE message ALTER COLUMN twitch_owner_id DROP DEFAULT;

ALTER TABLE message ALTER COLUMN discord_owner_id DROP NOT NULL;
ALTER TABLE message ALTER COLUMN discord_owner_id DROP DEFAULT;
