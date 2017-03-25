package handlers

import (
	"log"
)

// Event is the structure sent over websocket connections by both ends.
type Event struct {
	Cmd       string      `json:"cmd"`        // used to dispatch event to handlers
	Payload   interface{} `json:"payload"`    // actual data being transmitted
	RequestID string      `json:"request_id"` // used for req/resp to group events together
	Error     *Error      `json:"error"`      // used to indicate an error has occurred
}

// Session represents the state of the websocket connection.
type Session interface {
	Send(e Event) (err error) // TODO consider removing
	SetAuthentication(userID string)
	Authenticated() (userID string, authenticated bool)
	Logout()
}

// EventHandler represents something that can handle events from a websocket
// connecton.
type EventHandler interface {
	HandleEvent(e Event, s Session)
}

// EventHandlerFunc is a func that can handle events from a websocket
// connection.
type EventHandlerFunc func(Event, Session)

// HandleEvent calls the event handler func and complies with the EventHandler
// interface.
func (f EventHandlerFunc) HandleEvent(e Event, s Session) {
	f(e, s)
}

// Setup creates a generic response and a func that can be used to send that
// response.
func Setup(e Event, s Session) (*Event, func()) {
	resp := &Event{
		Cmd:       e.Cmd,
		RequestID: e.RequestID,
		Error:     UnknownError,
	}
	return resp, func() {
		err := s.Send(*resp)
		if err != nil {
			log.Printf("unable to tx: %s", err)
		}
	}
}
