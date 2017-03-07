package store

import (
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
	uuid "github.com/satori/go.uuid"

	"github.com/jasonkeene/anubot-server/stream"
)

// Bolt is a store backend for boltdb.
type Bolt struct {
	db *bolt.DB
}

// NewBolt creates a new bolt store.
func NewBolt(path string) (*Bolt, error) {
	db, err := bolt.Open(path, 0600, &bolt.Options{
		Timeout: time.Second,
	})
	if err != nil {
		return nil, err
	}
	b := &Bolt{
		db: db,
	}
	err = b.createBuckets()
	if err != nil {
		closeErr := db.Close()
		if closeErr != nil {
			log.Printf("got an error closing bolt db while in error state: %s", err)
		}
		return nil, err
	}
	return b, nil
}

func (b *Bolt) createBuckets() error {
	return b.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("users"))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte("nonces"))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte("messages"))
		return err
	})
}

// Close cleans up the boltdb resources.
func (b *Bolt) Close() error {
	return b.db.Close()
}

// RegisterUser registers a new user returning the user ID.
func (b *Bolt) RegisterUser(username, password string) (string, error) {
	userID := uuid.NewV4().String()
	ur := userRecord{
		UserID:   userID,
		Username: username,
		Password: password,
	}

	err := b.db.Update(func(tx *bolt.Tx) error {
		_, err := getUserRecordByUsername(username, tx)
		if err == nil {
			return ErrUsernameTaken
		}
		return upsertUserRecord(ur, tx)
	})
	if err != nil {
		return "", err
	}

	return userID, nil
}

// AuthenticateUser checks to see if the given user credentials are valid. If
// they are the user ID is returned with a bool to indicate success.
func (b *Bolt) AuthenticateUser(username, password string) (string, bool, error) {
	var ur userRecord
	err := b.db.View(func(tx *bolt.Tx) error {
		var err error
		ur, err = getUserRecordByUsername(username, tx)
		return err
	})
	if err != nil {
		return "", false, err
	}

	if ur.Password != password {
		return "", false, nil
	}
	return ur.UserID, true, nil
}

// OauthNonce gets the oauth nonce for a given user if it exists.
func (b *Bolt) OauthNonce(userID string, tu TwitchUser) (nonce string, err error) {
	err = b.db.View(func(tx *bolt.Tx) error {
		var err error
		nonce, err = getNonceRecordByUserID(userID, tu, tx)
		return err
	})
	if err != nil {
		return "", ErrUnknownNonce
	}
	return nonce, nil
}

// StoreOauthNonce stores the oauth nonce.
func (b *Bolt) StoreOauthNonce(userID string, tu TwitchUser, nonce string) error {
	nr := nonceRecord{
		Nonce:   nonce,
		UserID:  userID,
		TU:      tu,
		Created: time.Now(),
	}

	return b.db.Update(func(tx *bolt.Tx) error {
		return upsertNonceRecord(nr, tx)
	})
}

// OauthNonceExists tells you if the provided nonce was recently created and
// not yet finished.
func (b *Bolt) OauthNonceExists(nonce string) (bool, error) {
	err := b.db.View(func(tx *bolt.Tx) error {
		_, err := getNonceRecord(nonce, tx)
		return err
	})
	return err == nil, err
}

// FinishOauthNonce completes the oauth flow, removing the nonce and storing
// the oauth data.
func (b *Bolt) FinishOauthNonce(
	nonce string,
	twitchUsername string,
	twitchUserID int,
	od OauthData,
) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		nr, err := getNonceRecord(nonce, tx)
		if err != nil {
			return err
		}

		ur, err := getUserRecord(nr.UserID, tx)
		if err != nil {
			return err
		}

		switch nr.TU {
		case Streamer:
			ur.StreamerOD = od
			ur.StreamerUsername = twitchUsername
			ur.StreamerID = twitchUserID
		case Bot:
			ur.BotOD = od
			ur.BotUsername = twitchUsername
			ur.BotID = twitchUserID
		default:
			return errors.New("bad twitch user type, this should never happen")
		}

		err = deleteNonceRecord(nr.Nonce, tx)
		if err != nil {
			return err
		}

		return upsertUserRecord(ur, tx)
	})
}

// TwitchCredentials gives you the twitch credentials for a given users.
func (b *Bolt) TwitchCredentials(userID string) (TwitchCredentials, error) {
	var ur userRecord
	err := b.db.View(func(tx *bolt.Tx) error {
		var err error
		ur, err = getUserRecord(userID, tx)
		return err
	})
	if err != nil {
		return TwitchCredentials{}, err
	}

	return TwitchCredentials{
		StreamerAuthenticated: ur.StreamerOD.AccessToken != "",
		StreamerUsername:      ur.StreamerUsername,
		StreamerPassword:      ur.StreamerOD.AccessToken,
		StreamerTwitchUserID:  ur.StreamerID,
		BotAuthenticated:      ur.BotOD.AccessToken != "",
		BotUsername:           ur.BotUsername,
		BotPassword:           ur.BotOD.AccessToken,
		BotTwitchUserID:       ur.BotID,
	}, nil
}

// TwitchClearAuth removes all the auth data for twitch for the user.
func (b *Bolt) TwitchClearAuth(userID string) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		ur, err := getUserRecord(userID, tx)
		if err != nil {
			return err
		}
		ur.StreamerUsername = ""
		ur.StreamerOD = OauthData{}
		ur.StreamerID = 0
		ur.BotUsername = ""
		ur.BotOD = OauthData{}
		ur.BotID = 0
		return upsertUserRecord(ur, tx)
	})
}

// StoreMessage stores a message for a given user for later searching and
// scrollback history.
func (b *Bolt) StoreMessage(msg stream.RXMessage) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		return upsertMessage(msg, tx)
	})
}

// FetchRecentMessages gets the recent messages for the user's channel.
func (b *Bolt) FetchRecentMessages(userID string) ([]stream.RXMessage, error) {
	creds, err := b.TwitchCredentials(userID)
	if err != nil {
		return nil, err
	}
	if !creds.StreamerAuthenticated || !creds.BotAuthenticated {
		return nil, errors.New("user is not authenticated with twitch")
	}

	var mr messageRecord
	err = b.db.View(func(tx *bolt.Tx) error {
		var err error
		mr, err = getMessageRecord("twitch:"+strconv.Itoa(creds.StreamerTwitchUserID), tx)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("could not query messages for streamer: %s", err)
	}
	messages := []stream.RXMessage(mr)

	err = b.db.View(func(tx *bolt.Tx) error {
		var err error
		mr, err = getMessageRecord("twitch:"+strconv.Itoa(creds.BotTwitchUserID), tx)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("could not query messages for bot: %s", err)
	}
	messages = append(messages, []stream.RXMessage(mr)...)

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Twitch.Line.Time.Before(messages[j].Twitch.Line.Time)
	})

	return messages[:min(len(messages), 500)], nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// QueryMessages allows the user to search for messages that match a
// search string.
func (b *Bolt) QueryMessages(userID, search string) ([]stream.RXMessage, error) {
	panic("not implemented")
}
