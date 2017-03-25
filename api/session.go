package api

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
	"github.com/jasonkeene/anubot-server/api/internal/handlers"
)

// session represents the state of a given connection. This includes
// authentication information and the means of sending and receiving events to
// the connected client.
type session struct {
	id            string
	ws            *websocket.Conn
	authenticated bool
	userID        string
}

// Send sends an event to the user over the websocket connection.
func (s *session) Send(e handlers.Event) error {
	message, err := json.Marshal(e)
	if err != nil {
		return err
	}
	return s.ws.WriteMessage(websocket.TextMessage, message)
}

// Receive returns the next event from the websocket connection.
func (s *session) Receive() (handlers.Event, error) {
	var e handlers.Event
	mt, message, err := s.ws.ReadMessage()
	if err != nil {
		return e, err
	}
	if mt != websocket.TextMessage {
		return e, fmt.Errorf("read non-text message from websocket conn: %d", mt)
	}
	err = json.Unmarshal(message, &e)
	return e, err
}

// SetAuthentication sets the authentication for this session.
func (s *session) SetAuthentication(userID string) {
	s.authenticated = true
	s.userID = userID
}

// Authenticated lets you know what user this session is authenticated as.
func (s *session) Authenticated() (string, bool) {
	return s.userID, s.authenticated
}

// Logout clears the authentication for this session.
func (s *session) Logout() {
	s.authenticated = false
	s.userID = ""
}
