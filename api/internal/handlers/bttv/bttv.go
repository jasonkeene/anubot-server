package bttv

import (
	"log"

	"github.com/jasonkeene/anubot-server/api/internal/handlers"
	"github.com/jasonkeene/anubot-server/store"
)

// TwitchCredentialsProvider provides Twitch credentials.
type TwitchCredentialsProvider interface {
	TwitchCredentials(userID string) (creds store.TwitchCredentials, err error)
}

// EmojiProvider provides emoji from BTTV.
type EmojiProvider interface {
	Emoji(channel string) (emoji map[string]string, err error)
}

// EmojiHandler returns emoji from BTTV. If the user has authenticated their
// streamer user with Twitch it will also include channel specific emoji.
type EmojiHandler struct {
	creds TwitchCredentialsProvider
	emoji EmojiProvider
}

// NewEmojiHandler returns a new handle that provodes BTTV emoji.
func NewEmojiHandler(
	creds TwitchCredentialsProvider,
	emoji EmojiProvider,
) *EmojiHandler {
	return &EmojiHandler{
		creds: creds,
		emoji: emoji,
	}
}

// HandleEvent responds to a websocket event.
func (h *EmojiHandler) HandleEvent(e handlers.Event, s handlers.Session) {
	resp, send := handlers.Setup(e, s)
	defer send()

	userID, _ := s.Authenticated()
	creds, err := h.creds.TwitchCredentials(userID)
	if err != nil {
		log.Printf("error authenticating twitch streamer: %s", err)
		return
	}

	var streamerUsername string
	if creds.StreamerAuthenticated {
		streamerUsername = creds.StreamerUsername
	}

	payload, err := h.emoji.Emoji(streamerUsername)
	if err != nil {
		log.Printf("error getting bttv emoji: %s", err)
		resp.Error = handlers.BTTVUnavailable
		return
	}

	resp.Payload = payload
	resp.Error = nil
}
