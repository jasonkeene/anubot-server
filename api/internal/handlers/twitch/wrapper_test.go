package twitch_test

import (
	"errors"
	"testing"

	"github.com/a8m/expect"
	"github.com/jasonkeene/anubot-server/api/internal/handlers"
	"github.com/jasonkeene/anubot-server/api/internal/handlers/twitch"
	"github.com/jasonkeene/anubot-server/store"
)

func TestAuthenticatedUser(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{
		userID: "test-user-id",
	}
	spyCredsProvider := &SpyCredentialsProvider{
		creds: store.TwitchCredentials{
			StreamerAuthenticated: true,
			BotAuthenticated:      true,
		},
	}
	spyHandler := &SpyHandler{}
	wrapped := twitch.AuthenticateWrapper(spyCredsProvider, spyHandler)
	event := handlers.Event{
		Cmd:       "test-command",
		RequestID: "test-request-id",
	}

	wrapped.HandleEvent(event, spySession)

	expect(spyCredsProvider.calledWith).To.Equal("test-user-id")
	expect(spyHandler.called).To.Be.True()
}

func TestUnauthenticatedUser(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{}
	spyCredsProvider := &SpyCredentialsProvider{}
	spyHandler := &SpyHandler{}
	wrapped := twitch.AuthenticateWrapper(spyCredsProvider, spyHandler)
	event := handlers.Event{
		Cmd:       "test-command",
		RequestID: "test-request-id",
	}

	wrapped.HandleEvent(event, spySession)

	expected := handlers.Event{
		Cmd:       "test-command",
		RequestID: "test-request-id",
		Error:     handlers.TwitchAuthenticationError,
	}
	expect(spyHandler.called).To.Be.False()
	expect(spySession.sendCalledWith).To.Equal(expected)
}

func TestErrorWhenGettingCredentials(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{}
	spyCredsProvider := &SpyCredentialsProvider{
		err: errors.New("test-error"),
	}
	spyHandler := &SpyHandler{}
	wrapped := twitch.AuthenticateWrapper(spyCredsProvider, spyHandler)
	event := handlers.Event{
		Cmd:       "test-command",
		RequestID: "test-request-id",
	}

	wrapped.HandleEvent(event, spySession)

	expected := handlers.Event{
		Cmd:       "test-command",
		RequestID: "test-request-id",
		Error:     handlers.UnknownError,
	}
	expect(spyHandler.called).To.Be.False()
	expect(spySession.sendCalledWith).To.Equal(expected)
}
