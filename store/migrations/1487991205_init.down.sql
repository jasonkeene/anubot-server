DROP TRIGGER row_mod_on_user ON "user";
DROP TRIGGER row_mod_on_nonce ON nonce;
DROP TRIGGER row_mod_on_message ON message;
DROP FUNCTION update_row_modified();

DROP TABLE message;
DROP TYPE message_source;

DROP TABLE nonce;
DROP TYPE twitch_user;

DROP TABLE "user";
