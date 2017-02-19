package api

import (
	"log"

	"github.com/jasonkeene/anubot-server/store"
)

// registerHandler accepts registration information and authenticates the
// session.
func registerHandler(e event, s *session) {
	// setup default response event
	resp := event{
		Cmd:       "register",
		RequestID: e.RequestID,
		Error:     unknownError,
	}
	defer func() {
		err := s.Send(resp)
		if err != nil {
			log.Printf("unable to tx: %s", err)
		}
	}()

	// validate event
	if e.Payload == nil {
		resp.Error = invalidPayload
		return
	}
	payload, ok := e.Payload.(map[string]interface{})
	if !ok {
		resp.Error = invalidPayload
		return
	}
	username, ok := payload["username"].(string)
	if !ok {
		resp.Error = invalidPayload
		return
	}
	password, ok := payload["password"].(string)
	if !ok {
		resp.Error = invalidPayload
		return
	}
	if !validCredentials(username, password) {
		resp.Error = invalidPayload
		return
	}

	// attempt to register the user
	id, err := s.Store().RegisterUser(username, password)
	if err != nil {
		if err == store.ErrUsernameTaken {
			resp.Error = usernameTaken
		}
		return
	}
	s.SetAuthentication(id)
	resp.Error = nil
	return
}

// authenticateHandler authenticates the session.
func authenticateHandler(e event, s *session) {
	// setup default response event
	resp := event{
		Cmd:       "authenticate",
		RequestID: e.RequestID,
		Error:     unknownError,
	}
	defer func() {
		err := s.Send(resp)
		if err != nil {
			log.Printf("unable to tx: %s", err)
		}
	}()

	// validate event
	if e.Payload == nil {
		resp.Error = invalidPayload
		return
	}
	payload, ok := e.Payload.(map[string]interface{})
	if !ok {
		resp.Error = invalidPayload
		return
	}
	username, ok := payload["username"].(string)
	if !ok {
		resp.Error = invalidPayload
		return
	}
	password, ok := payload["password"].(string)
	if !ok {
		resp.Error = invalidPayload
		return
	}
	if !validCredentials(username, password) {
		resp.Error = invalidPayload
		return
	}

	// attempt to authenticate the user
	id, ok := s.Store().AuthenticateUser(username, password)
	if !ok {
		resp.Error = authenticationError
		return
	}

	s.SetAuthentication(id)
	resp.Error = nil
	return
}

// logoutHandler clears the authentication for the session.
func logoutHandler(e event, s *session) {
	s.Logout()
	err := s.Send(event{
		Cmd:       "logout",
		RequestID: e.RequestID,
	})
	if err != nil {
		log.Printf("unable to tx: %s", err)
	}
}

// authenticateWrapper wraps a handler and makes sure the session is
// authenticated.
func authenticateWrapper(f handlerFunc) handlerFunc {
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
