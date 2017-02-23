package dummy

import (
	"errors"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/satori/go.uuid"

	"github.com/jasonkeene/anubot-server/store"
	"github.com/jasonkeene/anubot-server/stream"
	"github.com/jasonkeene/anubot-server/twitch/oauth"
)

// Dummy is a store backend that stores everything in memory.
type Dummy struct {
	mu       sync.Mutex
	users    users
	nonces   map[string]nonceRecord
	messages map[string][]stream.RXMessage
}

// New creates a new Dummy store.
func New() *Dummy {
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
		return "", store.ErrUsernameTaken
	}

	id := uuid.NewV4().String()
	d.users[id] = userRecord{
		username: username,
		password: password,
	}
	return id, nil
}

// AuthenticateUser checks to see if the given user credentials are valid. If
// they are the user ID is returned with a bool to indicate success.
func (d *Dummy) AuthenticateUser(username, password string) (string, bool, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	id, c, exists := d.users.lookup(username)
	if !exists {
		return "", false, nil
	}
	if c.password != password {
		return "", false, nil
	}
	return id, true, nil
}

// CreateOauthNonce creates and returns a unique oauth nonce.
func (d *Dummy) CreateOauthNonce(userID string, tu store.TwitchUser) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	switch tu {
	case store.Streamer:
	case store.Bot:
	default:
		return "", errors.New("bad twitch user type in CreateOauthNonce")
	}

	nonce := oauth.GenerateNonce()
	d.nonces[nonce] = nonceRecord{
		userID:  userID,
		tu:      tu,
		created: time.Now(),
	}
	return nonce, nil
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
	username string,
	userID int,
	od store.OauthData,
) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	nr, ok := d.nonces[nonce]
	if !ok {
		return store.ErrUnknownNonce
	}

	ur := d.users[nr.userID]
	switch nr.tu {
	case store.Streamer:
		ur.streamerOD = od
		ur.streamerUsername = username
		ur.streamerID = userID
	case store.Bot:
		ur.botOD = od
		ur.botUsername = username
		ur.botID = userID
	default:
		return errors.New("bad twitch user type, this should never happen")
	}

	delete(d.nonces, nonce)
	d.users[nr.userID] = ur
	return nil
}

// TwitchCredentials gives you the twitch credentials for a given users.
func (d *Dummy) TwitchCredentials(userID string) (store.TwitchCredentials, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	ur := d.users[userID]
	return store.TwitchCredentials{
		StreamerAuthenticated: ur.streamerOD.AccessToken != "",
		StreamerUsername:      ur.streamerUsername,
		StreamerPassword:      ur.streamerOD.AccessToken,
		StreamerTwitchUserID:  ur.streamerID,
		BotAuthenticated:      ur.botOD.AccessToken != "",
		BotUsername:           ur.botUsername,
		BotPassword:           ur.botOD.AccessToken,
		BotTwitchUserID:       ur.botID,
	}, nil
}

// TwitchClearAuth removes all the auth data for twitch for the user.
func (d *Dummy) TwitchClearAuth(userID string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	ur := d.users[userID]
	ur.streamerOD = store.OauthData{}
	ur.streamerUsername = ""
	ur.streamerID = 0
	ur.botOD = store.OauthData{}
	ur.botUsername = ""
	ur.botID = 0
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// QueryMessages allows the user to search for messages that match a search
// string.
func (d *Dummy) QueryMessages(userID, search string) ([]stream.RXMessage, error) {
	panic("not implemented")
}
