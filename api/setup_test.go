package api_test

import (
	"crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/jasonkeene/anubot-server/api"
	"github.com/jasonkeene/anubot-server/store"
	"github.com/pebbe/zmq4"
	uuid "github.com/satori/go.uuid"
)

type event struct {
	Cmd       string      `json:"cmd"`
	Payload   interface{} `json:"payload"`
	RequestID string      `json:"request_id"`
	Error     *eventErr   `json:"error"`
}

type eventErr struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

func init() {
	flag.Parse()
	if !testing.Verbose() {
		log.SetOutput(ioutil.Discard)
	}
}

type server struct {
	s                 *httptest.Server
	url               string
	mockStreamManager *mockStreamManager
	mockStore         *mockStore
	mockTwitchClient  *mockTwitchClient
	mockBTTVClient    *mockBTTVClient
	subEndpoint       string
}

func setupServer() (server, func()) {
	msm := newMockStreamManager()
	ms := newMockStore()
	mtc := newMockTwitchClient()
	mbc := newMockBTTVClient()
	mcr := newMockOauthCallbackRegistrar()
	e := endpoint()
	ng := func() string {
		return "test-nonce"
	}
	s := httptest.NewServer(api.New(
		msm,
		ms,
		mtc,
		mcr,
		"some-client-id",
		"some-redirect-uri",
		api.WithBTTVClient(mbc),
		api.WithSubEndpoints([]string{e}),
		api.WithNonceGenerator(ng),
	))

	server := server{
		s:                 s,
		url:               strings.Replace(s.URL, "http://", "ws://", 1),
		mockStreamManager: msm,
		mockStore:         ms,
		mockTwitchClient:  mtc,
		mockBTTVClient:    mbc,
		subEndpoint:       e,
	}
	return server, func() {
		server.s.Close()
	}
}

type client struct {
	*websocket.Conn
}

func setupClient(url string) (client, func()) {
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		panic(err)
	}
	return client{
			Conn: c,
		}, func() {
			err := c.Close()
			if err != nil {
				panic(err)
			}
		}
}

func (c *client) SendEvent(e event) {
	data, err := json.Marshal(e)
	if err != nil {
		panic(err)
	}
	err = c.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		panic(err)
	}
}

func (c *client) ReadEvent() event {
	_, message, err := c.ReadMessage()
	if err != nil {
		panic(err)
	}

	var e event
	err = json.Unmarshal(message, &e)
	if err != nil {
		panic(err)
	}
	return e
}

func requestID() string {
	return uuid.NewV4().String()
}

func authenticate(s server, c client) {
	s.mockStore.RegisterUserOutput.UserID <- "some-user-id"
	s.mockStore.RegisterUserOutput.Err <- nil
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
	c.SendEvent(registerReq)
	registerResp := c.ReadEvent()
	if !reflect.DeepEqual(registerResp, expectedResp) {
		panic("register fail")
	}
}

func authenticateTwitch(s server) {
	s.mockStore.TwitchCredentialsOutput.Creds <- store.TwitchCredentials{
		StreamerAuthenticated: true,
		BotAuthenticated:      true,
	}
	s.mockStore.TwitchCredentialsOutput.Err <- nil
}

func endpoint() string {
	return "inproc://test-api-sub-" + randString()
}

func randString() string {
	b := make([]byte, 20)
	_, err := rand.Read(b)
	if err != nil {
		log.Panicf("unable to read randomness %s:", err)
	}
	return fmt.Sprintf("%x", b)
}

func setupPubSocket(endpoint string) *zmq4.Socket {
	pub, err := zmq4.NewSocket(zmq4.PUB)
	if err != nil {
		log.Panicf("unable to create pub socket %s:", err)
	}
	err = pub.Bind(endpoint)
	if err != nil {
		log.Panicf("unable to connect pub socket to endpoint: %s", err)
	}
	return pub
}
