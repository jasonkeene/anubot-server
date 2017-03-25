package auth

import (
	"github.com/jasonkeene/anubot-server/api/internal/handlers"
	"github.com/jasonkeene/anubot-server/store"
)

// UserRegistrar registers new users.
type UserRegistrar interface {
	RegisterUser(username, password string) (userID string, err error)
}

// RegisterHandler registers new users and authenticates the session.
type RegisterHandler struct {
	registrar UserRegistrar
}

// NewRegisterHandler returns a new RegisterHandler.
func NewRegisterHandler(registrar UserRegistrar) *RegisterHandler {
	return &RegisterHandler{
		registrar: registrar,
	}
}

// HandleEvent registers new users and authenticates the session.
func (h *RegisterHandler) HandleEvent(e handlers.Event, s handlers.Session) {
	resp, send := handlers.Setup(e, s)
	defer send()

	ok, payload := h.validatePayload(e.Payload)
	if !ok {
		resp.Error = handlers.InvalidPayload
		return
	}

	userID, err := h.registrar.RegisterUser(payload.username, payload.password)
	if err != nil {
		if err == store.ErrUsernameTaken {
			resp.Error = handlers.UsernameTaken
		}
		return
	}
	s.SetAuthentication(userID)
	resp.Error = nil
}

// registrationPayload represents the payload that should be sent when
// registering a new user.
type registrationPayload struct {
	username string
	password string
}

// validatePayload returns true if the payload is valid.
func (h *RegisterHandler) validatePayload(p interface{}) (bool, registrationPayload) {
	if p == nil {
		return false, registrationPayload{}
	}
	payload, ok := p.(map[string]interface{})
	if !ok {
		return false, registrationPayload{}
	}
	username, ok := payload["username"].(string)
	if !ok {
		return false, registrationPayload{}
	}
	password, ok := payload["password"].(string)
	if !ok {
		return false, registrationPayload{}
	}
	return validCredentials(username, password), registrationPayload{
		username: username,
		password: password,
	}
}
