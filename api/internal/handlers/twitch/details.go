package twitch

import (
	"log"

	"github.com/jasonkeene/anubot-server/api/internal/handlers"
	"github.com/jasonkeene/anubot-server/twitch"
)

// Client is used to communicate with Twitch's API.
type Client interface {
	User(token string) (userData twitch.UserData, err error)
	StreamInfo(channel string) (status, game string, err error)
	Games() (games []twitch.Game)
	UpdateDescription(status, game, channel, token string) (err error)
}

// UserDetailsHandler provides information on the Twitch streamer and bot users.
type UserDetailsHandler struct {
	creds  CredentialsProvider
	client Client
}

// NewUserDetailsHandler returns a new UserDetailsHandler.
func NewUserDetailsHandler(
	creds CredentialsProvider,
	client Client,
) *UserDetailsHandler {
	return &UserDetailsHandler{
		creds:  creds,
		client: client,
	}
}

// HandleEvent responds to a websocket event.
func (h *UserDetailsHandler) HandleEvent(e handlers.Event, s handlers.Session) {
	resp, send := handlers.Setup(e, s)
	defer send()

	p := map[string]interface{}{
		// auth info
		"streamer_authenticated": false,
		"streamer_username":      "",
		// user info
		"streamer_logo":         "",
		"streamer_display_name": "",
		// stream info
		"streamer_status": "",
		"streamer_game":   "",

		// auth info
		"bot_authenticated": false,
		"bot_username":      "",
	}

	userID, _ := s.Authenticated()
	creds, err := h.creds.TwitchCredentials(userID)
	if err != nil {
		log.Printf("unable to authenticate: %s", err)
		return
	}
	if !creds.StreamerAuthenticated {
		resp.Payload = p
		resp.Error = nil
		return
	}

	userData, err := h.client.User(creds.StreamerPassword)
	if err != nil {
		log.Printf("unable to fetch info for user: %s: %s", creds.StreamerUsername, err)
		return
	}
	p["streamer_logo"] = userData.Logo
	p["streamer_display_name"] = userData.DisplayName

	status, game, err := h.client.StreamInfo(creds.StreamerUsername)
	if err != nil {
		log.Printf("unable to fetch stream info for user: %s: %s", creds.StreamerUsername, err)
		return
	}
	p["streamer_authenticated"] = creds.StreamerAuthenticated
	p["streamer_username"] = creds.StreamerUsername
	p["streamer_status"] = status
	p["streamer_game"] = game

	if !creds.BotAuthenticated {
		resp.Payload = p
		resp.Error = nil
		return
	}

	p["bot_authenticated"] = creds.BotAuthenticated
	p["bot_username"] = creds.BotUsername
	resp.Payload = p
	resp.Error = nil
}
