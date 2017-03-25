package twitch

import (
	"log"

	"github.com/jasonkeene/anubot-server/api/internal/handlers"
	"github.com/jasonkeene/anubot-server/store"
	"github.com/jasonkeene/anubot-server/twitch/oauth"
)

// NonceStore stores nonces.
type NonceStore interface {
	OauthNonce(userID string, tu store.TwitchUser) (nonce string, err error)
	StoreOauthNonce(userID string, tu store.TwitchUser, nonce string) (err error)
}

// NonceGenerator generates a random nonce to be used in the oauth flow.
type NonceGenerator func() string

// OauthCallbackRegistrar registers callbacks that are invoked when the oauth
// flow for a given nonce is complete.
type OauthCallbackRegistrar interface {
	RegisterCompletionCallback(nonce string, f func())
}

// OauthStartHandler responds with a URL to start the Twitch oauth flow.
// The streamer user is required to be the first to begin the oauth flow,
// followed by the bot user.
type OauthStartHandler struct {
	creds            CredentialsProvider
	nonceGen         NonceGenerator
	nonceStore       NonceStore
	oauthClientID    string
	oauthRedirectURL string
	oauthCallbacks   OauthCallbackRegistrar
}

// NewOauthStartHandler returns a new OauthStartHandler.
func NewOauthStartHandler(
	creds CredentialsProvider,
	nonceGen NonceGenerator,
	nonceStore NonceStore,
	oauthClientID string,
	oauthRedirectURL string,
	oauthCallbacks OauthCallbackRegistrar,
) *OauthStartHandler {
	return &OauthStartHandler{
		creds:            creds,
		nonceGen:         nonceGen,
		nonceStore:       nonceStore,
		oauthClientID:    oauthClientID,
		oauthRedirectURL: oauthRedirectURL,
		oauthCallbacks:   oauthCallbacks,
	}
}

// HandleEvent responds to a websocket event.
func (h *OauthStartHandler) HandleEvent(e handlers.Event, s handlers.Session) {
	resp, send := handlers.Setup(e, s)
	defer send()

	ok, tu := h.validatePayload(e.Payload)
	if !ok {
		resp.Error = handlers.InvalidPayload
		return
	}

	userID, _ := s.Authenticated()

	if tu == store.Bot {
		creds, err := h.creds.TwitchCredentials(userID)
		if err != nil {
			return
		}
		if !creds.StreamerAuthenticated {
			resp.Error = handlers.TwitchOauthStartOrderError
			return
		}
	}

	nonce, err := h.nonceStore.OauthNonce(userID, tu)
	if err == nil {
		url := oauth.URL(
			h.oauthClientID,
			h.oauthRedirectURL,
			userID,
			tu,
			nonce,
		)
		resp.Payload = url
		resp.Error = nil
		return
	}

	nonce = h.nonceGen()
	err = h.nonceStore.StoreOauthNonce(userID, tu, nonce)
	if err != nil {
		log.Printf("got an err trying to store oauth nonce: %s", err)
		return
	}
	url := oauth.URL(
		h.oauthClientID,
		h.oauthRedirectURL,
		userID,
		tu,
		nonce,
	)

	h.oauthCallbacks.RegisterCompletionCallback(nonce, func() {
		resp := handlers.Event{
			Cmd:     "twitch-oauth-complete",
			Payload: tu.String(),
		}
		err := s.Send(resp)
		if err != nil {
			log.Printf("unable to tx: %s", err)
		}
	})
	resp.Payload = url
	resp.Error = nil
}

// validatePayload returns true if the payload is valid.
func (h *OauthStartHandler) validatePayload(p interface{}) (bool, store.TwitchUser) {
	payload, ok := p.(string)
	if !ok {
		return false, 0
	}
	if payload == "streamer" {
		return true, store.Streamer
	}
	if payload == "bot" {
		return true, store.Bot
	}
	return false, 0
}

// CredentialsProvider provides Twitch credentials.
type CredentialsProvider interface {
	TwitchCredentials(userID string) (creds store.TwitchCredentials, err error)
}

// AuthClearer removes the Twitch auth information associated with the user.
type AuthClearer interface {
	TwitchClearAuth(userID string) (err error)
}

// ClearAuthHandler clears all auth data for the user.
type ClearAuthHandler struct {
	authClearer AuthClearer
}

// NewClearAuthHandler returns a new ClearAuthHandler.
func NewClearAuthHandler(ac AuthClearer) *ClearAuthHandler {
	return &ClearAuthHandler{
		authClearer: ac,
	}
}

// HandleEvent responds to a websocket event.
func (h *ClearAuthHandler) HandleEvent(e handlers.Event, s handlers.Session) {
	resp, send := handlers.Setup(e, s)
	defer send()

	userID, _ := s.Authenticated()
	err := h.authClearer.TwitchClearAuth(userID)
	if err != nil {
		return
	}
	resp.Error = nil
}
