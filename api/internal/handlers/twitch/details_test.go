package twitch_test

import (
	"errors"
	"testing"

	"github.com/a8m/expect"
	"github.com/jasonkeene/anubot-server/api/internal/handlers"
	"github.com/jasonkeene/anubot-server/api/internal/handlers/twitch"
	"github.com/jasonkeene/anubot-server/store"
	twitchAPI "github.com/jasonkeene/anubot-server/twitch"
)

func TestFullyAuthenticated(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{
		userID: "test-user-id",
	}
	spyCredsProvider := &SpyCredentialsProvider{
		creds: store.TwitchCredentials{
			StreamerAuthenticated: true,
			StreamerUsername:      "test-streamer-username",
			StreamerPassword:      "test-streamer-password",
			StreamerTwitchUserID:  12345,

			BotAuthenticated: true,
			BotUsername:      "test-bot-username",
			BotPassword:      "test-bot-password",
			BotTwitchUserID:  54321,
		},
	}
	spyClient := &SpyClient{
		userData: twitchAPI.UserData{
			Logo:        "test-logo",
			DisplayName: "test-display-name",
		},
		status: "test-status",
		game:   "test-game",
	}
	handler := twitch.NewUserDetailsHandler(spyCredsProvider, spyClient)
	event := handlers.Event{
		Cmd:       "twitch-user-details",
		RequestID: "test-request-id",
	}

	handler.HandleEvent(event, spySession)

	expected := handlers.Event{
		Cmd:       "twitch-user-details",
		RequestID: "test-request-id",
		Payload: map[string]interface{}{
			"streamer_authenticated": true,
			"streamer_username":      "test-streamer-username",
			"streamer_logo":          "test-logo",
			"streamer_display_name":  "test-display-name",
			"streamer_status":        "test-status",
			"streamer_game":          "test-game",
			"bot_authenticated":      true,
			"bot_username":           "test-bot-username",
		},
	}
	expect(spyCredsProvider.calledWith).To.Equal("test-user-id")
	expect(spyClient.userCalledWith).To.Equal("test-streamer-password")
	expect(spyClient.streamInfoCalledWith).To.Equal("test-streamer-username")
	expect(spySession.sendCalledWith).To.Equal(expected)
}

func TestStreamerAuthenticatedOnly(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{
		userID: "test-user-id",
	}
	spyCredsProvider := &SpyCredentialsProvider{
		creds: store.TwitchCredentials{
			StreamerAuthenticated: true,
			StreamerUsername:      "test-streamer-username",
			StreamerPassword:      "test-streamer-password",
			StreamerTwitchUserID:  12345,
		},
	}
	spyClient := &SpyClient{
		userData: twitchAPI.UserData{
			Logo:        "test-logo",
			DisplayName: "test-display-name",
		},
		status: "test-status",
		game:   "test-game",
	}
	handler := twitch.NewUserDetailsHandler(spyCredsProvider, spyClient)
	event := handlers.Event{
		Cmd:       "twitch-user-details",
		RequestID: "test-request-id",
	}

	handler.HandleEvent(event, spySession)

	expected := handlers.Event{
		Cmd:       "twitch-user-details",
		RequestID: "test-request-id",
		Payload: map[string]interface{}{
			"streamer_authenticated": true,
			"streamer_username":      "test-streamer-username",
			"streamer_logo":          "test-logo",
			"streamer_display_name":  "test-display-name",
			"streamer_status":        "test-status",
			"streamer_game":          "test-game",
			"bot_authenticated":      false,
			"bot_username":           "",
		},
	}
	expect(spyCredsProvider.calledWith).To.Equal("test-user-id")
	expect(spyClient.userCalledWith).To.Equal("test-streamer-password")
	expect(spyClient.streamInfoCalledWith).To.Equal("test-streamer-username")
	expect(spySession.sendCalledWith).To.Equal(expected)
}

func TestUnauthenticated(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{
		userID: "test-user-id",
	}
	spyCredsProvider := &SpyCredentialsProvider{
		creds: store.TwitchCredentials{},
	}
	spyClient := &SpyClient{}
	handler := twitch.NewUserDetailsHandler(spyCredsProvider, spyClient)
	event := handlers.Event{
		Cmd:       "twitch-user-details",
		RequestID: "test-request-id",
	}

	handler.HandleEvent(event, spySession)

	expected := handlers.Event{
		Cmd:       "twitch-user-details",
		RequestID: "test-request-id",
		Payload: map[string]interface{}{
			"streamer_authenticated": false,
			"streamer_username":      "",
			"streamer_logo":          "",
			"streamer_display_name":  "",
			"streamer_status":        "",
			"streamer_game":          "",
			"bot_authenticated":      false,
			"bot_username":           "",
		},
	}
	expect(spyCredsProvider.calledWith).To.Equal("test-user-id")
	expect(spySession.sendCalledWith).To.Equal(expected)
}

func TestErrorWithCredsProvider(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{
		userID: "test-user-id",
	}
	spyCredsProvider := &SpyCredentialsProvider{
		err: errors.New("test-error"),
	}
	spyClient := &SpyClient{}
	handler := twitch.NewUserDetailsHandler(spyCredsProvider, spyClient)
	event := handlers.Event{
		Cmd:       "twitch-user-details",
		RequestID: "test-request-id",
	}

	handler.HandleEvent(event, spySession)

	expected := handlers.Event{
		Cmd:       "twitch-user-details",
		RequestID: "test-request-id",
		Error:     handlers.UnknownError,
	}
	expect(spyCredsProvider.calledWith).To.Equal("test-user-id")
	expect(spySession.sendCalledWith).To.Equal(expected)
}

func TestErrorWithGettingUserInfo(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{
		userID: "test-user-id",
	}
	spyCredsProvider := &SpyCredentialsProvider{
		creds: store.TwitchCredentials{
			StreamerAuthenticated: true,
			StreamerUsername:      "test-streamer-username",
			StreamerPassword:      "test-streamer-password",
			StreamerTwitchUserID:  12345,
		},
	}
	spyClient := &SpyClient{
		userErr: errors.New("test-error"),
	}
	handler := twitch.NewUserDetailsHandler(spyCredsProvider, spyClient)
	event := handlers.Event{
		Cmd:       "twitch-user-details",
		RequestID: "test-request-id",
	}

	handler.HandleEvent(event, spySession)

	expected := handlers.Event{
		Cmd:       "twitch-user-details",
		RequestID: "test-request-id",
		Error:     handlers.UnknownError,
	}
	expect(spyCredsProvider.calledWith).To.Equal("test-user-id")
	expect(spySession.sendCalledWith).To.Equal(expected)
}

func TestErrorWithGettingStreamInfo(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{
		userID: "test-user-id",
	}
	spyCredsProvider := &SpyCredentialsProvider{
		creds: store.TwitchCredentials{
			StreamerAuthenticated: true,
			StreamerUsername:      "test-streamer-username",
			StreamerPassword:      "test-streamer-password",
			StreamerTwitchUserID:  12345,
		},
	}
	spyClient := &SpyClient{
		streamInfoErr: errors.New("test-error"),
	}
	handler := twitch.NewUserDetailsHandler(spyCredsProvider, spyClient)
	event := handlers.Event{
		Cmd:       "twitch-user-details",
		RequestID: "test-request-id",
	}

	handler.HandleEvent(event, spySession)

	expected := handlers.Event{
		Cmd:       "twitch-user-details",
		RequestID: "test-request-id",
		Error:     handlers.UnknownError,
	}
	expect(spyCredsProvider.calledWith).To.Equal("test-user-id")
	expect(spySession.sendCalledWith).To.Equal(expected)
}
