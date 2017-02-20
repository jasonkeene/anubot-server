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
)

func TestItRepondsToValidMessages(t *testing.T) {
	expect := expect.New(t)

	server := httptest.NewServer(api.New(nil, nil, nil, ""))
	defer server.Close()

	url := strings.Replace(server.URL, "http://", "ws://", 1)
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	expect(err).To.Be.Nil()
	defer c.Close()

	ping := []byte(`{"cmd":"ping"}`)
	err = c.WriteMessage(websocket.TextMessage, ping)
	expect(err).To.Be.Nil()

	_, message, err := c.ReadMessage()
	expect(err).To.Be.Nil()
	var e event
	err = json.Unmarshal(message, &e)
	expect(err).To.Be.Nil()
	expect(e.Cmd).To.Equal("pong")
	expect(e.Error == nil).Not.To.Be.True()
}

func TestItRepondsToInvalidMessages(t *testing.T) {
	expect := expect.New(t)

	server := httptest.NewServer(api.New(nil, nil, nil, ""))
	defer server.Close()

	url := strings.Replace(server.URL, "http://", "ws://", 1)
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	expect(err).To.Be.Nil()
	defer c.Close()

	badMessage := []byte(`{"cmd":"foobar"}`)
	err = c.WriteMessage(websocket.TextMessage, badMessage)
	expect(err).To.Be.Nil()

	_, message, err := c.ReadMessage()
	expect(err).To.Be.Nil()
	var e event
	err = json.Unmarshal(message, &e)
	expect(err).To.Be.Nil()
	expect(e.Cmd).To.Equal("foobar")
	expect(e.Error).Not.To.Be.Nil()
	expect(e.Error.Code).To.Equal(1)
	expect(e.Error.Text).To.Equal("invalid command")
}

func TestItChecksOriginAndHostMatch(t *testing.T) {
	expect := expect.New(t)

	server := httptest.NewServer(api.New(nil, nil, nil, ""))
	defer server.Close()

	url := strings.Replace(server.URL, "http://", "ws://", 1)
	headers := http.Header{}
	headers.Set("origin", "bad-origin")
	_, _, err := websocket.DefaultDialer.Dial(url, headers)
	expect(err).Not.To.Be.Nil()
}

func TestItPingsPeriodically(t *testing.T) {
	expect := expect.New(t)

	api := api.New(nil, nil, nil, "", api.WithPingInterval(time.Millisecond))
	server := httptest.NewServer(api)
	defer server.Close()

	url := strings.Replace(server.URL, "http://", "ws://", 1)
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	expect(err).To.Be.Nil()
	defer c.Close()

	gotPing := make(chan struct{})
	var once sync.Once
	oldHandler := c.PingHandler()
	c.SetPingHandler(func(data string) error {
		once.Do(func() {
			close(gotPing)
		})
		return oldHandler(data)
	})

	badMessage := []byte(`{"cmd":"foobar"}`)
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
