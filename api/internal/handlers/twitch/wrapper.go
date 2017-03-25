package twitch

import (
	"log"

	"github.com/jasonkeene/anubot-server/api/internal/handlers"
)

// AuthenticateWrapper wraps a handler and makes sure the user attached
// to the session is properly authenticated with Twitch.
func AuthenticateWrapper(
	credsProvider CredentialsProvider,
	h handlers.EventHandler,
) handlers.EventHandler {
	return handlers.EventHandlerFunc(func(e handlers.Event, s handlers.Session) {
		userID, _ := s.Authenticated()
		creds, err := credsProvider.TwitchCredentials(userID)

		if err != nil {
			err := s.Send(handlers.Event{
				Cmd:       e.Cmd,
				RequestID: e.RequestID,
				Error:     handlers.UnknownError,
			})
			if err != nil {
				log.Printf("unable to tx: %s", err)
			}
			return
		}
		if creds.StreamerAuthenticated && creds.BotAuthenticated {
			h.HandleEvent(e, s)
			return
		}
		err = s.Send(handlers.Event{
			Cmd:       e.Cmd,
			RequestID: e.RequestID,
			Error:     handlers.TwitchAuthenticationError,
		})
		if err != nil {
			log.Printf("unable to tx: %s", err)
		}
	})
}
