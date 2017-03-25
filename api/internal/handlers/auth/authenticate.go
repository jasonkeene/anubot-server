package auth

import "github.com/jasonkeene/anubot-server/api/internal/handlers"

// UserAuthenticator authenticates existing users.
type UserAuthenticator interface {
	AuthenticateUser(username, password string) (userID string, authenticated bool, err error)
}

// AuthenticateHandler authenticates the session.
type AuthenticateHandler struct {
	auth UserAuthenticator
}

// NewAuthenticateHandler returns a new AuthenticateHandler.
func NewAuthenticateHandler(auth UserAuthenticator) *AuthenticateHandler {
	return &AuthenticateHandler{
		auth: auth,
	}
}

// HandleEvent responds to a websocket event.
func (h *AuthenticateHandler) HandleEvent(e handlers.Event, s handlers.Session) {
	resp, send := handlers.Setup(e, s)
	defer send()

	ok, payload := h.validatePayload(e.Payload)
	if !ok {
		resp.Error = handlers.InvalidPayload
		return
	}

	id, ok, err := h.auth.AuthenticateUser(payload.username, payload.password)
	if !ok || err != nil {
		resp.Error = handlers.AuthenticationError
		return
	}

	s.SetAuthentication(id)
	resp.Error = nil
}

// authenticationPayload represents the payload that should be sent when
// authenticating a user.
type authenticationPayload struct {
	username string
	password string
}

// validatePayload returns true if the payload is valid.
func (h *AuthenticateHandler) validatePayload(p interface{}) (bool, authenticationPayload) {
	if p == nil {
		return false, authenticationPayload{}
	}
	payload, ok := p.(map[string]interface{})
	if !ok {
		return false, authenticationPayload{}
	}
	username, ok := payload["username"].(string)
	if !ok {
		return false, authenticationPayload{}
	}
	password, ok := payload["password"].(string)
	if !ok {
		return false, authenticationPayload{}
	}
	return validCredentials(username, password), authenticationPayload{
		username: username,
		password: password,
	}
}
