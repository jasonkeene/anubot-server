package bttv_test

import (
	"errors"
	"testing"

	"github.com/a8m/expect"
	"github.com/jasonkeene/anubot-server/api/internal/handlers"
	"github.com/jasonkeene/anubot-server/api/internal/handlers/bttv"
	"github.com/jasonkeene/anubot-server/store"
)

func TestTwitchStreamerAuthed(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{
		userID:        "test-user-id",
		authenticated: true,
	}
	spyCredsProvider := &SpyTwitchCredentialsProvider{
		creds: store.TwitchCredentials{
			StreamerAuthenticated: true,
			StreamerUsername:      "test-streamer-username",
		},
	}
	emoji := map[string]string{
		"some": "emoji",
	}
	spyEmojiProvider := &SpyEmojiProvider{
		emoji: emoji,
	}
	handler := bttv.NewEmojiHandler(spyCredsProvider, spyEmojiProvider)
	event := handlers.Event{
		Cmd:       "bttv-emoji",
		RequestID: "test-request-id",
	}

	handler.HandleEvent(event, spySession)

	expected := handlers.Event{
		Cmd:       "bttv-emoji",
		RequestID: "test-request-id",
		Payload:   emoji,
	}
	expect(spyCredsProvider.calledWith).To.Equal("test-user-id")
	expect(spyEmojiProvider.calledWith).To.Equal("test-streamer-username")
	expect(spySession.sendCalledWith).To.Equal(expected)
}

func TestWithoutTwitchStreamerAuthed(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{
		userID:        "test-user-id",
		authenticated: true,
	}
	spyCredsProvider := &SpyTwitchCredentialsProvider{}
	emoji := map[string]string{
		"some": "emoji",
	}
	spyEmojiProvider := &SpyEmojiProvider{
		emoji: emoji,
	}
	handler := bttv.NewEmojiHandler(spyCredsProvider, spyEmojiProvider)
	event := handlers.Event{
		Cmd:       "bttv-emoji",
		RequestID: "test-request-id",
	}

	handler.HandleEvent(event, spySession)

	expected := handlers.Event{
		Cmd:       "bttv-emoji",
		RequestID: "test-request-id",
		Payload:   emoji,
	}
	expect(spyCredsProvider.calledWith).To.Equal("test-user-id")
	expect(spyEmojiProvider.calledWith).To.Equal("")
	expect(spySession.sendCalledWith).To.Equal(expected)
}

func TestWhenBTTVIsDown(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{}
	spyCredsProvider := &SpyTwitchCredentialsProvider{}
	spyEmojiProvider := &SpyEmojiProvider{
		err: errors.New("test-error"),
	}
	handler := bttv.NewEmojiHandler(spyCredsProvider, spyEmojiProvider)
	event := handlers.Event{
		Cmd:       "bttv-emoji",
		RequestID: "test-request-id",
	}

	handler.HandleEvent(event, spySession)

	expected := handlers.Event{
		Cmd:       "bttv-emoji",
		RequestID: "test-request-id",
		Error:     handlers.BTTVUnavailable,
	}
	expect(spySession.sendCalledWith).To.Equal(expected)
}

func TestErrorWhenGettingCredentials(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{}
	spyCredsProvider := &SpyTwitchCredentialsProvider{
		err: errors.New("test-error"),
	}
	spyEmojiProvider := &SpyEmojiProvider{}
	handler := bttv.NewEmojiHandler(spyCredsProvider, spyEmojiProvider)
	event := handlers.Event{
		Cmd:       "bttv-emoji",
		RequestID: "test-request-id",
	}

	handler.HandleEvent(event, spySession)

	expected := handlers.Event{
		Cmd:       "bttv-emoji",
		RequestID: "test-request-id",
		Error:     handlers.UnknownError,
	}
	expect(spySession.sendCalledWith).To.Equal(expected)
}
