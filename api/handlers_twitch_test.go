package api_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/a8m/expect"
	ircclient "github.com/fluffle/goirc/client"
	"github.com/jasonkeene/anubot-server/store"
	"github.com/jasonkeene/anubot-server/stream"
	"github.com/jasonkeene/anubot-server/twitch"
)

func TestTwitchOauthStartStreamer(t *testing.T) {
	expect := expect.New(t)

	server, cleanup := setupServer()
	defer cleanup()
	client, cleanup := setupClient(server.url)
	defer cleanup()
	authenticate(server, client)

	server.mockStore.CreateOauthNonceOutput.Nonce <- "some-nonce"
	server.mockStore.CreateOauthNonceOutput.Err <- nil

	oauthStartReq := event{
		Cmd:       "twitch-oauth-start",
		RequestID: requestID(),
		Payload:   "streamer",
	}
	expectedResp := event{
		Cmd:       "twitch-oauth-start",
		RequestID: oauthStartReq.RequestID,
		Payload:   `https://api.twitch.tv/kraken/oauth2/authorize?client_id=some-client-id&redirect_uri=https%3A%2F%2Fanubot.io%2Ftwitch_oauth%2Fdone&response_type=code&scope=user_read+user_blocks_edit+user_blocks_read+user_follows_edit+channel_read+channel_editor+channel_commercial+channel_stream+channel_subscriptions+user_subscriptions+channel_check_subscription+chat_login+channel_feed_read+channel_feed_edit&state=some-nonce`,
	}
	client.SendEvent(oauthStartReq)
	expect(<-server.mockStore.CreateOauthNonceInput.Tu).To.Equal(store.Streamer)
	expect(<-server.mockStore.CreateOauthNonceInput.UserID).To.Equal("some-user-id")
	expect(client.ReadEvent()).To.Equal(expectedResp)
}

func TestTwitchOauthStartBotAfterStreamer(t *testing.T) {
	expect := expect.New(t)

	server, cleanup := setupServer()
	defer cleanup()
	client, cleanup := setupClient(server.url)
	defer cleanup()
	authenticate(server, client)

	server.mockStore.TwitchStreamerAuthenticatedOutput.Authenticated <- true
	server.mockStore.TwitchStreamerAuthenticatedOutput.Err <- nil
	server.mockStore.CreateOauthNonceOutput.Nonce <- "some-nonce"
	server.mockStore.CreateOauthNonceOutput.Err <- nil

	oauthStartReq := event{
		Cmd:       "twitch-oauth-start",
		RequestID: requestID(),
		Payload:   "bot",
	}
	expectedResp := event{
		Cmd:       "twitch-oauth-start",
		RequestID: oauthStartReq.RequestID,
		Payload:   `https://api.twitch.tv/kraken/oauth2/authorize?client_id=some-client-id&redirect_uri=https%3A%2F%2Fanubot.io%2Ftwitch_oauth%2Fdone&response_type=code&scope=user_read+user_blocks_edit+user_blocks_read+user_follows_edit+channel_read+channel_editor+channel_commercial+channel_stream+channel_subscriptions+user_subscriptions+channel_check_subscription+chat_login+channel_feed_read+channel_feed_edit&state=some-nonce`,
	}
	client.SendEvent(oauthStartReq)
	expect(<-server.mockStore.CreateOauthNonceInput.Tu).To.Equal(store.Bot)
	expect(<-server.mockStore.CreateOauthNonceInput.UserID).To.Equal("some-user-id")
	expect(client.ReadEvent()).To.Equal(expectedResp)
}

func TestTwitchOauthStartBotBeforeStreamer(t *testing.T) {
	expect := expect.New(t)

	server, cleanup := setupServer()
	defer cleanup()
	client, cleanup := setupClient(server.url)
	defer cleanup()
	authenticate(server, client)

	server.mockStore.TwitchStreamerAuthenticatedOutput.Authenticated <- false
	server.mockStore.TwitchStreamerAuthenticatedOutput.Err <- nil

	oauthStartReq := event{
		Cmd:       "twitch-oauth-start",
		RequestID: requestID(),
		Payload:   "bot",
	}
	expectedResp := event{
		Cmd:       "twitch-oauth-start",
		RequestID: oauthStartReq.RequestID,
		Error: &eventErr{
			Code: 6,
			Text: "unable to start oauth flow for bot, streamer not finished",
		},
	}
	client.SendEvent(oauthStartReq)
	expect(client.ReadEvent()).To.Equal(expectedResp)
}

func TestTwitchClearAuth(t *testing.T) {
	expect := expect.New(t)

	server, cleanup := setupServer()
	defer cleanup()
	client, cleanup := setupClient(server.url)
	defer cleanup()
	authenticate(server, client)

	server.mockStore.TwitchClearAuthOutput.Err <- nil

	clearReq := event{
		Cmd:       "twitch-clear-auth",
		RequestID: requestID(),
	}
	expectedResp := event{
		Cmd:       "twitch-clear-auth",
		RequestID: clearReq.RequestID,
	}
	client.SendEvent(clearReq)
	expect(<-server.mockStore.TwitchClearAuthInput.UserID).To.Equal("some-user-id")
	expect(client.ReadEvent()).To.Equal(expectedResp)
}

func TestTwitchUserDetails(t *testing.T) {
	expect := expect.New(t)

	server, cleanup := setupServer()
	defer cleanup()
	client, cleanup := setupClient(server.url)
	defer cleanup()
	authenticate(server, client)

	server.mockStore.TwitchStreamerAuthenticatedOutput.Authenticated <- true
	server.mockStore.TwitchStreamerAuthenticatedOutput.Err <- nil
	server.mockStore.TwitchStreamerCredentialsOutput.Username <- "streamer-user"
	server.mockStore.TwitchStreamerCredentialsOutput.Password <- ""
	server.mockStore.TwitchStreamerCredentialsOutput.TwitchUserID <- 0
	server.mockStore.TwitchStreamerCredentialsOutput.Err <- nil
	server.mockTwitchClient.StreamInfoOutput.Status <- "streamer-status"
	server.mockTwitchClient.StreamInfoOutput.Game <- "streamer-game"
	server.mockTwitchClient.StreamInfoOutput.Err <- nil
	server.mockStore.TwitchBotAuthenticatedOutput.Authenticated <- true
	server.mockStore.TwitchBotAuthenticatedOutput.Err <- nil
	server.mockStore.TwitchBotCredentialsOutput.Username <- "bot-user"
	server.mockStore.TwitchBotCredentialsOutput.Password <- ""
	server.mockStore.TwitchBotCredentialsOutput.TwitchUserID <- 0
	server.mockStore.TwitchBotCredentialsOutput.Err <- nil

	userDetailsReq := event{
		Cmd:       "twitch-user-details",
		RequestID: requestID(),
	}
	expectedResp := event{
		Cmd:       "twitch-user-details",
		RequestID: userDetailsReq.RequestID,
		Payload: map[string]interface{}{
			"streamer_authenticated": true,
			"streamer_username":      "streamer-user",
			"streamer_status":        "streamer-status",
			"streamer_game":          "streamer-game",
			"bot_authenticated":      true,
			"bot_username":           "bot-user",
		},
	}
	client.SendEvent(userDetailsReq)
	actual := client.ReadEvent()
	expect(actual).To.Equal(expectedResp)
}

func TestTwitchGamesHandler(t *testing.T) {
	expect := expect.New(t)

	server, cleanup := setupServer()
	defer cleanup()
	client, cleanup := setupClient(server.url)
	defer cleanup()
	authenticate(server, client)

	server.mockTwitchClient.GamesOutput.Games <- []twitch.Game{
		{
			ID:         1234,
			Name:       "some-game",
			Popularity: 4321,
			Image:      "some-image",
		},
	}

	gamesReq := event{
		Cmd:       "twitch-games",
		RequestID: requestID(),
	}
	expectedResp := event{
		Cmd:       "twitch-games",
		RequestID: gamesReq.RequestID,
		Payload: []interface{}{
			map[string]interface{}{
				"id":         float64(1234),
				"name":       "some-game",
				"popularity": float64(4321),
				"image":      "some-image",
			},
		},
	}
	client.SendEvent(gamesReq)
	actual := client.ReadEvent()
	expect(actual).To.Equal(expectedResp)
}

func TestTwitchStreamMessages(t *testing.T) {
	expect := expect.New(t)

	server, cleanup := setupServer()
	defer cleanup()
	client, cleanup := setupClient(server.url)
	defer cleanup()
	pub := setupPubSocket(server.subEndpoint)
	authenticate(server, client)
	authenticateTwitch(server)

	recentMessages := []stream.RXMessage{
		{
			Type: stream.Twitch,
			Twitch: &stream.RXTwitch{
				OwnerID: 12345,
				Line: &ircclient.Line{
					Cmd:  "test-cmd",
					Nick: "test-nick",
					Args: []string{
						"test-target",
						"test-body",
					},
				},
			},
		},
	}
	streamerMessage := stream.RXMessage{
		Type: stream.Twitch,
		Twitch: &stream.RXTwitch{
			OwnerID: 12345,
			Line: &ircclient.Line{
				Cmd:  "test-streamer-cmd",
				Nick: "test-streamer-nick",
				Args: []string{
					"test-streamer-target",
					"test-streamer-body",
				},
			},
		},
	}
	botMessageFromStreamer := stream.RXMessage{
		Type: stream.Twitch,
		Twitch: &stream.RXTwitch{
			OwnerID: 12345,
			Line: &ircclient.Line{
				Cmd:  "test-bot-cmd",
				Nick: "streamer-user",
				Args: []string{
					"test-bot-target",
					"test-bot-body",
				},
			},
		},
	}
	botMessageFromOtherUser := stream.RXMessage{
		Type: stream.Twitch,
		Twitch: &stream.RXTwitch{
			OwnerID: 12345,
			Line: &ircclient.Line{
				Cmd:  "test-bot-cmd",
				Nick: "test-bot-nick",
				Args: []string{
					"test-bot-target",
					"test-bot-body",
				},
			},
		},
	}

	server.mockStore.TwitchStreamerCredentialsOutput.Username <- "streamer-user"
	server.mockStore.TwitchStreamerCredentialsOutput.Password <- "streamer-pass"
	server.mockStore.TwitchStreamerCredentialsOutput.TwitchUserID <- 0
	server.mockStore.TwitchStreamerCredentialsOutput.Err <- nil
	server.mockStore.TwitchBotCredentialsOutput.Username <- "bot-user"
	server.mockStore.TwitchBotCredentialsOutput.Password <- "bot-pass"
	server.mockStore.TwitchBotCredentialsOutput.TwitchUserID <- 0
	server.mockStore.TwitchBotCredentialsOutput.Err <- nil
	server.mockStore.FetchRecentMessagesOutput.Msgs <- recentMessages
	server.mockStore.FetchRecentMessagesOutput.Err <- nil

	streamReq := event{
		Cmd:       "twitch-stream-messages",
		RequestID: requestID(),
	}
	client.SendEvent(streamReq)

	expectedTwitch := make(map[string]interface{})
	expectedTwitch["cmd"] = "test-cmd"
	expectedTwitch["nick"] = "test-nick"
	expectedTwitch["target"] = "test-target"
	expectedTwitch["body"] = "test-body"
	expectedTwitch["time"] = "0001-01-01T00:00:00Z"
	expectedTwitch["tags"] = nil
	expectedRecentMessage := make(map[string]interface{})
	expectedRecentMessage["type"] = float64(0)
	expectedRecentMessage["twitch"] = expectedTwitch
	expectedRecentMessage["discord"] = nil

	expected := event{
		Cmd:     "chat-message",
		Payload: expectedRecentMessage,
	}
	actual := client.ReadEvent()
	expect(actual).To.Equal(expected)

	expect(<-server.mockStreamManager.ConnectTwitchInput.User).To.Equal("streamer-user")
	expect(<-server.mockStreamManager.ConnectTwitchInput.Pass).To.Equal("oauth:streamer-pass")
	expect(<-server.mockStreamManager.ConnectTwitchInput.Channel).To.Equal("#streamer-user")
	expect(<-server.mockStreamManager.ConnectTwitchInput.User).To.Equal("bot-user")
	expect(<-server.mockStreamManager.ConnectTwitchInput.Pass).To.Equal("oauth:bot-pass")
	expect(<-server.mockStreamManager.ConnectTwitchInput.Channel).To.Equal("#streamer-user")

	streamerTopic := "twitch:streamer-user"
	streamerBytes, err := json.Marshal(streamerMessage)
	expect(err).To.Be.Nil()
	_, err = pub.SendMessage(streamerTopic, streamerBytes)
	expect(err).To.Be.Nil()

	expectedTwitch = make(map[string]interface{})
	expectedTwitch["cmd"] = "test-streamer-cmd"
	expectedTwitch["nick"] = "test-streamer-nick"
	expectedTwitch["target"] = "test-streamer-target"
	expectedTwitch["body"] = "test-streamer-body"
	expectedTwitch["time"] = "0001-01-01T00:00:00Z"
	expectedTwitch["tags"] = nil
	expectedStreamerMessage := make(map[string]interface{})
	expectedStreamerMessage["type"] = float64(0)
	expectedStreamerMessage["twitch"] = expectedTwitch
	expectedStreamerMessage["discord"] = nil
	expected = event{
		Cmd:     "chat-message",
		Payload: expectedStreamerMessage,
	}
	actual = client.ReadEvent()
	expect(actual).To.Equal(expected)

	botTopic := "twitch:bot-user"
	botBytes, err := json.Marshal(botMessageFromStreamer)
	expect(err).To.Be.Nil()
	_, err = pub.SendMessage(botTopic, botBytes)
	expect(err).To.Be.Nil()

	expectedTwitch = make(map[string]interface{})
	expectedTwitch["cmd"] = "test-bot-cmd"
	expectedTwitch["nick"] = "streamer-user"
	expectedTwitch["target"] = "test-bot-target"
	expectedTwitch["body"] = "test-bot-body"
	expectedTwitch["time"] = "0001-01-01T00:00:00Z"
	expectedTwitch["tags"] = nil
	expectedBotMessage := make(map[string]interface{})
	expectedBotMessage["type"] = float64(0)
	expectedBotMessage["twitch"] = expectedTwitch
	expectedBotMessage["discord"] = nil
	expected = event{
		Cmd:     "chat-message",
		Payload: expectedBotMessage,
	}
	actual = client.ReadEvent()
	expect(actual).To.Equal(expected)

	botTopic = "twitch:bot-user"
	botBytes, err = json.Marshal(botMessageFromOtherUser)
	expect(err).To.Be.Nil()
	_, err = pub.SendMessage(botTopic, botBytes)
	expect(err).To.Be.Nil()

	done := make(chan struct{})
	go func() {
		defer func() {
			recover()
			close(done)
		}()
		client.ReadEvent()
	}()

	select {
	case <-done:
		t.Error("received event from bot connection that was not the streamer")
	case <-time.After(50 * time.Millisecond):
	}
}

func TestTwitchSendMessage(t *testing.T) {
	expect := expect.New(t)

	server, cleanup := setupServer()
	defer cleanup()
	client, cleanup := setupClient(server.url)
	defer cleanup()
	authenticate(server, client)

	authenticateTwitch(server)
	server.mockStore.TwitchStreamerCredentialsOutput.Username <- "streamer-user"
	server.mockStore.TwitchStreamerCredentialsOutput.Password <- "streamer-pass"
	server.mockStore.TwitchStreamerCredentialsOutput.TwitchUserID <- 0
	server.mockStore.TwitchStreamerCredentialsOutput.Err <- nil
	server.mockStore.TwitchBotCredentialsOutput.Username <- "bot-user"
	server.mockStore.TwitchBotCredentialsOutput.Password <- "bot-pass"
	server.mockStore.TwitchBotCredentialsOutput.TwitchUserID <- 0
	server.mockStore.TwitchBotCredentialsOutput.Err <- nil

	sendMessageReq := event{
		Cmd:       "twitch-send-message",
		RequestID: requestID(),
		Payload: map[string]string{
			"user_type": "streamer",
			"message":   "test-streamer-message",
		},
	}
	client.SendEvent(sendMessageReq)
	expectedResp := event{
		Cmd:       "twitch-send-message",
		RequestID: sendMessageReq.RequestID,
	}
	expect(client.ReadEvent()).To.Equal(expectedResp)
	expect(<-server.mockStreamManager.ConnectTwitchInput.User).To.Equal("streamer-user")
	expect(<-server.mockStreamManager.ConnectTwitchInput.Pass).To.Equal("oauth:streamer-pass")
	expect(<-server.mockStreamManager.ConnectTwitchInput.Channel).To.Equal("#streamer-user")
	expect(<-server.mockStreamManager.SendInput.Msg).To.Equal(stream.TXMessage{
		Type: stream.Twitch,
		Twitch: &stream.TXTwitch{
			Username: "streamer-user",
			To:       "#streamer-user",
			Message:  "test-streamer-message",
		},
	})

	authenticateTwitch(server)
	server.mockStore.TwitchStreamerCredentialsOutput.Username <- "streamer-user"
	server.mockStore.TwitchStreamerCredentialsOutput.Password <- "streamer-pass"
	server.mockStore.TwitchStreamerCredentialsOutput.TwitchUserID <- 0
	server.mockStore.TwitchStreamerCredentialsOutput.Err <- nil
	server.mockStore.TwitchBotCredentialsOutput.Username <- "bot-user"
	server.mockStore.TwitchBotCredentialsOutput.Password <- "bot-pass"
	server.mockStore.TwitchBotCredentialsOutput.TwitchUserID <- 0
	server.mockStore.TwitchBotCredentialsOutput.Err <- nil

	sendMessageReq = event{
		Cmd:       "twitch-send-message",
		RequestID: requestID(),
		Payload: map[string]string{
			"user_type": "bot",
			"message":   "test-bot-message",
		},
	}
	client.SendEvent(sendMessageReq)
	expectedResp = event{
		Cmd:       "twitch-send-message",
		RequestID: sendMessageReq.RequestID,
	}
	expect(client.ReadEvent()).To.Equal(expectedResp)
	expect(<-server.mockStreamManager.ConnectTwitchInput.User).To.Equal("bot-user")
	expect(<-server.mockStreamManager.ConnectTwitchInput.Pass).To.Equal("oauth:bot-pass")
	expect(<-server.mockStreamManager.ConnectTwitchInput.Channel).To.Equal("#streamer-user")
	expect(<-server.mockStreamManager.SendInput.Msg).To.Equal(stream.TXMessage{
		Type: stream.Twitch,
		Twitch: &stream.TXTwitch{
			Username: "bot-user",
			To:       "#streamer-user",
			Message:  "test-bot-message",
		},
	})

	authenticateTwitch(server)
	server.mockStore.TwitchStreamerCredentialsOutput.Username <- "streamer-user"
	server.mockStore.TwitchStreamerCredentialsOutput.Password <- "streamer-pass"
	server.mockStore.TwitchStreamerCredentialsOutput.TwitchUserID <- 0
	server.mockStore.TwitchStreamerCredentialsOutput.Err <- nil
	server.mockStore.TwitchBotCredentialsOutput.Username <- "bot-user"
	server.mockStore.TwitchBotCredentialsOutput.Password <- "bot-pass"
	server.mockStore.TwitchBotCredentialsOutput.TwitchUserID <- 0
	server.mockStore.TwitchBotCredentialsOutput.Err <- nil

	sendMessageReq = event{
		Cmd:       "twitch-send-message",
		RequestID: requestID(),
		Payload: map[string]string{
			"user_type": "wrong-uer-type",
			"message":   "test-message",
		},
	}
	expectedResp = event{
		Cmd:       "twitch-send-message",
		RequestID: sendMessageReq.RequestID,
		Error: &eventErr{
			Code: 7,
			Text: "you specified an invalid user type",
		},
	}
	client.SendEvent(sendMessageReq)
	expect(client.ReadEvent()).To.Equal(expectedResp)
}

func TestTwitchUpdateChatDescription(t *testing.T) {
	expect := expect.New(t)

	server, cleanup := setupServer()
	defer cleanup()
	client, cleanup := setupClient(server.url)
	defer cleanup()
	authenticate(server, client)

	authenticateTwitch(server)
	server.mockStore.TwitchStreamerCredentialsOutput.Username <- "streamer-user"
	server.mockStore.TwitchStreamerCredentialsOutput.Password <- "streamer-pass"
	server.mockStore.TwitchStreamerCredentialsOutput.TwitchUserID <- 0
	server.mockStore.TwitchStreamerCredentialsOutput.Err <- nil
	server.mockTwitchClient.UpdateDescriptionOutput.Err <- nil

	sendMessageReq := event{
		Cmd:       "twitch-update-chat-description",
		RequestID: requestID(),
		Payload: map[string]string{
			"status": "test-status",
			"game":   "test-game",
		},
	}
	client.SendEvent(sendMessageReq)
	expect(<-server.mockTwitchClient.UpdateDescriptionInput.Game).To.Equal("test-game")
	expect(<-server.mockTwitchClient.UpdateDescriptionInput.Status).To.Equal("test-status")
	expect(<-server.mockTwitchClient.UpdateDescriptionInput.Channel).To.Equal("streamer-user")
	expect(<-server.mockTwitchClient.UpdateDescriptionInput.Token).To.Equal("streamer-pass")
	expectedResp := event{
		Cmd:       "twitch-update-chat-description",
		RequestID: sendMessageReq.RequestID,
	}
	expect(client.ReadEvent()).To.Equal(expectedResp)
}

func TestRequestFailsWhenNotAuthedWithTwitch(t *testing.T) {
	expect := expect.New(t)

	server, cleanup := setupServer()
	defer cleanup()
	client, cleanup := setupClient(server.url)
	defer cleanup()
	authenticate(server, client)

	server.mockStore.TwitchAuthenticatedOutput.Authenticated <- false
	server.mockStore.TwitchAuthenticatedOutput.Err <- nil
	sendMessageReq := event{
		Cmd:       "twitch-update-chat-description",
		RequestID: requestID(),
		Payload: map[string]string{
			"status": "test-status",
			"game":   "test-game",
		},
	}
	client.SendEvent(sendMessageReq)
	expectedResp := event{
		Cmd:       "twitch-update-chat-description",
		RequestID: sendMessageReq.RequestID,
		Error: &eventErr{
			Code: 5,
			Text: "authentication error with twitch",
		},
	}
	expect(client.ReadEvent()).To.Equal(expectedResp)
}
