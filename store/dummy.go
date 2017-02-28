package store

import (
	"errors"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/satori/go.uuid"

	"github.com/jasonkeene/anubot-server/stream"
)

// Dummy is a store backend that stores everything in memory.
type Dummy struct {
	mu       sync.Mutex
	users    users
	nonces   map[string]nonceRecord
	messages map[string][]stream.RXMessage
}

// NewDummy creates a new Dummy store.
func NewDummy() *Dummy {
	return &Dummy{
		users:    make(users),
		nonces:   make(map[string]nonceRecord),
		messages: make(map[string][]stream.RXMessage),
	}
}

// Close is a NOP on the dummy store.
func (d *Dummy) Close() error {
	return nil
}

// RegisterUser registers a new user returning the user ID.
func (d *Dummy) RegisterUser(username, password string) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.users.exists(username) {
		return "", ErrUsernameTaken
	}

	id := uuid.NewV4().String()
	d.users[id] = userRecord{
		Username: username,
		Password: password,
	}
	return id, nil
}

// AuthenticateUser checks to see if the given user credentials are valid. If
// they are the user ID is returned with a bool to indicate success.
func (d *Dummy) AuthenticateUser(username, password string) (string, bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	id, ur, exists := d.users.lookup(username)
	if !exists {
		return "", false, nil
	}
	if ur.Password != password {
		return "", false, nil
	}
	return id, true, nil
}

// StoreOauthNonce stores the oauth nonce.
func (d *Dummy) StoreOauthNonce(userID string, tu TwitchUser, nonce string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	switch tu {
	case Streamer:
	case Bot:
	default:
		return errors.New("bad twitch user type in CreateOauthNonce")
	}

	d.nonces[nonce] = nonceRecord{
		Nonce:   nonce,
		UserID:  userID,
		TU:      tu,
		Created: time.Now(),
	}
	return nil
}

// OauthNonceExists tells you if the provided nonce was recently created and
// not yet finished.
func (d *Dummy) OauthNonceExists(nonce string) (bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, ok := d.nonces[nonce]
	return ok, nil
}

// FinishOauthNonce completes the oauth flow, removing the nonce and storing
// the oauth data.
func (d *Dummy) FinishOauthNonce(
	nonce string,
	twitchUsername string,
	twitchUserID int,
	od OauthData,
) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	nr, ok := d.nonces[nonce]
	if !ok {
		return ErrUnknownNonce
	}

	ur := d.users[nr.UserID]
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

	delete(d.nonces, nonce)
	d.users[nr.UserID] = ur
	return nil
}

// TwitchCredentials gives you the twitch credentials for a given users.
func (d *Dummy) TwitchCredentials(userID string) (TwitchCredentials, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	ur := d.users[userID]
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
func (d *Dummy) TwitchClearAuth(userID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	ur := d.users[userID]
	ur.StreamerOD = OauthData{}
	ur.StreamerUsername = ""
	ur.StreamerID = 0
	ur.BotOD = OauthData{}
	ur.BotUsername = ""
	ur.BotID = 0
	d.users[userID] = ur
	return nil
}

// StoreMessage stores a message for a given user for later searching and
// scrollback history.
func (d *Dummy) StoreMessage(msg stream.RXMessage) error {
	key, err := messageKey(msg)
	if err != nil {
		return err
	}
	// TODO: dedupe messages?
	d.messages[key] = append(d.messages[key], msg)
	return nil
}

// FetchRecentMessages gets the recent messages for the user's channel.
func (d *Dummy) FetchRecentMessages(userID string) ([]stream.RXMessage, error) {
	creds, err := d.TwitchCredentials(userID)
	if err != nil {
		return nil, err
	}
	if !creds.StreamerAuthenticated || !creds.BotAuthenticated {
		return nil, errors.New("user is not authenticated with twitch")
	}

	messages := d.messages["twitch:"+strconv.Itoa(creds.StreamerTwitchUserID)]
	messages = append(
		messages,
		d.messages["twitch:"+strconv.Itoa(creds.BotTwitchUserID)]...,
	)

	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Twitch.Line.Time.Before(messages[j].Twitch.Line.Time)
	})

	return messages[:min(len(messages), 500)], nil
}

// QueryMessages allows the user to search for messages that match a search
// string.
func (d *Dummy) QueryMessages(userID, search string) ([]stream.RXMessage, error) {
	panic("not implemented")
}
