package twitch_test

import (
	"errors"
	"testing"

	"github.com/a8m/expect"
	"github.com/jasonkeene/anubot-server/api/internal/handlers"
	"github.com/jasonkeene/anubot-server/api/internal/handlers/twitch"
	"github.com/jasonkeene/anubot-server/store"
)

func TestStreamerOauthStart(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{
		userID: "test-user-id",
	}
	spyNonceGen := func() string {
		return "test-nonce"
	}
	spyNonceStore := &SpyNonceStore{
		err: errors.New("test-error"),
	}
	spyCallbackRegistrar := &SpyOauthCallbackRegistrar{}
	handler := twitch.NewOauthStartHandler(
		nil,
		spyNonceGen,
		spyNonceStore,
		"test-oauth-client-id",
		"test-redirect-url",
		spyCallbackRegistrar,
	)
	event := handlers.Event{
		Cmd:       "twitch-oauth-start",
		RequestID: "test-request-id",
		Payload:   "streamer",
	}

	handler.HandleEvent(event, spySession)

	expected := handlers.Event{
		Cmd:       "twitch-oauth-start",
		RequestID: "test-request-id",
		Payload:   "https://api.twitch.tv/kraken/oauth2/authorize?client_id=test-oauth-client-id&redirect_uri=test-redirect-url&response_type=code&scope=user_read+user_blocks_edit+user_blocks_read+user_follows_edit+channel_read+channel_editor+channel_commercial+channel_stream+channel_subscriptions+user_subscriptions+channel_check_subscription+chat_login+channel_feed_read+channel_feed_edit&state=test-nonce",
	}
	expect(spyNonceStore.storeCalledWithUserID).To.Equal("test-user-id")
	expect(spyNonceStore.storeCalledWithTwitchUser).To.Equal(store.Streamer)
	expect(spyNonceStore.storeCalledWithNonce).To.Equal("test-nonce")
	expect(spySession.sendCalledWith).To.Equal(expected)
}

func TestNonceAlreadyExists(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{
		userID: "test-user-id",
	}
	spyNonceGen := func() string {
		return "test-nonce"
	}
	spyNonceStore := &SpyNonceStore{
		nonce: "existing-nonce",
	}
	spyCallbackRegistrar := &SpyOauthCallbackRegistrar{}
	handler := twitch.NewOauthStartHandler(
		nil,
		spyNonceGen,
		spyNonceStore,
		"test-oauth-client-id",
		"test-redirect-url",
		spyCallbackRegistrar,
	)
	event := handlers.Event{
		Cmd:       "twitch-oauth-start",
		RequestID: "test-request-id",
		Payload:   "streamer",
	}

	handler.HandleEvent(event, spySession)

	expected := handlers.Event{
		Cmd:       "twitch-oauth-start",
		RequestID: "test-request-id",
		Payload:   "https://api.twitch.tv/kraken/oauth2/authorize?client_id=test-oauth-client-id&redirect_uri=test-redirect-url&response_type=code&scope=user_read+user_blocks_edit+user_blocks_read+user_follows_edit+channel_read+channel_editor+channel_commercial+channel_stream+channel_subscriptions+user_subscriptions+channel_check_subscription+chat_login+channel_feed_read+channel_feed_edit&state=existing-nonce",
	}
	expect(spySession.sendCalledWith).To.Equal(expected)
}

func TestErrorWhenStoringNonce(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{
		userID: "test-user-id",
	}
	spyNonceGen := func() string {
		return "test-nonce"
	}
	spyNonceStore := &SpyNonceStore{
		err:      errors.New("test-error"),
		storeErr: errors.New("test-error"),
	}
	spyCallbackRegistrar := &SpyOauthCallbackRegistrar{}
	handler := twitch.NewOauthStartHandler(
		nil,
		spyNonceGen,
		spyNonceStore,
		"test-oauth-client-id",
		"test-redirect-url",
		spyCallbackRegistrar,
	)
	event := handlers.Event{
		Cmd:       "twitch-oauth-start",
		RequestID: "test-request-id",
		Payload:   "streamer",
	}

	handler.HandleEvent(event, spySession)

	expected := handlers.Event{
		Cmd:       "twitch-oauth-start",
		RequestID: "test-request-id",
		Error:     handlers.UnknownError,
	}
	expect(spyNonceStore.storeCalledWithUserID).To.Equal("test-user-id")
	expect(spyNonceStore.storeCalledWithTwitchUser).To.Equal(store.Streamer)
	expect(spyNonceStore.storeCalledWithNonce).To.Equal("test-nonce")
	expect(spySession.sendCalledWith).To.Equal(expected)
}

func TestBotOauthStart(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{
		userID: "test-user-id",
	}
	spyNonceGen := func() string {
		return "test-nonce"
	}
	spyNonceStore := &SpyNonceStore{
		err: errors.New("test-error"),
	}
	spyCallbackRegistrar := &SpyOauthCallbackRegistrar{}
	spyCredsProvider := &SpyCredentialsProvider{
		creds: store.TwitchCredentials{
			StreamerAuthenticated: true,
		},
	}
	handler := twitch.NewOauthStartHandler(
		spyCredsProvider,
		spyNonceGen,
		spyNonceStore,
		"test-oauth-client-id",
		"test-redirect-url",
		spyCallbackRegistrar,
	)
	event := handlers.Event{
		Cmd:       "twitch-oauth-start",
		RequestID: "test-request-id",
		Payload:   "bot",
	}

	handler.HandleEvent(event, spySession)

	expected := handlers.Event{
		Cmd:       "twitch-oauth-start",
		RequestID: "test-request-id",
		Payload:   "https://api.twitch.tv/kraken/oauth2/authorize?client_id=test-oauth-client-id&redirect_uri=test-redirect-url&response_type=code&scope=user_read+user_blocks_edit+user_blocks_read+user_follows_edit+channel_read+channel_editor+channel_commercial+channel_stream+channel_subscriptions+user_subscriptions+channel_check_subscription+chat_login+channel_feed_read+channel_feed_edit&state=test-nonce",
	}
	expect(spyNonceStore.storeCalledWithUserID).To.Equal("test-user-id")
	expect(spyNonceStore.storeCalledWithTwitchUser).To.Equal(store.Bot)
	expect(spyNonceStore.storeCalledWithNonce).To.Equal("test-nonce")
	expect(spySession.sendCalledWith).To.Equal(expected)
}

func TestBotOauthStartBeforeStreamer(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{
		userID: "test-user-id",
	}
	spyNonceGen := func() string {
		return "test-nonce"
	}
	spyNonceStore := &SpyNonceStore{
		err: errors.New("test-error"),
	}
	spyCallbackRegistrar := &SpyOauthCallbackRegistrar{}
	spyCredsProvider := &SpyCredentialsProvider{
		creds: store.TwitchCredentials{
			StreamerAuthenticated: false,
		},
	}
	handler := twitch.NewOauthStartHandler(
		spyCredsProvider,
		spyNonceGen,
		spyNonceStore,
		"test-oauth-client-id",
		"test-redirect-url",
		spyCallbackRegistrar,
	)
	event := handlers.Event{
		Cmd:       "twitch-oauth-start",
		RequestID: "test-request-id",
		Payload:   "bot",
	}

	handler.HandleEvent(event, spySession)

	expected := handlers.Event{
		Cmd:       "twitch-oauth-start",
		RequestID: "test-request-id",
		Error:     handlers.TwitchOauthStartOrderError,
	}
	expect(spySession.sendCalledWith).To.Equal(expected)
}

func TestInvalidPayloads(t *testing.T) {
	expect := expect.New(t)

	cases := map[string]interface{}{
		"empty payload":       nil,
		"invalid twitch user": "invalid",
	}

	for _, payload := range cases {
		spySession := &SpySession{}
		handler := twitch.NewOauthStartHandler(nil, nil, nil, "", "", nil)
		event := handlers.Event{
			Cmd:       "twitch-oauth-start",
			RequestID: "test-request-id",
			Payload:   payload,
		}

		handler.HandleEvent(event, spySession)

		expected := handlers.Event{
			Cmd:       "twitch-oauth-start",
			RequestID: "test-request-id",
			Error:     handlers.InvalidPayload,
		}
		expect(spySession.sendCalledWith).To.Equal(expected)
	}
}

func TestErrorWhenGettingTwitchCredentials(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{
		userID: "test-user-id",
	}
	spyNonceGen := func() string {
		return "test-nonce"
	}
	spyNonceStore := &SpyNonceStore{
		err: errors.New("test-error"),
	}
	spyCallbackRegistrar := &SpyOauthCallbackRegistrar{}
	spyCredsProvider := &SpyCredentialsProvider{
		err: errors.New("test-error"),
	}
	handler := twitch.NewOauthStartHandler(
		spyCredsProvider,
		spyNonceGen,
		spyNonceStore,
		"test-oauth-client-id",
		"test-redirect-url",
		spyCallbackRegistrar,
	)
	event := handlers.Event{
		Cmd:       "twitch-oauth-start",
		RequestID: "test-request-id",
		Payload:   "bot",
	}

	handler.HandleEvent(event, spySession)

	expected := handlers.Event{
		Cmd:       "twitch-oauth-start",
		RequestID: "test-request-id",
		Error:     handlers.UnknownError,
	}
	expect(spySession.sendCalledWith).To.Equal(expected)
}

func TestClearAuth(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{
		userID: "test-user-id",
	}
	spyAuthClearer := &SpyAuthClearer{}
	handler := twitch.NewClearAuthHandler(spyAuthClearer)
	event := handlers.Event{
		Cmd:       "twitch-clear-auth",
		RequestID: "test-request-id",
	}

	handler.HandleEvent(event, spySession)

	expected := handlers.Event{
		Cmd:       "twitch-clear-auth",
		RequestID: "test-request-id",
	}
	expect(spyAuthClearer.calledWith).To.Equal("test-user-id")
	expect(spySession.sendCalledWith).To.Equal(expected)
}

func TestWhenUnableToClearAuth(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{
		userID: "test-user-id",
	}
	spyAuthClearer := &SpyAuthClearer{
		err: errors.New("test-error"),
	}
	handler := twitch.NewClearAuthHandler(spyAuthClearer)
	event := handlers.Event{
		Cmd:       "twitch-clear-auth",
		RequestID: "test-request-id",
	}

	handler.HandleEvent(event, spySession)

	expected := handlers.Event{
		Cmd:       "twitch-clear-auth",
		RequestID: "test-request-id",
		Error:     handlers.UnknownError,
	}
	expect(spyAuthClearer.calledWith).To.Equal("test-user-id")
	expect(spySession.sendCalledWith).To.Equal(expected)
}
