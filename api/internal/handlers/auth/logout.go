package auth

import "github.com/jasonkeene/anubot-server/api/internal/handlers"

// LogoutHandler clears the authentication for the session.
func LogoutHandler(e handlers.Event, s handlers.Session) {
	resp, send := handlers.Setup(e, s)
	defer send()

	s.Logout()
	resp.Error = nil
}
