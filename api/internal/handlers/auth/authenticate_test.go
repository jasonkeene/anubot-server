package auth_test

import (
	"errors"
	"testing"

	"github.com/a8m/expect"
	"github.com/jasonkeene/anubot-server/api/internal/handlers"
	"github.com/jasonkeene/anubot-server/api/internal/handlers/auth"
)

func TestAuthentication(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{}
	spyAuthenticator := &SpyAuthenticator{
		userID:        "test-user-id",
		authenticated: true,
	}
	event := handlers.Event{
		Cmd: "authenticate",
		Payload: map[string]interface{}{
			"username": "test-username",
			"password": "test-password",
		},
		RequestID: "test-request-id",
	}
	handler := auth.NewAuthenticateHandler(spyAuthenticator)

	handler.HandleEvent(event, spySession)

	expected := handlers.Event{
		Cmd:       "authenticate",
		RequestID: "test-request-id",
	}
	expect(spyAuthenticator.username).To.Equal("test-username")
	expect(spyAuthenticator.password).To.Equal("test-password")
	expect(spySession.setAuthenticationCalledWith).To.Equal("test-user-id")
	expect(spySession.sendCalledWith).To.Equal(expected)
}

func TestFailedAuthentication(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{}
	spyAuthenticator := &SpyAuthenticator{
		authenticated: false,
	}
	event := handlers.Event{
		Cmd: "authenticate",
		Payload: map[string]interface{}{
			"username": "test-username",
			"password": "test-password",
		},
		RequestID: "test-request-id",
	}
	handler := auth.NewAuthenticateHandler(spyAuthenticator)

	handler.HandleEvent(event, spySession)

	expected := handlers.Event{
		Cmd:       "authenticate",
		RequestID: "test-request-id",
		Error:     handlers.AuthenticationError,
	}
	expect(spyAuthenticator.username).To.Equal("test-username")
	expect(spyAuthenticator.password).To.Equal("test-password")
	expect(spySession.sendCalledWith).To.Equal(expected)
}

func TestAuthenticationError(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{}
	spyAuthenticator := &SpyAuthenticator{
		authenticated: true,
		err:           errors.New("test-error"),
	}
	event := handlers.Event{
		Cmd: "authenticate",
		Payload: map[string]interface{}{
			"username": "test-username",
			"password": "test-password",
		},
		RequestID: "test-request-id",
	}
	handler := auth.NewAuthenticateHandler(spyAuthenticator)

	handler.HandleEvent(event, spySession)

	expected := handlers.Event{
		Cmd:       "authenticate",
		RequestID: "test-request-id",
		Error:     handlers.AuthenticationError,
	}
	expect(spyAuthenticator.username).To.Equal("test-username")
	expect(spyAuthenticator.password).To.Equal("test-password")
	expect(spySession.sendCalledWith).To.Equal(expected)
}

func TestInvalidAuthenticationRequest(t *testing.T) {
	expect := expect.New(t)

	cases := map[string]interface{}{
		"empty payload":     nil,
		"invalid structure": struct{}{},
		"username is empty": map[string]interface{}{
			"username": "",
			"password": "test-password",
		},
		"username not string": map[string]interface{}{
			"username": 1234,
			"password": "test-password",
		},
		"password is empty": map[string]interface{}{
			"username": "test-username",
			"password": "",
		},
		"password not string": map[string]interface{}{
			"username": "test-username",
			"password": 1234,
		},
	}

	for _, payload := range cases {
		spySession := &SpySession{}
		spyAuthenticator := &SpyAuthenticator{}
		event := handlers.Event{
			Cmd:       "authenticate",
			Payload:   payload,
			RequestID: "test-request-id",
		}
		handler := auth.NewAuthenticateHandler(spyAuthenticator)

		handler.HandleEvent(event, spySession)

		expected := handlers.Event{
			Cmd:       "authenticate",
			RequestID: "test-request-id",
			Error:     handlers.InvalidPayload,
		}
		expect(spySession.sendCalledWith).To.Equal(expected)
	}
}
