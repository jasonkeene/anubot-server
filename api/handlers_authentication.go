package api

import (
	"log"

	"github.com/jasonkeene/anubot-server/store"
)

// registerHandler accepts registration information and authenticates the
// session.
func registerHandler(e event, s *session) {
	resp, send := setup(e, s)
	defer send()

	ok, payload := validCredentialsPayload(e.Payload)
	if !ok {
		resp.Error = invalidPayload
		return
	}

	id, err := s.api.store.RegisterUser(payload.Username, payload.Password)
	if err != nil {
		if err == store.ErrUsernameTaken {
			resp.Error = usernameTaken
		}
		return
	}
	s.SetAuthentication(id)
	resp.Error = nil
}

// authenticateHandler authenticates the session.
func authenticateHandler(e event, s *session) {
	resp, send := setup(e, s)
	defer send()

	ok, payload := validCredentialsPayload(e.Payload)
	if !ok {
		resp.Error = invalidPayload
		return
	}

	id, ok, err := s.api.store.AuthenticateUser(payload.Username, payload.Password)
	if !ok || err != nil {
		resp.Error = authenticationError
		return
	}

	s.SetAuthentication(id)
	resp.Error = nil
}

// logoutHandler clears the authentication for the session.
func logoutHandler(e event, s *session) {
	resp, send := setup(e, s)
	defer send()

	s.Logout()
	resp.Error = nil
}

// authenticateWrapper wraps a handler and makes sure the session is
// authenticated.
func authenticateWrapper(f eventHandler) eventHandler {
	return func(e event, s *session) {
		_, ok := s.Authenticated()
		if ok {
			f(e, s)
			return
		}
		err := s.Send(event{
			Cmd:       e.Cmd,
			RequestID: e.RequestID,
			Error:     authenticationError,
		})
		if err != nil {
			log.Printf("unable to tx: %s", err)
		}
	}
}

// validCredentials validates that the credentials follow some sane rules.
func validCredentials(username, password string) bool {
	return username != "" && password != ""
}

type registrationPayload struct {
	Username string
	Password string
}

// validCredentialsPayload returns true if the payload is valid.
func validCredentialsPayload(p interface{}) (bool, registrationPayload) {
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
		Username: username,
		Password: password,
	}
}
