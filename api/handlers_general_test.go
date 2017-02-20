package api_test

import (
	"testing"

	"github.com/a8m/expect"
)

func TestPingHandler(t *testing.T) {
	expect := expect.New(t)

	server, cleanup := setupServer()
	defer cleanup()
	client, cleanup := setupClient(server.url)
	defer cleanup()

	pingReq := event{
		Cmd:       "ping",
		RequestID: requestID(),
	}
	expectedResp := event{
		Cmd:       "pong",
		RequestID: pingReq.RequestID,
	}
	client.SendEvent(pingReq)
	expect(client.ReadEvent()).To.Equal(expectedResp)
}

func TestMethodsHandler(t *testing.T) {
	expect := expect.New(t)

	server, cleanup := setupServer()
	defer cleanup()
	client, cleanup := setupClient(server.url)
	defer cleanup()

	methodsReq := event{
		Cmd:       "methods",
		RequestID: requestID(),
	}
	expectedResp := event{
		Cmd:       "methods",
		RequestID: methodsReq.RequestID,
		Payload: []string{
			"authenticate",
			"bttv-emoji",
			"logout",
			"methods",
			"ping",
			"register",
			"twitch-clear-auth",
			"twitch-games",
			"twitch-oauth-start",
			"twitch-send-message",
			"twitch-stream-messages",
			"twitch-update-chat-description",
			"twitch-user-details",
		},
	}
	client.SendEvent(methodsReq)
	e := client.ReadEvent()
	expect(e.Cmd).To.Equal(expectedResp.Cmd)
	expect(e.Error).To.Equal(expectedResp.Error)
	expect(e.RequestID).To.Equal(expectedResp.RequestID)

	payload := e.Payload.([]interface{})
	expect(len(payload)).To.Equal(len(expectedResp.Payload.([]string)))
	for i, m := range payload {
		expect(m.(string)).To.Equal(expectedResp.Payload.([]string)[i])
	}
}
