package api_test

import (
	"errors"
	"testing"

	"github.com/a8m/expect"
	"github.com/jasonkeene/anubot-server/store"
)

func TestBTTVEmojiWithTwitchStreamerAuthed(t *testing.T) {
	expect := expect.New(t)

	server, cleanup := setupServer()
	defer cleanup()
	client, cleanup := setupClient(server.url)
	defer cleanup()
	authenticate(server, client)

	server.mockStore.TwitchCredentialsOutput.Creds <- store.TwitchCredentials{
		StreamerAuthenticated: true,
		StreamerUsername:      "test-username",
		StreamerPassword:      "test-password",
		StreamerTwitchUserID:  3,
	}
	server.mockStore.TwitchCredentialsOutput.Err <- nil
	expectedEmoji := map[string]string{
		"PawFive": "https://cdn.betterttv.net/emote/570fc2e801764c68585059b9/1x",
	}
	server.mockBTTVClient.EmojiOutput.Emoji <- expectedEmoji
	server.mockBTTVClient.EmojiOutput.Err <- nil

	emojiReq := event{
		Cmd:       "bttv-emoji",
		RequestID: requestID(),
	}
	expectedResp := event{
		Cmd:       "bttv-emoji",
		RequestID: emojiReq.RequestID,
		Payload:   expectedEmoji,
	}
	client.SendEvent(emojiReq)
	expect(<-server.mockBTTVClient.EmojiInput.Channel).To.Equal("test-username")
	emojiResp := client.ReadEvent()
	expect(emojiResp.Cmd).To.Equal(expectedResp.Cmd)
	expect(emojiResp.RequestID).To.Equal(expectedResp.RequestID)
	payload := emojiResp.Payload.(map[string]interface{})
	expect(payload["PawFive"].(string)).To.Equal(payload["PawFive"])
	expect(emojiResp.Error).To.Equal(expectedResp.Error)
}

func TestBTTVEmojiWithoutTwitchStreamerAuthed(t *testing.T) {
	expect := expect.New(t)

	server, cleanup := setupServer()
	defer cleanup()
	client, cleanup := setupClient(server.url)
	defer cleanup()
	authenticate(server, client)

	server.mockStore.TwitchCredentialsOutput.Creds <- store.TwitchCredentials{}
	server.mockStore.TwitchCredentialsOutput.Err <- nil
	expectedEmoji := map[string]string{
		"PawFive": "https://cdn.betterttv.net/emote/570fc2e801764c68585059b9/1x",
	}
	server.mockBTTVClient.EmojiOutput.Emoji <- expectedEmoji
	server.mockBTTVClient.EmojiOutput.Err <- nil

	emojiReq := event{
		Cmd:       "bttv-emoji",
		RequestID: requestID(),
	}
	expectedResp := event{
		Cmd:       "bttv-emoji",
		RequestID: emojiReq.RequestID,
		Payload:   expectedEmoji,
	}
	client.SendEvent(emojiReq)
	expect(<-server.mockBTTVClient.EmojiInput.Channel).To.Equal("")
	emojiResp := client.ReadEvent()
	expect(emojiResp.Cmd).To.Equal(expectedResp.Cmd)
	expect(emojiResp.RequestID).To.Equal(expectedResp.RequestID)
	payload := emojiResp.Payload.(map[string]interface{})
	expect(payload["PawFive"].(string)).To.Equal(payload["PawFive"])
	expect(emojiResp.Error).To.Equal(expectedResp.Error)
}

func TestBTTVEmojiWhenBTTVIsDown(t *testing.T) {
	expect := expect.New(t)

	server, cleanup := setupServer()
	defer cleanup()
	client, cleanup := setupClient(server.url)
	defer cleanup()
	authenticate(server, client)

	server.mockStore.TwitchCredentialsOutput.Creds <- store.TwitchCredentials{}
	server.mockStore.TwitchCredentialsOutput.Err <- nil
	server.mockBTTVClient.EmojiOutput.Emoji <- nil
	server.mockBTTVClient.EmojiOutput.Err <- errors.New("foobar")

	emojiReq := event{
		Cmd:       "bttv-emoji",
		RequestID: requestID(),
	}
	expectedResp := event{
		Cmd:       "bttv-emoji",
		RequestID: emojiReq.RequestID,
		Error: &eventErr{
			Code: 8,
			Text: "unable to gather emoji from bttv api",
		},
	}
	client.SendEvent(emojiReq)
	expect(<-server.mockBTTVClient.EmojiInput.Channel).To.Equal("")
	expect(client.ReadEvent()).To.Equal(expectedResp)
}

func TestBTTVEmojiUnauthenticated(t *testing.T) {
	expect := expect.New(t)

	server, cleanup := setupServer()
	defer cleanup()
	client, cleanup := setupClient(server.url)
	defer cleanup()

	emojiReq := event{
		Cmd:       "bttv-emoji",
		RequestID: requestID(),
	}
	expectedResp := event{
		Cmd:       "bttv-emoji",
		RequestID: emojiReq.RequestID,
		Error: &eventErr{
			Code: 4,
			Text: "authentication error",
		},
	}
	client.SendEvent(emojiReq)
	expect(client.ReadEvent()).To.Equal(expectedResp)
}
