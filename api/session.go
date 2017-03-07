package api

import (
	"encoding/json"
	"fmt"

	"github.com/gorilla/websocket"
)

// session stores objects handlers need when responding to events.
type session struct {
	id  string
	ws  *websocket.Conn
	api *Server

	authenticated bool
	userID        string
}

// Send sends an event to the user over the websocket connection.
func (s *session) Send(e event) error {
	message, err := json.Marshal(e)
	if err != nil {
		return err
	}
	return s.ws.WriteMessage(websocket.TextMessage, message)
}

// Receive returns the next event from the websocket connection.
func (s *session) Receive() (event, error) {
	var e event
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
func (s *session) SetAuthentication(id string) {
	s.authenticated = true
	s.userID = id
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
