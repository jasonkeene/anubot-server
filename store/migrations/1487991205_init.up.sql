
CREATE TABLE "user" (
    created  TIMESTAMP WITH TIME ZONE DEFAULT CLOCK_TIMESTAMP(),
    modified TIMESTAMP WITH TIME ZONE DEFAULT CLOCK_TIMESTAMP(),

    user_id       UUID PRIMARY KEY,
    username      VARCHAR(255),
    password_hash VARCHAR(255),

    streamer_id         INTEGER,
    streamer_username   VARCHAR(255), -- this is denormalized data, hit api for actual value
    streamer_oauth_data TEXT,

    bot_id         INTEGER,
    bot_username   VARCHAR(255), -- this is denormalized data, hit api for actual value
    bot_oauth_data TEXT,

    UNIQUE(username),
    UNIQUE(streamer_id),
    UNIQUE(streamer_username),
    UNIQUE(bot_id),
    UNIQUE(bot_username)
);

CREATE TYPE twitch_user AS ENUM ('Streamer', 'Bot');

CREATE TABLE nonce (
    created  TIMESTAMP WITH TIME ZONE DEFAULT CLOCK_TIMESTAMP(),
    modified TIMESTAMP WITH TIME ZONE DEFAULT CLOCK_TIMESTAMP(),

    user_id     UUID REFERENCES "user",
    twitch_user twitch_user,
    nonce       TEXT,

    PRIMARY KEY(user_id, twitch_user),
    UNIQUE(nonce)
);

CREATE TYPE message_source AS ENUM ('Twitch', 'Discord');

CREATE TABLE message (
    created  TIMESTAMP WITH TIME ZONE DEFAULT CLOCK_TIMESTAMP(),
    modified TIMESTAMP WITH TIME ZONE DEFAULT CLOCK_TIMESTAMP(),

    source  message_source,
    message TEXT,
    twitch_owner_id INTEGER,
    discord_owner_id VARCHAR(20)
);

CREATE OR REPLACE FUNCTION update_row_modified()
RETURNS TRIGGER
AS $$
BEGIN
    NEW.modified = CLOCK_TIMESTAMP();
    RETURN NEW;
END;
$$ LANGUAGE PLPGSQL;

CREATE TRIGGER row_mod_on_user
BEFORE UPDATE
ON "user"
FOR EACH ROW
EXECUTE PROCEDURE update_row_modified();

CREATE TRIGGER row_mod_on_nonce
BEFORE UPDATE
ON nonce
FOR EACH ROW
EXECUTE PROCEDURE update_row_modified();

CREATE TRIGGER row_mod_on_message
BEFORE UPDATE
ON message
FOR EACH ROW
EXECUTE PROCEDURE update_row_modified();
