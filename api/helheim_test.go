// This file was generated by github.com/nelsam/hel.  Do not
// edit this code by hand unless you *really* know what you're
// doing.  Expect any changes made manually to be overwritten
// the next time hel regenerates this file.

package api_test

import (
	"github.com/jasonkeene/anubot-server/store"
	"github.com/jasonkeene/anubot-server/stream"
	"github.com/jasonkeene/anubot-server/twitch"
)

type mockStreamManager struct {
	ConnectTwitchCalled chan bool
	ConnectTwitchInput  struct {
		User, Pass, Channel chan string
	}
	SendCalled chan bool
	SendInput  struct {
		Msg chan stream.TXMessage
	}
}

func newMockStreamManager() *mockStreamManager {
	m := &mockStreamManager{}
	m.ConnectTwitchCalled = make(chan bool, 100)
	m.ConnectTwitchInput.User = make(chan string, 100)
	m.ConnectTwitchInput.Pass = make(chan string, 100)
	m.ConnectTwitchInput.Channel = make(chan string, 100)
	m.SendCalled = make(chan bool, 100)
	m.SendInput.Msg = make(chan stream.TXMessage, 100)
	return m
}
func (m *mockStreamManager) ConnectTwitch(user, pass, channel string) {
	m.ConnectTwitchCalled <- true
	m.ConnectTwitchInput.User <- user
	m.ConnectTwitchInput.Pass <- pass
	m.ConnectTwitchInput.Channel <- channel
}
func (m *mockStreamManager) Send(msg stream.TXMessage) {
	m.SendCalled <- true
	m.SendInput.Msg <- msg
}

type mockBTTVClient struct {
	EmojiCalled chan bool
	EmojiInput  struct {
		Channel chan string
	}
	EmojiOutput struct {
		Emoji chan map[string]string
		Err   chan error
	}
}

func newMockBTTVClient() *mockBTTVClient {
	m := &mockBTTVClient{}
	m.EmojiCalled = make(chan bool, 100)
	m.EmojiInput.Channel = make(chan string, 100)
	m.EmojiOutput.Emoji = make(chan map[string]string, 100)
	m.EmojiOutput.Err = make(chan error, 100)
	return m
}
func (m *mockBTTVClient) Emoji(channel string) (emoji map[string]string, err error) {
	m.EmojiCalled <- true
	m.EmojiInput.Channel <- channel
	return <-m.EmojiOutput.Emoji, <-m.EmojiOutput.Err
}

type mockTwitchClient struct {
	UserCalled chan bool
	UserInput  struct {
		Token chan string
	}
	UserOutput struct {
		UserData chan twitch.UserData
		Err      chan error
	}
	StreamInfoCalled chan bool
	StreamInfoInput  struct {
		Channel chan string
	}
	StreamInfoOutput struct {
		Status, Game chan string
		Err          chan error
	}
	GamesCalled chan bool
	GamesOutput struct {
		Games chan []twitch.Game
	}
	UpdateDescriptionCalled chan bool
	UpdateDescriptionInput  struct {
		Status, Game, Channel, Token chan string
	}
	UpdateDescriptionOutput struct {
		Err chan error
	}
}

func newMockTwitchClient() *mockTwitchClient {
	m := &mockTwitchClient{}
	m.UserCalled = make(chan bool, 100)
	m.UserInput.Token = make(chan string, 100)
	m.UserOutput.UserData = make(chan twitch.UserData, 100)
	m.UserOutput.Err = make(chan error, 100)
	m.StreamInfoCalled = make(chan bool, 100)
	m.StreamInfoInput.Channel = make(chan string, 100)
	m.StreamInfoOutput.Status = make(chan string, 100)
	m.StreamInfoOutput.Game = make(chan string, 100)
	m.StreamInfoOutput.Err = make(chan error, 100)
	m.GamesCalled = make(chan bool, 100)
	m.GamesOutput.Games = make(chan []twitch.Game, 100)
	m.UpdateDescriptionCalled = make(chan bool, 100)
	m.UpdateDescriptionInput.Status = make(chan string, 100)
	m.UpdateDescriptionInput.Game = make(chan string, 100)
	m.UpdateDescriptionInput.Channel = make(chan string, 100)
	m.UpdateDescriptionInput.Token = make(chan string, 100)
	m.UpdateDescriptionOutput.Err = make(chan error, 100)
	return m
}
func (m *mockTwitchClient) User(token string) (userData twitch.UserData, err error) {
	m.UserCalled <- true
	m.UserInput.Token <- token
	return <-m.UserOutput.UserData, <-m.UserOutput.Err
}
func (m *mockTwitchClient) StreamInfo(channel string) (status, game string, err error) {
	m.StreamInfoCalled <- true
	m.StreamInfoInput.Channel <- channel
	return <-m.StreamInfoOutput.Status, <-m.StreamInfoOutput.Game, <-m.StreamInfoOutput.Err
}
func (m *mockTwitchClient) Games() (games []twitch.Game) {
	m.GamesCalled <- true
	return <-m.GamesOutput.Games
}
func (m *mockTwitchClient) UpdateDescription(status, game, channel, token string) (err error) {
	m.UpdateDescriptionCalled <- true
	m.UpdateDescriptionInput.Status <- status
	m.UpdateDescriptionInput.Game <- game
	m.UpdateDescriptionInput.Channel <- channel
	m.UpdateDescriptionInput.Token <- token
	return <-m.UpdateDescriptionOutput.Err
}

type mockStore struct {
	RegisterUserCalled chan bool
	RegisterUserInput  struct {
		Username, Password chan string
	}
	RegisterUserOutput struct {
		UserID chan string
		Err    chan error
	}
	AuthenticateUserCalled chan bool
	AuthenticateUserInput  struct {
		Username, Password chan string
	}
	AuthenticateUserOutput struct {
		UserID        chan string
		Authenticated chan bool
		Err           chan error
	}
	OauthNonceCalled chan bool
	OauthNonceInput  struct {
		UserID chan string
		Tu     chan store.TwitchUser
	}
	OauthNonceOutput struct {
		Nonce chan string
		Err   chan error
	}
	StoreOauthNonceCalled chan bool
	StoreOauthNonceInput  struct {
		UserID chan string
		Tu     chan store.TwitchUser
		Nonce  chan string
	}
	StoreOauthNonceOutput struct {
		Err chan error
	}
	OauthNonceExistsCalled chan bool
	OauthNonceExistsInput  struct {
		Nonce chan string
	}
	OauthNonceExistsOutput struct {
		Exists chan bool
		Err    chan error
	}
	FinishOauthNonceCalled chan bool
	FinishOauthNonceInput  struct {
		Nonce, TwitchUsername chan string
		TwitchUserID          chan int
		Od                    chan store.OauthData
	}
	FinishOauthNonceOutput struct {
		Err chan error
	}
	TwitchCredentialsCalled chan bool
	TwitchCredentialsInput  struct {
		UserID chan string
	}
	TwitchCredentialsOutput struct {
		Creds chan store.TwitchCredentials
		Err   chan error
	}
	TwitchClearAuthCalled chan bool
	TwitchClearAuthInput  struct {
		UserID chan string
	}
	TwitchClearAuthOutput struct {
		Err chan error
	}
	FetchRecentMessagesCalled chan bool
	FetchRecentMessagesInput  struct {
		UserID chan string
	}
	FetchRecentMessagesOutput struct {
		Msgs chan []stream.RXMessage
		Err  chan error
	}
}

func newMockStore() *mockStore {
	m := &mockStore{}
	m.RegisterUserCalled = make(chan bool, 100)
	m.RegisterUserInput.Username = make(chan string, 100)
	m.RegisterUserInput.Password = make(chan string, 100)
	m.RegisterUserOutput.UserID = make(chan string, 100)
	m.RegisterUserOutput.Err = make(chan error, 100)
	m.AuthenticateUserCalled = make(chan bool, 100)
	m.AuthenticateUserInput.Username = make(chan string, 100)
	m.AuthenticateUserInput.Password = make(chan string, 100)
	m.AuthenticateUserOutput.UserID = make(chan string, 100)
	m.AuthenticateUserOutput.Authenticated = make(chan bool, 100)
	m.AuthenticateUserOutput.Err = make(chan error, 100)
	m.OauthNonceCalled = make(chan bool, 100)
	m.OauthNonceInput.UserID = make(chan string, 100)
	m.OauthNonceInput.Tu = make(chan store.TwitchUser, 100)
	m.OauthNonceOutput.Nonce = make(chan string, 100)
	m.OauthNonceOutput.Err = make(chan error, 100)
	m.StoreOauthNonceCalled = make(chan bool, 100)
	m.StoreOauthNonceInput.UserID = make(chan string, 100)
	m.StoreOauthNonceInput.Tu = make(chan store.TwitchUser, 100)
	m.StoreOauthNonceInput.Nonce = make(chan string, 100)
	m.StoreOauthNonceOutput.Err = make(chan error, 100)
	m.OauthNonceExistsCalled = make(chan bool, 100)
	m.OauthNonceExistsInput.Nonce = make(chan string, 100)
	m.OauthNonceExistsOutput.Exists = make(chan bool, 100)
	m.OauthNonceExistsOutput.Err = make(chan error, 100)
	m.FinishOauthNonceCalled = make(chan bool, 100)
	m.FinishOauthNonceInput.Nonce = make(chan string, 100)
	m.FinishOauthNonceInput.TwitchUsername = make(chan string, 100)
	m.FinishOauthNonceInput.TwitchUserID = make(chan int, 100)
	m.FinishOauthNonceInput.Od = make(chan store.OauthData, 100)
	m.FinishOauthNonceOutput.Err = make(chan error, 100)
	m.TwitchCredentialsCalled = make(chan bool, 100)
	m.TwitchCredentialsInput.UserID = make(chan string, 100)
	m.TwitchCredentialsOutput.Creds = make(chan store.TwitchCredentials, 100)
	m.TwitchCredentialsOutput.Err = make(chan error, 100)
	m.TwitchClearAuthCalled = make(chan bool, 100)
	m.TwitchClearAuthInput.UserID = make(chan string, 100)
	m.TwitchClearAuthOutput.Err = make(chan error, 100)
	m.FetchRecentMessagesCalled = make(chan bool, 100)
	m.FetchRecentMessagesInput.UserID = make(chan string, 100)
	m.FetchRecentMessagesOutput.Msgs = make(chan []stream.RXMessage, 100)
	m.FetchRecentMessagesOutput.Err = make(chan error, 100)
	return m
}
func (m *mockStore) RegisterUser(username, password string) (userID string, err error) {
	m.RegisterUserCalled <- true
	m.RegisterUserInput.Username <- username
	m.RegisterUserInput.Password <- password
	return <-m.RegisterUserOutput.UserID, <-m.RegisterUserOutput.Err
}
func (m *mockStore) AuthenticateUser(username, password string) (userID string, authenticated bool, err error) {
	m.AuthenticateUserCalled <- true
	m.AuthenticateUserInput.Username <- username
	m.AuthenticateUserInput.Password <- password
	return <-m.AuthenticateUserOutput.UserID, <-m.AuthenticateUserOutput.Authenticated, <-m.AuthenticateUserOutput.Err
}
func (m *mockStore) OauthNonce(userID string, tu store.TwitchUser) (nonce string, err error) {
	m.OauthNonceCalled <- true
	m.OauthNonceInput.UserID <- userID
	m.OauthNonceInput.Tu <- tu
	return <-m.OauthNonceOutput.Nonce, <-m.OauthNonceOutput.Err
}
func (m *mockStore) StoreOauthNonce(userID string, tu store.TwitchUser, nonce string) (err error) {
	m.StoreOauthNonceCalled <- true
	m.StoreOauthNonceInput.UserID <- userID
	m.StoreOauthNonceInput.Tu <- tu
	m.StoreOauthNonceInput.Nonce <- nonce
	return <-m.StoreOauthNonceOutput.Err
}
func (m *mockStore) OauthNonceExists(nonce string) (exists bool, err error) {
	m.OauthNonceExistsCalled <- true
	m.OauthNonceExistsInput.Nonce <- nonce
	return <-m.OauthNonceExistsOutput.Exists, <-m.OauthNonceExistsOutput.Err
}
func (m *mockStore) FinishOauthNonce(nonce, twitchUsername string, twitchUserID int, od store.OauthData) (err error) {
	m.FinishOauthNonceCalled <- true
	m.FinishOauthNonceInput.Nonce <- nonce
	m.FinishOauthNonceInput.TwitchUsername <- twitchUsername
	m.FinishOauthNonceInput.TwitchUserID <- twitchUserID
	m.FinishOauthNonceInput.Od <- od
	return <-m.FinishOauthNonceOutput.Err
}
func (m *mockStore) TwitchCredentials(userID string) (creds store.TwitchCredentials, err error) {
	m.TwitchCredentialsCalled <- true
	m.TwitchCredentialsInput.UserID <- userID
	return <-m.TwitchCredentialsOutput.Creds, <-m.TwitchCredentialsOutput.Err
}
func (m *mockStore) TwitchClearAuth(userID string) (err error) {
	m.TwitchClearAuthCalled <- true
	m.TwitchClearAuthInput.UserID <- userID
	return <-m.TwitchClearAuthOutput.Err
}
func (m *mockStore) FetchRecentMessages(userID string) (msgs []stream.RXMessage, err error) {
	m.FetchRecentMessagesCalled <- true
	m.FetchRecentMessagesInput.UserID <- userID
	return <-m.FetchRecentMessagesOutput.Msgs, <-m.FetchRecentMessagesOutput.Err
}

type mockOauthCallbackRegistrar struct {
	RegisterCompletionCallbackCalled chan bool
	RegisterCompletionCallbackInput  struct {
		Nonce chan string
		F     chan func()
	}
}

func newMockOauthCallbackRegistrar() *mockOauthCallbackRegistrar {
	m := &mockOauthCallbackRegistrar{}
	m.RegisterCompletionCallbackCalled = make(chan bool, 100)
	m.RegisterCompletionCallbackInput.Nonce = make(chan string, 100)
	m.RegisterCompletionCallbackInput.F = make(chan func(), 100)
	return m
}
func (m *mockOauthCallbackRegistrar) RegisterCompletionCallback(nonce string, f func()) {
	m.RegisterCompletionCallbackCalled <- true
	m.RegisterCompletionCallbackInput.Nonce <- nonce
	m.RegisterCompletionCallbackInput.F <- f
}
