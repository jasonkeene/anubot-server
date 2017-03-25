package auth_test

import "github.com/jasonkeene/anubot-server/api/internal/handlers"

type SpyAuthenticator struct {
	userID        string
	authenticated bool
	err           error
	username      string
	password      string
}

func (s *SpyAuthenticator) AuthenticateUser(
	username string,
	password string,
) (userID string, authenticated bool, err error) {
	s.username = username
	s.password = password
	return s.userID, s.authenticated, s.err
}

type SpySession struct {
	setAuthenticationCalledWith string
	sendCalledWith              handlers.Event
	logoutCalled                bool
	userID                      string
	authenticated               bool
}

func (s *SpySession) SetAuthentication(userID string) {
	s.setAuthenticationCalledWith = userID
}

func (s *SpySession) Send(e handlers.Event) error {
	s.sendCalledWith = e
	return nil
}

func (s *SpySession) Logout() {
	s.logoutCalled = true
}

func (s *SpySession) Authenticated() (userID string, authenticated bool) {
	return s.userID, s.authenticated
}

type SpyRegistrar struct {
	userID   string
	err      error
	username string
	password string
}

func (s *SpyRegistrar) RegisterUser(username, password string) (userID string, err error) {
	s.username = username
	s.password = password
	return s.userID, s.err
}

type SpyHandler struct {
	called bool
}

func (s *SpyHandler) HandleEvent(handlers.Event, handlers.Session) {
	println("i was called!")
	s.called = true
}
