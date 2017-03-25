package twitch

import "github.com/jasonkeene/anubot-server/api/internal/handlers"

// GamesHandler returns the available games.
type GamesHandler struct {
	client Client
}

// NewGamesHandler returns a new GamesHandler.
func NewGamesHandler(tc Client) *GamesHandler {
	return &GamesHandler{
		client: tc,
	}
}

// HandleEvent responds to a websocket event.
func (h GamesHandler) HandleEvent(e handlers.Event, s handlers.Session) {
	resp, send := handlers.Setup(e, s)
	defer send()

	resp.Payload = h.client.Games()
	resp.Error = nil
}
