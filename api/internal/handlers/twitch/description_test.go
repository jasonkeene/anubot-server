package twitch_test

import (
	"errors"
	"testing"

	"github.com/a8m/expect"
	"github.com/jasonkeene/anubot-server/api/internal/handlers"
	"github.com/jasonkeene/anubot-server/api/internal/handlers/twitch"
	"github.com/jasonkeene/anubot-server/store"
)

func TestUpdatingDescription(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{
		userID: "test-user-id",
	}
	spyCredsProvider := &SpyCredentialsProvider{
		creds: store.TwitchCredentials{
			StreamerUsername: "test-streamer-username",
			StreamerPassword: "test-streamer-password",
		},
	}
	spyClient := &SpyClient{}
	handler := twitch.NewUpdateChatDescriptionHandler(spyCredsProvider, spyClient)
	event := handlers.Event{
		Cmd:       "twitch-update-chat-description",
		RequestID: "test-request-id",
		Payload: map[string]interface{}{
			"status": "test-status",
			"game":   "test-game",
		},
	}

	handler.HandleEvent(event, spySession)

	expected := handlers.Event{
		Cmd:       "twitch-update-chat-description",
		RequestID: "test-request-id",
	}
	expect(spyCredsProvider.calledWith).To.Equal("test-user-id")
	expect(spyClient.updateDescriptionCalledWithStatus).To.Equal("test-status")
	expect(spyClient.updateDescriptionCalledWithGame).To.Equal("test-game")
	expect(spyClient.updateDescriptionCalledWithChannel).To.Equal("test-streamer-username")
	expect(spyClient.updateDescriptionCalledWithToken).To.Equal("test-streamer-password")
	expect(spySession.sendCalledWith).To.Equal(expected)
}

func TestInvalidUpdateRequest(t *testing.T) {
	expect := expect.New(t)

	cases := map[string]interface{}{
		"empty payload": nil,
		"invalid status": map[string]interface{}{
			"status": 1234,
			"game":   "test-game",
		},
		"invalid game": map[string]interface{}{
			"status": "test-status",
			"game":   1234,
		},
	}

	for _, payload := range cases {
		spySession := &SpySession{}
		spyCredsProvider := &SpyCredentialsProvider{}
		spyClient := &SpyClient{}
		handler := twitch.NewUpdateChatDescriptionHandler(spyCredsProvider, spyClient)
		event := handlers.Event{
			Cmd:       "twitch-update-chat-description",
			RequestID: "test-request-id",
			Payload:   payload,
		}

		handler.HandleEvent(event, spySession)

		expected := handlers.Event{
			Cmd:       "twitch-update-chat-description",
			RequestID: "test-request-id",
			Error:     handlers.InvalidPayload,
		}
		expect(spySession.sendCalledWith).To.Equal(expected)
	}
}

func TestUnableToGetTwitchCredentials(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{}
	spyCredsProvider := &SpyCredentialsProvider{
		err: errors.New("test-error"),
	}
	spyClient := &SpyClient{}
	handler := twitch.NewUpdateChatDescriptionHandler(spyCredsProvider, spyClient)
	event := handlers.Event{
		Cmd:       "twitch-update-chat-description",
		RequestID: "test-request-id",
		Payload: map[string]interface{}{
			"status": "test-status",
			"game":   "test-game",
		},
	}

	handler.HandleEvent(event, spySession)

	expected := handlers.Event{
		Cmd:       "twitch-update-chat-description",
		RequestID: "test-request-id",
		Error:     handlers.UnknownError,
	}
	expect(spySession.sendCalledWith).To.Equal(expected)
}

func TestUnableToUpdateChatDescription(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{}
	spyCredsProvider := &SpyCredentialsProvider{}
	spyClient := &SpyClient{
		updateErr: errors.New("test-error"),
	}
	handler := twitch.NewUpdateChatDescriptionHandler(spyCredsProvider, spyClient)
	event := handlers.Event{
		Cmd:       "twitch-update-chat-description",
		RequestID: "test-request-id",
		Payload: map[string]interface{}{
			"status": "test-status",
			"game":   "test-game",
		},
	}

	handler.HandleEvent(event, spySession)

	expected := handlers.Event{
		Cmd:       "twitch-update-chat-description",
		RequestID: "test-request-id",
		Error:     handlers.UnknownError,
	}
	expect(spySession.sendCalledWith).To.Equal(expected)
}
