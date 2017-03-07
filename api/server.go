package api

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"

	"github.com/jasonkeene/anubot-server/bttv"
	"github.com/jasonkeene/anubot-server/store"
	"github.com/jasonkeene/anubot-server/stream"
	"github.com/jasonkeene/anubot-server/twitch"
	"github.com/jasonkeene/anubot-server/twitch/oauth"
)

// Store is used to persist data.
type Store interface {
	RegisterUser(username, password string) (userID string, err error)
	AuthenticateUser(username, password string) (userID string, authenticated bool, err error)

	OauthNonce(userID string, tu store.TwitchUser) (nonce string, err error)
	StoreOauthNonce(userID string, tu store.TwitchUser, nonce string) (err error)
	OauthNonceExists(nonce string) (exists bool, err error)
	FinishOauthNonce(nonce, twitchUsername string, twitchUserID int, od store.OauthData) (err error)

	TwitchCredentials(userID string) (creds store.TwitchCredentials, err error)
	TwitchClearAuth(userID string) (err error)

	FetchRecentMessages(userID string) (msgs []stream.RXMessage, err error)
}

// StreamManager is used to connect and send to third party chat.
type StreamManager interface {
	ConnectTwitch(user, pass, channel string)
	Send(msg stream.TXMessage)
}

// TwitchClient is used to communicate with Twitch's API.
type TwitchClient interface {
	User(token string) (userData twitch.UserData, err error)
	StreamInfo(channel string) (status, game string, err error)
	Games() (games []twitch.Game)
	UpdateDescription(status, game, channel, token string) (err error)
}

// BTTVClient is used to communicate with BTTV's API.
type BTTVClient interface {
	Emoji(channel string) (emoji map[string]string, err error)
}

// OauthCallbackRegistrar registers callbacks that are invoked when the oauth
// flow for a given nonce is complete.
type OauthCallbackRegistrar interface {
	RegisterCompletionCallback(nonce string, f func())
}

// NonceGenerator generates a random nonce to be used in the oauth flow.
type NonceGenerator func() string

// Server responds to websocket events sent from the client.
type Server struct {
	streamManager          StreamManager
	subEndpoints           []string
	store                  Store
	twitchClient           TwitchClient
	twitchOauthClientID    string
	twitchOauthRedirectURL string
	bttvClient             BTTVClient
	pingInterval           time.Duration
	twitchOauthCallbacks   OauthCallbackRegistrar
	nonceGen               NonceGenerator
}

// Option is used to configure a Server.
type Option func(*Server)

// WithSubEndpoints allows you to override the default endpoints that the
// server will attempt to subscribe to.
func WithSubEndpoints(endpoints []string) Option {
	return func(s *Server) {
		s.subEndpoints = endpoints
	}
}

// WithPingInterval allows you to configure how fast to send pings.
func WithPingInterval(pingInterval time.Duration) Option {
	return func(s *Server) {
		s.pingInterval = pingInterval
	}
}

// WithBTTVClient allows you to override the default BTTV API client.
func WithBTTVClient(b BTTVClient) Option {
	return func(s *Server) {
		s.bttvClient = b
	}
}

// WithNonceGenerator allows you to override the default nonce generator.
func WithNonceGenerator(ng NonceGenerator) Option {
	return func(s *Server) {
		s.nonceGen = ng
	}
}

// New creates a new Server.
func New(
	streamManager StreamManager,
	store Store,
	twitchClient TwitchClient,
	twitchOauthCallbacks OauthCallbackRegistrar,
	twitchOauthClientID string,
	twitchOauthRedirectURL string,
	opts ...Option,
) *Server {
	s := &Server{
		streamManager:          streamManager,
		subEndpoints:           []string{"inproc://dispatch-pub"},
		store:                  store,
		twitchClient:           twitchClient,
		twitchOauthCallbacks:   twitchOauthCallbacks,
		twitchOauthClientID:    twitchOauthClientID,
		twitchOauthRedirectURL: twitchOauthRedirectURL,
		bttvClient:             bttv.New(),
		pingInterval:           5 * time.Second,
		nonceGen:               oauth.GenerateNonce,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

var upgrader = websocket.Upgrader{}

// ServeHTTP processes the websocket connection, reading and writing events.
func (api *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("got error while upgrading ws conn: ", err)
		return
	}
	defer func() {
		err := ws.Close()
		if err != nil {
			log.Printf("got error while closing ws conn: %s", err)
		}
	}()
	go api.writePings(ws)

	s := &session{
		id:  uuid.NewV4().String(),
		ws:  ws,
		api: api,
	}
	log.Printf("serving session: %s", s.id)
	defer log.Printf("done serving session: %s", s.id)

	for {
		e, err := s.Receive()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				log.Printf("got error while rx event: %s %s", s.id, err)
			}
			return
		}
		handler, ok := eventHandlers[e.Cmd]
		if !ok {
			log.Printf("received invalid command: %s %s", s.id, e.Cmd)
			err = s.Send(event{
				Cmd:       e.Cmd,
				RequestID: e.RequestID,
				Error:     invalidCommand,
			})
			if err != nil {
				log.Printf("unable to tx: %s", err)
			}
			continue
		}
		handler(e, s)
	}
}

func (api *Server) writePings(ws *websocket.Conn) {
	for {
		err := ws.WriteControl(websocket.PingMessage, nil, time.Now().Add(time.Second))
		if err != nil {
			log.Print("unable to write ping control message: ", err)
			return
		}
		time.Sleep(api.pingInterval)
	}
}
