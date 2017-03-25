package auth_test

import (
	"testing"

	"github.com/a8m/expect"
	"github.com/jasonkeene/anubot-server/api/internal/handlers"
	"github.com/jasonkeene/anubot-server/api/internal/handlers/auth"
)

func TestSessionIsCleared(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{}
	event := handlers.Event{
		Cmd:       "logout",
		RequestID: "test-request-id",
	}

	auth.LogoutHandler(event, spySession)

	expected := handlers.Event{
		Cmd:       "logout",
		RequestID: "test-request-id",
	}
	expect(spySession.logoutCalled).To.Be.True()
	expect(spySession.sendCalledWith).To.Equal(expected)
}
