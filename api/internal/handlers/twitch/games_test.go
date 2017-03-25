package twitch_test

import (
	"testing"

	"github.com/a8m/expect"
	"github.com/jasonkeene/anubot-server/api/internal/handlers"
	"github.com/jasonkeene/anubot-server/api/internal/handlers/twitch"
	twitchAPI "github.com/jasonkeene/anubot-server/twitch"
)

func TestGamesRequest(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{}
	games := []twitchAPI.Game{}
	spyClient := &SpyClient{
		games: games,
	}
	handler := twitch.NewGamesHandler(spyClient)
	event := handlers.Event{
		Cmd:       "twitch-games",
		RequestID: "test-request-id",
	}

	handler.HandleEvent(event, spySession)

	expected := handlers.Event{
		Cmd:       "twitch-games",
		RequestID: "test-request-id",
		Payload:   games,
	}
	expect(spySession.sendCalledWith).To.Equal(expected)
}
