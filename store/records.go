package store

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/boltdb/bolt"

	"github.com/jasonkeene/anubot-server/stream"
)

type userRecord struct {
	UserID           string    `json:"user_id"`
	Username         string    `json:"username"`
	Password         string    `json:"password"`
	StreamerUsername string    `json:"streamer_username"`
	StreamerOD       OauthData `json:"streamer_od"`
	StreamerID       int       `json:"streamer_id"`
	BotUsername      string    `json:"bot_username"`
	BotOD            OauthData `json:"bot_od"`
	BotID            int       `json:"bot_id"`
}

type nonceRecord struct {
	Nonce   string     `json:"nonce"`
	UserID  string     `json:"user_id"`
	TU      TwitchUser `json:"tu"`
	Created time.Time  `json:"created"`
}

type messageRecord []stream.RXMessage

func upsertNonceRecord(nr nonceRecord, tx *bolt.Tx) error {
	b := tx.Bucket([]byte("nonces"))

	nrb, err := json.Marshal(nr)
	if err != nil {
		return err
	}
	return b.Put([]byte(nr.Nonce), nrb)
}

func getNonceRecordByUserID(userID string, tu TwitchUser, tx *bolt.Tx) (string, error) {
	b := tx.Bucket([]byte("nonces"))

	c := b.Cursor()
	for k, v := c.First(); k != nil; k, v = c.Next() {
		var nr nonceRecord
		err := json.Unmarshal(v, &nr)
		if err != nil {
			continue
		}
		if nr.UserID == userID && nr.TU == tu {
			return nr.Nonce, nil
		}
	}
	return "", ErrUnknownNonce
}

func getNonceRecord(nonce string, tx *bolt.Tx) (nonceRecord, error) {
	b := tx.Bucket([]byte("nonces"))

	read := b.Get([]byte(nonce))
	if read == nil {
		return nonceRecord{}, ErrUnknownNonce
	}

	var nr nonceRecord
	err := json.Unmarshal(read, &nr)
	if err != nil {
		return nonceRecord{}, err
	}
	return nr, nil
}

func deleteNonceRecord(nonce string, tx *bolt.Tx) error {
	b := tx.Bucket([]byte("nonces"))

	return b.Delete([]byte(nonce))
}

func upsertUserRecord(ur userRecord, tx *bolt.Tx) error {
	b := tx.Bucket([]byte("users"))

	urb, err := json.Marshal(ur)
	if err != nil {
		return err
	}
	return b.Put([]byte(ur.UserID), urb)
}

func getUserRecord(userID string, tx *bolt.Tx) (userRecord, error) {
	b := tx.Bucket([]byte("users"))

	read := b.Get([]byte(userID))
	if read == nil {
		return userRecord{}, ErrUnknownUserID
	}

	var ur userRecord
	err := json.Unmarshal(read, &ur)
	if err != nil {
		return userRecord{}, err
	}
	return ur, nil
}

func getUserRecordByUsername(username string, tx *bolt.Tx) (userRecord, error) {
	b := tx.Bucket([]byte("users"))

	c := b.Cursor()
	for k, v := c.First(); k != nil; k, v = c.Next() {
		var ur userRecord
		err := json.Unmarshal(v, &ur)
		if err != nil {
			continue
		}
		if username == ur.Username {
			return ur, nil
		}
	}
	return userRecord{}, ErrUnknownUsername
}

func upsertMessage(msg stream.RXMessage, tx *bolt.Tx) error {
	b := tx.Bucket([]byte("messages"))

	key, err := getMessageKey(msg)
	if err != nil {
		println("return via nil key")
		return err
	}

	mrb := b.Get([]byte(key))
	var mr messageRecord
	if mrb != nil {
		err = json.Unmarshal(mrb, &mr)
		if err != nil {
			println("return via unmarshal err")
			return err
		}
	}

	mr = append(mr, msg)
	mrb, err = json.Marshal(mr)
	if err != nil {
		println("return via marshal err")
		return err
	}

	return b.Put([]byte(key), mrb)
}

func getMessageRecord(key string, tx *bolt.Tx) (messageRecord, error) {
	b := tx.Bucket([]byte("messages"))

	read := b.Get([]byte(key))
	if read == nil {
		return make(messageRecord, 0), nil
	}

	var mr messageRecord
	err := json.Unmarshal(read, &mr)
	if err != nil {
		return nil, err
	}
	return mr, nil
}

func getMessageKey(msg stream.RXMessage) (string, error) {
	switch msg.Type {
	case stream.Twitch:
		return "twitch:" + strconv.Itoa(msg.Twitch.OwnerID), nil
	case stream.Discord:
		return "discord:" + msg.Discord.OwnerID, nil
	default:
		return "", errors.New("invalid message type")
	}
}

type users map[string]userRecord

func (u users) lookup(username string) (string, userRecord, bool) {
	for id, creds := range u {
		if creds.Username == username {
			return id, creds, true
		}
	}
	return "", userRecord{}, false
}

func (u users) exists(username string) bool {
	_, _, exists := u.lookup(username)
	return exists
}

func messageKey(msg stream.RXMessage) (string, error) {
	switch msg.Type {
	case stream.Twitch:
		return "twitch:" + strconv.Itoa(msg.Twitch.OwnerID), nil
	case stream.Discord:
		return "discord:" + msg.Discord.OwnerID, nil
	default:
		return "", errors.New("invalid message type")
	}
}
