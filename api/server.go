package api

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"

	"github.com/jasonkeene/anubot-server/api/internal/handlers"
	"github.com/jasonkeene/anubot-server/api/internal/handlers/auth"
	"github.com/jasonkeene/anubot-server/api/internal/handlers/bttv"
	"github.com/jasonkeene/anubot-server/api/internal/handlers/general"
	"github.com/jasonkeene/anubot-server/api/internal/handlers/twitch"
	bttvAPI "github.com/jasonkeene/anubot-server/bttv"
	"github.com/jasonkeene/anubot-server/store"
	"github.com/jasonkeene/anubot-server/stream"
	twitchAPI "github.com/jasonkeene/anubot-server/twitch"
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
	User(token string) (userData twitchAPI.UserData, err error)
	StreamInfo(channel string) (status, game string, err error)
	Games() (games []twitchAPI.Game)
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
	handlers               map[string]handlers.EventHandler
	upgrader               websocket.Upgrader
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
		bttvClient:             bttvAPI.New(),
		pingInterval:           5 * time.Second,
		nonceGen:               oauth.GenerateNonce,
	}
	s.createHandlers()
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *Server) createHandlers() {
	s.handlers = make(map[string]handlers.EventHandler)

	// public
	{
		// general
		s.handlers["ping"] = handlers.EventHandlerFunc(general.PingHandler)
		s.handlers["methods"] = general.NewMethodsHandler(s.handlers)

		// authentication
		s.handlers["register"] = auth.NewRegisterHandler(s.store)
		s.handlers["authenticate"] = auth.NewAuthenticateHandler(s.store)
		s.handlers["logout"] = handlers.EventHandlerFunc(auth.LogoutHandler)
	}

	// authenticated
	{
		// twitch oauth
		s.handlers["twitch-oauth-start"] = auth.AuthenticateWrapper(
			twitch.NewOauthStartHandler(
				s.store,
				twitch.NonceGenerator(s.nonceGen),
				s.store,
				s.twitchOauthClientID,
				s.twitchOauthRedirectURL,
				s.twitchOauthCallbacks,
			),
		)
		s.handlers["twitch-clear-auth"] = auth.AuthenticateWrapper(
			twitch.NewClearAuthHandler(s.store),
		)

		// user information
		s.handlers["twitch-user-details"] = auth.AuthenticateWrapper(
			twitch.NewUserDetailsHandler(s.store, s.twitchClient),
		)

		// twitch
		s.handlers["twitch-games"] = auth.AuthenticateWrapper(
			twitch.NewGamesHandler(s.twitchClient),
		)

		// bttv
		s.handlers["bttv-emoji"] = auth.AuthenticateWrapper(
			bttv.NewEmojiHandler(s.store, s.bttvClient),
		)
	}

	// twitch authenticated
	{
		// twitch chat
		s.handlers["twitch-stream-messages"] = auth.AuthenticateWrapper(
			twitch.AuthenticateWrapper(
				s.store,
				twitch.NewStreamMessagesHandler(
					s.store,
					s.streamManager,
					s.subEndpoints,
				),
			),
		)
		s.handlers["twitch-send-message"] = auth.AuthenticateWrapper(
			twitch.AuthenticateWrapper(
				s.store,
				twitch.NewSendMessageHandler(s.store, s.streamManager),
			),
		)
		s.handlers["twitch-update-chat-description"] = auth.AuthenticateWrapper(
			twitch.AuthenticateWrapper(
				s.store,
				twitch.NewUpdateChatDescriptionHandler(s.store, s.twitchClient),
			),
		)
	}
}

// ServeHTTP processes the websocket connection, reading and writing events.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ws, err := s.upgrader.Upgrade(w, r, nil)
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
	go s.writePings(ws)

	sess := &session{
		id: uuid.NewV4().String(),
		ws: ws,
	}
	log.Printf("serving session: %s", sess.id)
	defer log.Printf("done serving session: %s", sess.id)

	for {
		e, err := sess.Receive()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				log.Printf("got error while rx event: %s %s", sess.id, err)
			}
			return
		}
		handler, ok := s.handlers[e.Cmd]
		if !ok {
			log.Printf("received invalid command: %s %s", sess.id, e.Cmd)
			err = sess.Send(handlers.Event{
				Cmd:       e.Cmd,
				RequestID: e.RequestID,
				Error:     handlers.InvalidCommand,
			})
			if err != nil {
				log.Printf("unable to tx: %s", err)
			}
			continue
		}
		handler.HandleEvent(e, sess)
	}
}

func (s *Server) writePings(ws *websocket.Conn) {
	for {
		err := ws.WriteControl(websocket.PingMessage, nil, time.Now().Add(time.Second))
		if err != nil {
			log.Print("unable to write ping control message: ", err)
			return
		}
		time.Sleep(s.pingInterval)
	}
}
