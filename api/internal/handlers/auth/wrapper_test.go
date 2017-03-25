package auth_test

import (
	"testing"

	"github.com/a8m/expect"
	"github.com/jasonkeene/anubot-server/api/internal/handlers"
	"github.com/jasonkeene/anubot-server/api/internal/handlers/auth"
)

func TestAuthenticatedUser(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{
		userID:        "test-user-id",
		authenticated: true,
	}
	spyHandler := &SpyHandler{}
	wrappedHandler := auth.AuthenticateWrapper(spyHandler)
	event := handlers.Event{
		Cmd:       "test-cmd",
		RequestID: "test-request-id",
	}

	wrappedHandler.HandleEvent(event, spySession)

	expect(spyHandler.called).To.Be.True()
}

func TestUnauthenticatedUser(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{
		authenticated: false,
	}
	spyHandler := &SpyHandler{}
	wrappedHandler := auth.AuthenticateWrapper(spyHandler)
	event := handlers.Event{
		Cmd:       "test-cmd",
		RequestID: "test-request-id",
	}

	wrappedHandler.HandleEvent(event, spySession)

	expected := handlers.Event{
		Cmd:       "test-cmd",
		RequestID: "test-request-id",
		Error:     handlers.AuthenticationError,
	}
	expect(spyHandler.called).To.Be.False()
	expect(spySession.sendCalledWith).To.Equal(expected)
}
