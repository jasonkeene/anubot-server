package twitch

import (
	"log"

	"github.com/jasonkeene/anubot-server/api/internal/handlers"
	"github.com/jasonkeene/anubot-server/stream"
)

// RecentMessagesFetcher fetches recent messages for the user.
type RecentMessagesFetcher interface {
	FetchRecentMessages(userID string) (msgs []stream.RXMessage, err error)
}

// StreamMessagesStore fetches recent messages and twitch credentials for the
// user.
type StreamMessagesStore interface {
	CredentialsProvider
	RecentMessagesFetcher
}

// Connector ensures a connection is made to twitch for the user.
type Connector interface {
	ConnectTwitch(user, pass, channel string)
}

// StreamMessagesHandler writes chat messages to websocket connection.
type StreamMessagesHandler struct {
	store        StreamMessagesStore
	connector    Connector
	subEndpoints []string
}

// NewStreamMessagesHandler returns a new StreamMessagesHandler.
func NewStreamMessagesHandler(
	store StreamMessagesStore,
	connector Connector,
	subEndpoints []string,
) *StreamMessagesHandler {
	return &StreamMessagesHandler{
		store:        store,
		connector:    connector,
		subEndpoints: subEndpoints,
	}
}

// HandleEvent responds to a websocket event.
func (h *StreamMessagesHandler) HandleEvent(e handlers.Event, s handlers.Session) {
	userID, _ := s.Authenticated()
	creds, err := h.store.TwitchCredentials(userID)
	if err != nil {
		log.Printf("unable to get creds: %s", err)
		return
	}

	recent, err := h.store.FetchRecentMessages(userID)
	if err == nil {
		for _, msg := range recent {
			if msg.Type != stream.Twitch {
				continue
			}
			if msg.Twitch.OwnerID == creds.BotTwitchUserID &&
				!UserMessage(&msg, creds.StreamerUsername) {
				continue
			}

			err = s.Send(handlers.Event{
				Cmd:       "chat-message",
				RequestID: e.RequestID,
				Payload: Message{
					Type: msg.Type,
					Twitch: &TMessage{
						Cmd:    msg.Twitch.Line.Cmd,
						Nick:   msg.Twitch.Line.Nick,
						Target: msg.Twitch.Line.Args[0],
						Body:   msg.Twitch.Line.Args[1],
						Time:   msg.Twitch.Line.Time,
						Tags:   msg.Twitch.Line.Tags,
					},
				},
			})
			if err != nil {
				log.Printf("unable to tx: %s", err)
			}
		}
	}

	h.connector.ConnectTwitch(
		creds.StreamerUsername,
		"oauth:"+creds.StreamerPassword,
		"#"+creds.StreamerUsername,
	)
	h.connector.ConnectTwitch(
		creds.BotUsername,
		"oauth:"+creds.BotPassword,
		"#"+creds.StreamerUsername,
	)

	mw, err := newMessageWriter(
		creds.StreamerUsername,
		"twitch:"+creds.StreamerUsername,
		"twitch:"+creds.BotUsername,
		h.subEndpoints,
		s,
		e.RequestID,
	)
	if err != nil {
		log.Printf("unable to stream messages: %s", err)
		return
	}
	go mw.StartStreamer()
	go mw.StartBot()
}

// StreamManager is used to connect and send to third party chat.
type StreamManager interface {
	Connector
	Send(msg stream.TXMessage)
}

// SendMessageHandler accepts messages to send via Twitch chat.
type SendMessageHandler struct {
	creds         CredentialsProvider
	streamManager StreamManager
}

// NewSendMessageHandler returns a new SendMessageHandler.
func NewSendMessageHandler(
	creds CredentialsProvider,
	streamManager StreamManager,
) *SendMessageHandler {
	return &SendMessageHandler{
		creds:         creds,
		streamManager: streamManager,
	}
}

// HandleEvent responds to a websocket event.
func (h *SendMessageHandler) HandleEvent(e handlers.Event, s handlers.Session) {
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

	var username, password string
	switch payload.userType {
	case "streamer":
		username, password = creds.StreamerUsername, creds.StreamerPassword
	case "bot":
		username, password = creds.BotUsername, creds.BotPassword
	}
	h.streamManager.ConnectTwitch(
		username,
		"oauth:"+password,
		"#"+creds.StreamerUsername,
	)
	h.streamManager.Send(stream.TXMessage{
		Type: stream.Twitch,
		Twitch: &stream.TXTwitch{
			Username: username,
			To:       "#" + creds.StreamerUsername,
			Message:  payload.message,
		},
	})
	resp.Error = nil
}

// sendMessagePayload represents the payload that should be sent when
// sending a message.
type sendMessagePayload struct {
	userType string
	message  string
}

// validatePayload returns true if the payload is valid.
func (h *SendMessageHandler) validatePayload(p interface{}) (bool, sendMessagePayload) {
	data, ok := p.(map[string]interface{})
	if !ok {
		return false, sendMessagePayload{}
	}
	userType, ok := data["user_type"].(string)
	if !ok {
		return false, sendMessagePayload{}
	}
	if userType != "streamer" && userType != "bot" {
		return false, sendMessagePayload{}
	}
	message, ok := data["message"].(string)
	if !ok {
		return false, sendMessagePayload{}
	}
	return true, sendMessagePayload{
		userType: userType,
		message:  message,
	}
}
