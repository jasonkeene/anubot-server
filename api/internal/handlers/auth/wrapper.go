package auth

import (
	"log"

	"github.com/jasonkeene/anubot-server/api/internal/handlers"
)

// AuthenticateWrapper wraps a handler and makes sure the session is
// authenticated.
func AuthenticateWrapper(h handlers.EventHandler) handlers.EventHandler {
	return handlers.EventHandlerFunc(func(e handlers.Event, s handlers.Session) {
		_, ok := s.Authenticated()
		if ok {
			h.HandleEvent(e, s)
			return
		}
		err := s.Send(handlers.Event{
			Cmd:       e.Cmd,
			RequestID: e.RequestID,
			Error:     handlers.AuthenticationError,
		})
		if err != nil {
			log.Printf("unable to tx: %s", err)
		}
	})
}
