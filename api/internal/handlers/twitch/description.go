package twitch

import (
	"log"

	"github.com/jasonkeene/anubot-server/api/internal/handlers"
)

// UpdateChatDescriptionHandler updates the chat description for Twitch.
type UpdateChatDescriptionHandler struct {
	creds  CredentialsProvider
	client Client
}

// NewUpdateChatDescriptionHandler returns a new UpdateChatDescriptionHandler.
func NewUpdateChatDescriptionHandler(
	creds CredentialsProvider,
	client Client,
) *UpdateChatDescriptionHandler {
	return &UpdateChatDescriptionHandler{
		creds:  creds,
		client: client,
	}
}

// HandleEvent responds to a websocket event.
func (h *UpdateChatDescriptionHandler) HandleEvent(e handlers.Event, s handlers.Session) {
	resp, send := handlers.Setup(e, s)
	defer send()

	ok, payload := h.validatePayload(e.Payload)
	if !ok {
		resp.Error = handlers.InvalidPayload
		return
	}

	userID, _ := s.Authenticated()
	creds, err := h.creds.TwitchCredentials(userID)
	if err != nil {
		log.Printf("unable to get creds: %s", err)
		return
	}

	err = h.client.UpdateDescription(
		payload.status,
		payload.game,
		creds.StreamerUsername,
		creds.StreamerPassword,
	)
	if err != nil {
		log.Println("unable to update chat description, got error:", err)
		return
	}
	resp.Error = nil
}

// updateChatDescriptionPayload represents the payload that should be sent when
// updating the chat description.
type updateChatDescriptionPayload struct {
	status string
	game   string
}

// validatePayload returns true if the payload is valid.
func (h *UpdateChatDescriptionHandler) validatePayload(p interface{}) (bool, updateChatDescriptionPayload) {
	payload, ok := p.(map[string]interface{})
	if !ok {
		return false, updateChatDescriptionPayload{}
	}
	status, ok := payload["status"].(string)
	if !ok {
		return false, updateChatDescriptionPayload{}
	}
	game, ok := payload["game"].(string)
	if !ok {
		return false, updateChatDescriptionPayload{}
	}

	return true, updateChatDescriptionPayload{
		status: status,
		game:   game,
	}
}
