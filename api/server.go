package api

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	uuid "github.com/satori/go.uuid"

	"github.com/jasonkeene/anubot-server/bot"
	"github.com/jasonkeene/anubot-server/stream"
	"github.com/jasonkeene/anubot-server/twitch"
	"github.com/jasonkeene/anubot-server/twitch/oauth"
)

// Store is the object the Server uses to persist data.
type Store interface {
	RegisterUser(username, password string) (userID string, err error)
	AuthenticateUser(username, password string) (userID string, authenticated bool)
	TwitchClearAuth(userID string)
	TwitchAuthenticated(userID string) (authenticated bool)
	TwitchStreamerAuthenticated(userID string) (authenticated bool)
	TwitchStreamerCredentials(userID string) (string, string, int)
	TwitchBotAuthenticated(userID string) (authenticated bool)
	TwitchBotCredentials(userID string) (string, string, int)
	FetchRecentMessages(userID string) ([]stream.RXMessage, error)

	oauth.NonceStore
}

// Server responds to websocket events sent from the client.
type Server struct {
	botManager          *bot.Manager
	streamManager       *stream.Manager
	subEndpoints        []string
	store               Store
	twitchClient        *twitch.API
	twitchOauthClientID string
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

// New creates a new Server.
func New(
	botManager *bot.Manager,
	streamManager *stream.Manager,
	store Store,
	twitchClient *twitch.API,
	twitchOauthClientID string,
	opts ...Option,
) *Server {
	s := &Server{
		botManager:          botManager,
		streamManager:       streamManager,
		subEndpoints:        []string{"inproc://dispatch-pub"},
		store:               store,
		twitchClient:        twitchClient,
		twitchOauthClientID: twitchOauthClientID,
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
		log.Print("got error while upgrading ws conn:", err)
		return
	}
	defer func() {
		err := ws.Close()
		if err != nil {
			log.Printf("got error while closing ws conn: %s", err)
		}
	}()

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
		handler.HandleEvent(e, s)
	}
}
