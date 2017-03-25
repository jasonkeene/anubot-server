package auth_test

import (
	"errors"
	"testing"

	"github.com/a8m/expect"
	"github.com/jasonkeene/anubot-server/api/internal/handlers"
	"github.com/jasonkeene/anubot-server/api/internal/handlers/auth"
	"github.com/jasonkeene/anubot-server/store"
)

func TestRegistration(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{}
	spyRegistrar := &SpyRegistrar{
		userID: "test-user-id",
	}
	event := handlers.Event{
		Cmd: "register",
		Payload: map[string]interface{}{
			"username": "test-username",
			"password": "test-password",
		},
		RequestID: "test-request-id",
	}
	handler := auth.NewRegisterHandler(spyRegistrar)

	handler.HandleEvent(event, spySession)

	expected := handlers.Event{
		Cmd:       "register",
		RequestID: "test-request-id",
	}
	expect(spyRegistrar.username).To.Equal("test-username")
	expect(spyRegistrar.password).To.Equal("test-password")
	expect(spySession.setAuthenticationCalledWith).To.Equal("test-user-id")
	expect(spySession.sendCalledWith).To.Equal(expected)
}

func TestErrorWithRegistration(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{}
	spyRegistrar := &SpyRegistrar{
		err: errors.New("test-error"),
	}
	event := handlers.Event{
		Cmd: "register",
		Payload: map[string]interface{}{
			"username": "test-username",
			"password": "test-password",
		},
		RequestID: "test-request-id",
	}
	handler := auth.NewRegisterHandler(spyRegistrar)

	handler.HandleEvent(event, spySession)

	expected := handlers.Event{
		Cmd:       "register",
		RequestID: "test-request-id",
		Error:     handlers.UnknownError,
	}
	expect(spySession.sendCalledWith).To.Equal(expected)
}

func TestRegisteringAUsernameThatAlreadyExists(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{}
	spyRegistrar := &SpyRegistrar{
		err: store.ErrUsernameTaken,
	}
	event := handlers.Event{
		Cmd: "register",
		Payload: map[string]interface{}{
			"username": "test-username",
			"password": "test-password",
		},
		RequestID: "test-request-id",
	}
	handler := auth.NewRegisterHandler(spyRegistrar)

	handler.HandleEvent(event, spySession)

	expected := handlers.Event{
		Cmd:       "register",
		RequestID: "test-request-id",
		Error:     handlers.UsernameTaken,
	}
	expect(spySession.sendCalledWith).To.Equal(expected)
}

func TestInvalidRegisterRequest(t *testing.T) {
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
		spyRegistrar := &SpyRegistrar{}
		event := handlers.Event{
			Cmd:       "register",
			Payload:   payload,
			RequestID: "test-request-id",
		}
		handler := auth.NewRegisterHandler(spyRegistrar)

		handler.HandleEvent(event, spySession)

		expected := handlers.Event{
			Cmd:       "register",
			RequestID: "test-request-id",
			Error:     handlers.InvalidPayload,
		}
		expect(spySession.sendCalledWith).To.Equal(expected)
	}
}
