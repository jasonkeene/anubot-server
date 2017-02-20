package api_test

import (
	"errors"
	"testing"

	"github.com/a8m/expect"
	"github.com/jasonkeene/anubot-server/store"
)

func TestRegistration(t *testing.T) {
	expect := expect.New(t)

	server, cleanup := setupServer()
	defer cleanup()
	client, cleanup := setupClient(server.url)
	defer cleanup()

	server.mockStore.RegisterUserOutput.UserID <- "some-user-id"
	server.mockStore.RegisterUserOutput.Err <- nil
	registerReq := event{
		Cmd:       "register",
		RequestID: requestID(),
		Payload: map[string]string{
			"username": "foo",
			"password": "bar",
		},
	}
	expectedResp := event{
		Cmd:       "register",
		RequestID: registerReq.RequestID,
	}
	client.SendEvent(registerReq)
	expect(client.ReadEvent()).To.Equal(expectedResp)
}

func TestDuplicateRegistration(t *testing.T) {
	expect := expect.New(t)

	server, cleanup := setupServer()
	defer cleanup()
	client, cleanup := setupClient(server.url)
	defer cleanup()

	server.mockStore.RegisterUserOutput.UserID <- ""
	server.mockStore.RegisterUserOutput.Err <- store.ErrUsernameTaken
	registerReq := event{
		Cmd:       "register",
		RequestID: requestID(),
		Payload: map[string]string{
			"username": "foo",
			"password": "bar",
		},
	}
	expectedResp := event{
		Cmd:       "register",
		RequestID: registerReq.RequestID,
		Error: &eventErr{
			Code: 3,
			Text: "username has already been taken",
		},
	}
	client.SendEvent(registerReq)
	expect(client.ReadEvent()).To.Equal(expectedResp)
}

func TestInvalidRegistrationPayloads(t *testing.T) {
	cases := make(map[string]interface{})
	cases["nil"] = nil
	cases["invalid payload type"] = []string{}
	invalidUsernameType := make(map[string]interface{})
	invalidUsernameType["username"] = 1
	invalidUsernameType["password"] = "password"
	cases["invalid username type"] = invalidUsernameType
	invalidPasswordType := make(map[string]interface{})
	invalidPasswordType["username"] = "username"
	invalidPasswordType["password"] = 1
	cases["invalid password type"] = invalidPasswordType
	cases["empty username"] = map[string]string{
		"username": "",
		"password": "password",
	}
	cases["empty password"] = map[string]string{
		"username": "username",
		"password": "",
	}

	expect := expect.New(t)

	server, cleanup := setupServer()
	defer cleanup()
	client, cleanup := setupClient(server.url)
	defer cleanup()

	for _, tc := range cases {
		registerReq := event{
			Cmd:       "register",
			RequestID: requestID(),
			Payload:   tc,
		}
		expectedResp := event{
			Cmd:       "register",
			RequestID: registerReq.RequestID,
			Error: &eventErr{
				Code: 2,
				Text: "invalid payload data",
			},
		}
		client.SendEvent(registerReq)
		expect(client.ReadEvent()).To.Equal(expectedResp)
	}
}

func TestAuthentication(t *testing.T) {
	cases := map[string]struct {
		Authenticated bool
		Err           error
		Error         *eventErr
	}{
		"authenticated passed with no error": {
			Authenticated: true,
			Err:           nil,
			Error:         nil,
		},
		"authenticated did not pass and there was no error": {
			Authenticated: false,
			Err:           nil,
			Error: &eventErr{
				Code: 4,
				Text: "authentication error",
			},
		},
		"authenticated passed but with an error": {
			Authenticated: true,
			Err:           errors.New("some error"),
			Error: &eventErr{
				Code: 4,
				Text: "authentication error",
			},
		},
		"authenticated did not pass and there was an error": {
			Authenticated: false,
			Err:           errors.New("some error"),
			Error: &eventErr{
				Code: 4,
				Text: "authentication error",
			},
		},
	}

	expect := expect.New(t)

	server, cleanup := setupServer()
	defer cleanup()
	client, cleanup := setupClient(server.url)
	defer cleanup()

	for _, tc := range cases {
		server.mockStore.AuthenticateUserOutput.UserID <- "user-id"
		server.mockStore.AuthenticateUserOutput.Authenticated <- tc.Authenticated
		server.mockStore.AuthenticateUserOutput.Err <- tc.Err
		authenticateReq := event{
			Cmd:       "authenticate",
			RequestID: requestID(),
			Payload: map[string]string{
				"username": "username",
				"password": "password",
			},
		}
		expectedResp := event{
			Cmd:       "authenticate",
			RequestID: authenticateReq.RequestID,
			Error:     tc.Error,
		}
		client.SendEvent(authenticateReq)
		expect(client.ReadEvent()).To.Equal(expectedResp)
	}
}

func TestInvalidAuthenticationPayloads(t *testing.T) {
	cases := make(map[string]interface{})
	cases["nil"] = nil
	cases["invalid payload type"] = []string{}
	invalidUsernameType := make(map[string]interface{})
	invalidUsernameType["username"] = 1
	invalidUsernameType["password"] = "password"
	cases["invalid username type"] = invalidUsernameType
	invalidPasswordType := make(map[string]interface{})
	invalidPasswordType["username"] = "username"
	invalidPasswordType["password"] = 1
	cases["invalid password type"] = invalidPasswordType
	cases["empty username"] = map[string]string{
		"username": "",
		"password": "password",
	}
	cases["empty password"] = map[string]string{
		"username": "username",
		"password": "",
	}

	expect := expect.New(t)

	server, cleanup := setupServer()
	defer cleanup()
	client, cleanup := setupClient(server.url)
	defer cleanup()

	for _, tc := range cases {
		authenticateReq := event{
			Cmd:       "authenticate",
			RequestID: requestID(),
			Payload:   tc,
		}
		expectedResp := event{
			Cmd:       "authenticate",
			RequestID: authenticateReq.RequestID,
			Error: &eventErr{
				Code: 2,
				Text: "invalid payload data",
			},
		}
		client.SendEvent(authenticateReq)
		expect(client.ReadEvent()).To.Equal(expectedResp)
	}
}

func TestLogout(t *testing.T) {
	expect := expect.New(t)

	server, cleanup := setupServer()
	defer cleanup()
	client, cleanup := setupClient(server.url)
	defer cleanup()

	logoutReq := event{
		Cmd:       "logout",
		RequestID: requestID(),
	}
	client.SendEvent(logoutReq)
	expect(client.ReadEvent()).To.Equal(logoutReq)
}
