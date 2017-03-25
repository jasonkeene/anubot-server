package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/a8m/expect"
	"github.com/gorilla/websocket"
	"github.com/jasonkeene/anubot-server/api"
	"github.com/jasonkeene/anubot-server/api/internal/handlers"
	"github.com/jasonkeene/anubot-server/store"
)

func TestItRepondsToValidMessages(t *testing.T) {
	expect := expect.New(t)

	server := httptest.NewServer(api.New(nil, nil, nil, nil, "", ""))
	defer server.Close()

	url := strings.Replace(server.URL, "http://", "ws://", 1)
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	expect(err).To.Be.Nil()
	defer func() {
		_ = c.Close()
	}()

	event := handlers.Event{
		Cmd:       "ping",
		RequestID: "test-request-id",
	}
	ping, err := json.Marshal(event)
	expect(err).To.Be.Nil()
	err = c.WriteMessage(websocket.TextMessage, ping)
	expect(err).To.Be.Nil()

	_, message, err := c.ReadMessage()
	expect(err).To.Be.Nil()
	var actual handlers.Event
	err = json.Unmarshal(message, &actual)
	expect(err).To.Be.Nil()

	expected := handlers.Event{
		Cmd:       "pong",
		RequestID: "test-request-id",
	}
	expect(actual).To.Equal(expected)
}

func TestItRepondsToInvalidMessages(t *testing.T) {
	expect := expect.New(t)

	server := httptest.NewServer(api.New(nil, nil, nil, nil, "", ""))
	defer server.Close()

	url := strings.Replace(server.URL, "http://", "ws://", 1)
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	expect(err).To.Be.Nil()
	defer func() {
		_ = c.Close()
	}()

	event := handlers.Event{
		Cmd:       "foobarbaz",
		RequestID: "test-request-id",
	}
	badMessage, err := json.Marshal(event)
	expect(err).To.Be.Nil()
	err = c.WriteMessage(websocket.TextMessage, badMessage)
	expect(err).To.Be.Nil()

	_, message, err := c.ReadMessage()
	expect(err).To.Be.Nil()
	var actual handlers.Event
	err = json.Unmarshal(message, &actual)
	expect(err).To.Be.Nil()

	expected := handlers.Event{
		Cmd:       "foobarbaz",
		RequestID: "test-request-id",
		Error:     handlers.InvalidCommand,
	}
	expect(actual).To.Equal(expected)
}

func TestItChecksOriginAndHostMatch(t *testing.T) {
	expect := expect.New(t)

	server := httptest.NewServer(api.New(nil, nil, nil, nil, "", ""))
	defer server.Close()

	url := strings.Replace(server.URL, "http://", "ws://", 1)
	headers := http.Header{}
	headers.Set("origin", "bad-origin")
	_, _, err := websocket.DefaultDialer.Dial(url, headers)
	expect(err).Not.To.Be.Nil()
}

func TestItPingsPeriodically(t *testing.T) {
	expect := expect.New(t)

	api := api.New(nil, nil, nil, nil, "", "", api.WithPingInterval(time.Millisecond))
	server := httptest.NewServer(api)
	defer server.Close()

	url := strings.Replace(server.URL, "http://", "ws://", 1)
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	expect(err).To.Be.Nil()
	defer func() {
		_ = c.Close()
	}()

	gotPing := make(chan struct{})
	var once sync.Once
	oldHandler := c.PingHandler()
	c.SetPingHandler(func(data string) error {
		once.Do(func() {
			close(gotPing)
		})
		return oldHandler(data)
	})

	event := handlers.Event{
		Cmd:       "foobarbaz",
		RequestID: "test-request-id",
	}
	badMessage, err := json.Marshal(event)
	expect(err).To.Be.Nil()
	err = c.WriteMessage(websocket.TextMessage, badMessage)
	expect(err).To.Be.Nil()

	_, _, err = c.ReadMessage()
	expect(err).To.Be.Nil()

	select {
	case <-gotPing:
	case <-time.After(5 * time.Second):
		t.Error("did not get ping after 5 second")
	}
}

func TestItWiresUpUnauthenticatedHandlers(t *testing.T) {
	expect := expect.New(t)

	api := api.New(nil, nil, nil, nil, "", "")
	server := httptest.NewServer(api)
	defer server.Close()

	url := strings.Replace(server.URL, "http://", "ws://", 1)
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	expect(err).To.Be.Nil()
	defer func() {
		_ = c.Close()
	}()

	cases := []string{
		"ping",
		"methods",
		"register",
		"authenticate",
		"logout",
	}
	for _, method := range cases {
		event := handlers.Event{
			Cmd:       method,
			RequestID: "test-request-id",
		}
		bytes, err := json.Marshal(event)
		expect(err).To.Be.Nil()
		err = c.WriteMessage(websocket.TextMessage, bytes)
		expect(err).To.Be.Nil()

		_, resp, err := c.ReadMessage()
		expect(err).To.Be.Nil()
		var actual handlers.Event
		err = json.Unmarshal(resp, &actual)
		expect(err).To.Be.Nil()

		expect(actual.Error).Not.To.Equal(handlers.AuthenticationError)
	}
}

func TestItWiresUpAuthenticatedHandlers(t *testing.T) {
	expect := expect.New(t)

	spyStore := &SpyStore{}
	spyTwitchClient := &SpyTwitchClient{}
	api := api.New(nil, spyStore, spyTwitchClient, nil, "", "")
	server := httptest.NewServer(api)
	defer server.Close()

	url := strings.Replace(server.URL, "http://", "ws://", 1)
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	expect(err).To.Be.Nil()
	defer func() {
		_ = c.Close()
	}()

	cases := []string{
		"twitch-oauth-start",
		"twitch-clear-auth",
		"twitch-user-details",
		"twitch-games",
		"bttv-emoji",
		"twitch-stream-messages",
		"twitch-send-message",
		"twitch-update-chat-description",
	}
	for _, method := range cases {
		event := handlers.Event{
			Cmd:       method,
			RequestID: "test-request-id",
		}
		bytes, err := json.Marshal(event)
		expect(err).To.Be.Nil()
		err = c.WriteMessage(websocket.TextMessage, bytes)
		expect(err).To.Be.Nil()

		_, resp, err := c.ReadMessage()
		expect(err).To.Be.Nil()
		var actual handlers.Event
		err = json.Unmarshal(resp, &actual)
		expect(err).To.Be.Nil()

		expect(actual.Error).To.Equal(handlers.AuthenticationError)
	}

	spyStore.userID = "test-user-id"
	spyStore.authenticated = true
	event := handlers.Event{
		Cmd:       "authenticate",
		RequestID: "test-request-id",
		Payload: map[string]interface{}{
			"username": "test-username",
			"password": "test-password",
		},
	}
	bytes, err := json.Marshal(event)
	expect(err).To.Be.Nil()
	err = c.WriteMessage(websocket.TextMessage, bytes)
	expect(err).To.Be.Nil()

	_, _, err = c.ReadMessage()
	expect(err).To.Be.Nil()

	for _, method := range cases {
		event := handlers.Event{
			Cmd:       method,
			RequestID: "test-request-id",
		}
		bytes, err := json.Marshal(event)
		expect(err).To.Be.Nil()
		err = c.WriteMessage(websocket.TextMessage, bytes)
		expect(err).To.Be.Nil()

		_, resp, err := c.ReadMessage()
		expect(err).To.Be.Nil()
		var actual handlers.Event
		err = json.Unmarshal(resp, &actual)
		expect(err).To.Be.Nil()

		expect(actual.Error).Not.To.Equal(handlers.AuthenticationError)
	}
}

func TestItWiresUpTwitchAuthenticatedHandlers(t *testing.T) {
	expect := expect.New(t)

	spyStreamManager := &SpyStreamManager{}
	spyStore := &SpyStore{}
	spyTwitchClient := &SpyTwitchClient{}
	api := api.New(spyStreamManager, spyStore, spyTwitchClient, nil, "", "")
	server := httptest.NewServer(api)
	defer server.Close()

	url := strings.Replace(server.URL, "http://", "ws://", 1)
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	expect(err).To.Be.Nil()
	defer func() {
		_ = c.Close()
	}()

	spyStore.userID = "test-user-id"
	spyStore.authenticated = true
	event := handlers.Event{
		Cmd:       "authenticate",
		RequestID: "test-request-id",
		Payload: map[string]interface{}{
			"username": "test-username",
			"password": "test-password",
		},
	}
	bytes, err := json.Marshal(event)
	expect(err).To.Be.Nil()
	err = c.WriteMessage(websocket.TextMessage, bytes)
	expect(err).To.Be.Nil()

	_, _, err = c.ReadMessage()
	expect(err).To.Be.Nil()

	cases := []string{
		"twitch-send-message",
		"twitch-update-chat-description",
		"twitch-stream-messages",
	}
	for _, method := range cases {
		event := handlers.Event{
			Cmd:       method,
			RequestID: "test-request-id",
		}
		bytes, err := json.Marshal(event)
		expect(err).To.Be.Nil()
		err = c.WriteMessage(websocket.TextMessage, bytes)
		expect(err).To.Be.Nil()

		_, resp, err := c.ReadMessage()
		expect(err).To.Be.Nil()
		var actual handlers.Event
		err = json.Unmarshal(resp, &actual)
		expect(err).To.Be.Nil()

		expect(actual.Error).To.Equal(handlers.TwitchAuthenticationError)
	}

	spyStore.creds = store.TwitchCredentials{
		StreamerAuthenticated: true,
		BotAuthenticated:      true,
	}

	for _, method := range cases {
		event := handlers.Event{
			Cmd:       method,
			RequestID: "test-request-id",
		}
		bytes, err := json.Marshal(event)
		expect(err).To.Be.Nil()
		err = c.WriteMessage(websocket.TextMessage, bytes)
		expect(err).To.Be.Nil()

		_, resp, err := c.ReadMessage()
		expect(err).To.Be.Nil()
		var actual handlers.Event
		err = json.Unmarshal(resp, &actual)
		expect(err).To.Be.Nil()

		expect(actual.Error).Not.To.Equal(handlers.TwitchAuthenticationError)
	}
}
