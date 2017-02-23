package store

import "github.com/jasonkeene/anubot-server/stream"

// Store is the interface all storage backends implement.
type Store interface {
	// Close cleans up the resources associated with the storage backend.
	Close() (err error)

	// RegisterUser registers a new user returning the user ID.
	RegisterUser(username, password string) (userID string, err error)

	// AuthenticateUser checks to see if the given user credentials are valid.
	// If they are the user ID is returned with a bool to indicate success.
	AuthenticateUser(username, password string) (userID string, success bool, err error)

	// StoreOauthNonce stores the oauth nonce.
	StoreOauthNonce(userID string, tu TwitchUser, nonce string) (err error)

	// OauthNonceExists tells you if the provided nonce was recently created
	// and not yet finished.
	OauthNonceExists(nonce string) (exists bool, err error)

	// FinishOauthNonce completes the oauth flow, removing the nonce and
	// storing the oauth data.
	FinishOauthNonce(nonce, username string, userID int, od OauthData) (err error)

	// TwitchCredentials gives you the status of the user's authentication
	// with twitch.
	TwitchCredentials(userID string) (creds TwitchCredentials, err error)

	// TwitchClearAuth removes all the auth data for twitch for the user.
	TwitchClearAuth(userID string) (err error)

	// StoreMessage stores a message for a given user for later searching and
	// scrollback history.
	StoreMessage(msg stream.RXMessage) (err error)

	// FetchRecentMessages gets the recent messages for the user's channel.
	FetchRecentMessages(userID string) (msgs []stream.RXMessage, err error)

	// QueryMessages allows the user to search for messages that match a
	// search string.
	QueryMessages(userID, search string) (msgs []stream.RXMessage, err error)
}

// TwitchCredentials represents a user's twitch authentication information for
// both streamer and bot users.
type TwitchCredentials struct {
	StreamerAuthenticated bool
	BotAuthenticated      bool

	StreamerUsername     string
	StreamerPassword     string
	StreamerTwitchUserID int
	BotUsername          string
	BotPassword          string
	BotTwitchUserID      int
}
