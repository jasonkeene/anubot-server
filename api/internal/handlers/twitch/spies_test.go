package twitch_test

import (
	"sync"

	"github.com/jasonkeene/anubot-server/api/internal/handlers"
	"github.com/jasonkeene/anubot-server/store"
	"github.com/jasonkeene/anubot-server/stream"

	twitchAPI "github.com/jasonkeene/anubot-server/twitch"
)

type SpySession struct {
	handlers.Session
	sendCalledWith handlers.Event
	mu             sync.Mutex
	sendCalls_     []handlers.Event
	userID         string
	authenticated  bool
}

func (s *SpySession) Send(e handlers.Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sendCalledWith = e
	s.sendCalls_ = append(s.sendCalls_, e)
	return nil
}

func (s *SpySession) sendCalls() []handlers.Event {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sendCalls_
}

func (s *SpySession) Authenticated() (userID string, authenticated bool) {
	return s.userID, s.authenticated
}

type SpyCredentialsProvider struct {
	calledWith string
	creds      store.TwitchCredentials
	err        error
}

func (s *SpyCredentialsProvider) TwitchCredentials(userID string) (creds store.TwitchCredentials, err error) {
	s.calledWith = userID
	return s.creds, s.err
}

type SpyHandler struct {
	called bool
}

func (s *SpyHandler) HandleEvent(handlers.Event, handlers.Session) {
	s.called = true
}

type SpyClient struct {
	updateDescriptionCalledWithStatus  string
	updateDescriptionCalledWithGame    string
	updateDescriptionCalledWithChannel string
	updateDescriptionCalledWithToken   string
	updateErr                          error

	userCalledWith string
	userData       twitchAPI.UserData
	userErr        error

	streamInfoCalledWith string
	status               string
	game                 string
	streamInfoErr        error

	games []twitchAPI.Game
}

func (s *SpyClient) UpdateDescription(status, game, channel, token string) error {
	s.updateDescriptionCalledWithStatus = status
	s.updateDescriptionCalledWithGame = game
	s.updateDescriptionCalledWithChannel = channel
	s.updateDescriptionCalledWithToken = token
	return s.updateErr
}

func (s *SpyClient) User(token string) (twitchAPI.UserData, error) {
	s.userCalledWith = token
	return s.userData, s.userErr
}

func (s *SpyClient) StreamInfo(channel string) (status, game string, err error) {
	s.streamInfoCalledWith = channel
	return s.status, s.game, s.streamInfoErr
}

func (s *SpyClient) Games() (games []twitchAPI.Game) {
	return s.games
}

type SpyNonceStore struct {
	nonce string
	err   error

	storeCalledWithUserID     string
	storeCalledWithTwitchUser store.TwitchUser
	storeCalledWithNonce      string
	storeErr                  error
}

func (s *SpyNonceStore) OauthNonce(userID string, tu store.TwitchUser) (nonce string, err error) {
	return s.nonce, s.err
}
func (s *SpyNonceStore) StoreOauthNonce(userID string, tu store.TwitchUser, nonce string) (err error) {
	s.storeCalledWithUserID = userID
	s.storeCalledWithTwitchUser = tu
	s.storeCalledWithNonce = nonce
	return s.storeErr
}

type SpyOauthCallbackRegistrar struct{}

func (s *SpyOauthCallbackRegistrar) RegisterCompletionCallback(nonce string, f func()) {}

type SpyAuthClearer struct {
	calledWith string
	err        error
}

func (s *SpyAuthClearer) TwitchClearAuth(userID string) (err error) {
	s.calledWith = userID
	return s.err
}

type SpyStreamMessagesStore struct {
	recentMessagesCalledWith string
	recentMessages           []stream.RXMessage
	recentMessagesErr        error

	credsCalledWith string
	creds           store.TwitchCredentials
	credsErr        error
}

func (s *SpyStreamMessagesStore) FetchRecentMessages(userID string) (msgs []stream.RXMessage, err error) {
	s.recentMessagesCalledWith = userID
	return s.recentMessages, s.recentMessagesErr
}

func (s *SpyStreamMessagesStore) TwitchCredentials(userID string) (creds store.TwitchCredentials, err error) {
	s.credsCalledWith = userID
	return s.creds, s.credsErr
}

type connectArgs struct {
	user    string
	pass    string
	channel string
}

type SpyConnector struct {
	connectCalls []connectArgs
}

func (s *SpyConnector) ConnectTwitch(user, pass, channel string) {
	s.connectCalls = append(s.connectCalls, connectArgs{
		user:    user,
		pass:    pass,
		channel: channel,
	})
}

type SpyStreamManager struct {
	connectCalledWithUser    string
	connectCalledWithPass    string
	connectCalledWithChannel string

	messageSent stream.TXMessage
}

func (s *SpyStreamManager) ConnectTwitch(user string, pass string, channel string) {
	s.connectCalledWithUser = user
	s.connectCalledWithPass = pass
	s.connectCalledWithChannel = channel
}

func (s *SpyStreamManager) Send(msg stream.TXMessage) {
	s.messageSent = msg
}
