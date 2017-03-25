package twitch_test

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/a8m/expect"
	"github.com/fluffle/goirc/client"
	"github.com/jasonkeene/anubot-server/api/internal/handlers"
	"github.com/jasonkeene/anubot-server/api/internal/handlers/twitch"
	"github.com/jasonkeene/anubot-server/store"
	"github.com/jasonkeene/anubot-server/stream"
	"github.com/pebbe/zmq4"
)

func TestRecentMessageStreaming(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{
		userID: "test-user-id",
	}
	now := time.Now()
	spyMessagesStore := &SpyStreamMessagesStore{
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
		recentMessages: []stream.RXMessage{
			{
				Type: stream.Twitch,
				Twitch: &stream.RXTwitch{
					OwnerID: 12345,
					Line: &client.Line{
						Cmd:  "PRIVMSG",
						Nick: "test-nick",
						Args: []string{
							"test-target",
							"message-received-by-streamer",
						},
						Time: now,
						Tags: map[string]string{
							"test": "tag",
						},
					},
				},
			},
			{
				Type: stream.Twitch,
				Twitch: &stream.RXTwitch{
					OwnerID: 54321,
					Line: &client.Line{
						Cmd:  "PRIVMSG",
						Nick: "test-nick",
						Args: []string{
							"test-target",
							"message-received-by-bot",
						},
						Time: now,
						Tags: map[string]string{
							"test": "tag",
						},
					},
				},
			},
			{
				Type: stream.Twitch,
				Twitch: &stream.RXTwitch{
					OwnerID: 54321,
					Line: &client.Line{
						Cmd:  "PRIVMSG",
						Nick: "test-streamer-username",
						Args: []string{
							"test-target",
							"message-received-by-bot-from-streamer",
						},
						Time: now,
						Tags: map[string]string{
							"test": "tag",
						},
					},
				},
			},
		},
	}
	spyConnector := &SpyConnector{}
	handler := twitch.NewStreamMessagesHandler(
		spyMessagesStore,
		spyConnector,
		[]string{},
	)
	event := handlers.Event{
		Cmd:       "twitch-stream-messages",
		RequestID: "test-request-id",
	}

	handler.HandleEvent(event, spySession)

	expected := []handlers.Event{
		{
			Cmd:       "chat-message",
			RequestID: "test-request-id",
			Payload: twitch.Message{
				Type: stream.Twitch,
				Twitch: &twitch.TMessage{
					Cmd:    "PRIVMSG",
					Nick:   "test-nick",
					Target: "test-target",
					Body:   "message-received-by-streamer",
					Time:   now,
					Tags: map[string]string{
						"test": "tag",
					},
				},
			},
		},
		{
			Cmd:       "chat-message",
			RequestID: "test-request-id",
			Payload: twitch.Message{
				Type: stream.Twitch,
				Twitch: &twitch.TMessage{
					Cmd:    "PRIVMSG",
					Nick:   "test-streamer-username",
					Target: "test-target",
					Body:   "message-received-by-bot-from-streamer",
					Time:   now,
					Tags: map[string]string{
						"test": "tag",
					},
				},
			},
		},
	}
	expectedConnects := []connectArgs{
		{
			user:    "test-streamer-username",
			pass:    "oauth:test-streamer-password",
			channel: "#test-streamer-username",
		},
		{
			user:    "test-bot-username",
			pass:    "oauth:test-bot-password",
			channel: "#test-streamer-username",
		},
	}
	expect(spyConnector.connectCalls).To.Equal(expectedConnects)
	expect(spyMessagesStore.credsCalledWith).To.Equal("test-user-id")
	expect(spyMessagesStore.recentMessagesCalledWith).To.Equal("test-user-id")
	expect(spySession.sendCalls()).To.Equal(expected)
}

func TestStreamingNewMessages(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{
		userID: "test-user-id",
	}
	endpoint := endpoint()
	pub := setupPubSocket(endpoint)

	spyMessagesStore := &SpyStreamMessagesStore{
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
	handler := twitch.NewStreamMessagesHandler(
		spyMessagesStore,
		&SpyConnector{},
		[]string{endpoint},
	)
	event := handlers.Event{
		Cmd:       "twitch-stream-messages",
		RequestID: "test-request-id",
	}

	handler.HandleEvent(event, spySession)

	streamerMessage := stream.RXMessage{
		Type: stream.Twitch,
		Twitch: &stream.RXTwitch{
			OwnerID: 12345,
			Line: &client.Line{
				Cmd:  "test-cmd",
				Nick: "test-nick",
				Args: []string{
					"test-target",
					"message-received-by-streamer",
				},
			},
		},
	}
	botMessage := stream.RXMessage{
		Type: stream.Twitch,
		Twitch: &stream.RXTwitch{
			OwnerID: 54321,
			Line: &client.Line{
				Cmd:  "test-cmd",
				Nick: "test--nick",
				Args: []string{
					"test-target",
					"message-received-by-bot",
				},
			},
		},
	}
	botMessageFromStreamer := stream.RXMessage{
		Type: stream.Twitch,
		Twitch: &stream.RXTwitch{
			OwnerID: 54321,
			Line: &client.Line{
				Cmd:  "test-cmd",
				Nick: "test-streamer-username",
				Args: []string{
					"test-target",
					"message-received-by-bot-from-streamer",
				},
			},
		},
	}

	streamerBytes, err := json.Marshal(streamerMessage)
	expect(err).To.Be.Nil()
	botBytes, err := json.Marshal(botMessage)
	expect(err).To.Be.Nil()
	botBytesFromStreamer, err := json.Marshal(botMessageFromStreamer)
	expect(err).To.Be.Nil()

	_, err = pub.SendMessage("twitch:test-streamer-username", streamerBytes)
	expect(err).To.Be.Nil()
	_, err = pub.SendMessage("twitch:test-bot-username", botBytes)
	expect(err).To.Be.Nil()
	_, err = pub.SendMessage("twitch:test-bot-username", botBytesFromStreamer)
	expect(err).To.Be.Nil()

	for {
		if len(spySession.sendCalls()) == 2 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	expected := []handlers.Event{
		{
			Cmd:       "chat-message",
			RequestID: "test-request-id",
			Payload: twitch.Message{
				Type: stream.Twitch,
				Twitch: &twitch.TMessage{
					Cmd:    "test-cmd",
					Nick:   "test-nick",
					Target: "test-target",
					Body:   "message-received-by-streamer",
				},
			},
		},
		{
			Cmd:       "chat-message",
			RequestID: "test-request-id",
			Payload: twitch.Message{
				Type: stream.Twitch,
				Twitch: &twitch.TMessage{
					Cmd:    "test-cmd",
					Nick:   "test-streamer-username",
					Target: "test-target",
					Body:   "message-received-by-bot-from-streamer",
				},
			},
		},
	}
	expect(spySession.sendCalls()).To.Equal(expected)
}

func endpoint() string {
	return "inproc://test-message-streaming-pub-" + randString()
}

func randString() string {
	b := make([]byte, 20)
	_, err := rand.Read(b)
	if err != nil {
		log.Panicf("unable to read randomness %s:", err)
	}
	return fmt.Sprintf("%x", b)
}

func setupPubSocket(endpoint string) *zmq4.Socket {
	pub, err := zmq4.NewSocket(zmq4.PUB)
	if err != nil {
		log.Panicf("unable to create pub socket %s:", err)
	}
	err = pub.Bind(endpoint)
	if err != nil {
		log.Panicf("unable to connect pub socket to endpoint: %s", err)
	}
	return pub
}

func TestSendingMessage(t *testing.T) {
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
	spyStreamManager := &SpyStreamManager{}
	handler := twitch.NewSendMessageHandler(
		spyCredsProvider,
		spyStreamManager,
	)

	cases := map[string]struct {
		event   handlers.Event
		user    string
		pass    string
		channel string
		message stream.TXMessage
	}{
		"streamer message": {
			event: handlers.Event{
				Cmd:       "twitch-send-message",
				RequestID: "test-request-id",
				Payload: map[string]interface{}{
					"user_type": "streamer",
					"message":   "test-streamer-message",
				},
			},
			user:    "test-streamer-username",
			pass:    "oauth:test-streamer-password",
			channel: "#test-streamer-username",
			message: stream.TXMessage{
				Type: stream.Twitch,
				Twitch: &stream.TXTwitch{
					Username: "test-streamer-username",
					To:       "#test-streamer-username",
					Message:  "test-streamer-message",
				},
			},
		},
		"bot message": {
			event: handlers.Event{
				Cmd:       "twitch-send-message",
				RequestID: "test-request-id",
				Payload: map[string]interface{}{
					"user_type": "bot",
					"message":   "test-bot-message",
				},
			},
			user:    "test-bot-username",
			pass:    "oauth:test-bot-password",
			channel: "#test-streamer-username",
			message: stream.TXMessage{
				Type: stream.Twitch,
				Twitch: &stream.TXTwitch{
					Username: "test-bot-username",
					To:       "#test-streamer-username",
					Message:  "test-bot-message",
				},
			},
		},
	}

	for _, testCase := range cases {
		handler.HandleEvent(testCase.event, spySession)

		expect(spyStreamManager.connectCalledWithUser).To.Equal(testCase.user)
		expect(spyStreamManager.connectCalledWithPass).To.Equal(testCase.pass)
		expect(spyStreamManager.connectCalledWithChannel).To.Equal(testCase.channel)
		expect(spyStreamManager.messageSent).To.Equal(testCase.message)
	}
}

func TestSendingInvalidMessages(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{}
	handler := twitch.NewSendMessageHandler(
		&SpyCredentialsProvider{},
		&SpyStreamManager{},
	)

	cases := map[string]interface{}{
		"nil payload": nil,
		"invalid user_type type": map[string]interface{}{
			"user_type": 12345,
		},
		"invalid message": map[string]interface{}{
			"user_type": "streamer",
			"message":   12345,
		},
		"invalid user_type": map[string]interface{}{
			"user_type": "something-odd",
			"message":   "test-message",
		},
	}

	for _, payload := range cases {
		event := handlers.Event{
			Cmd:       "twitch-send-message",
			RequestID: "test-request-id",
			Payload:   payload,
		}
		handler.HandleEvent(event, spySession)
		expected := handlers.Event{
			Cmd:       "twitch-send-message",
			RequestID: "test-request-id",
			Error:     handlers.InvalidPayload,
		}
		expect(spySession.sendCalledWith).To.Equal(expected)
	}
}

func TestWhenCredentialsProviderErrors(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{}
	handler := twitch.NewSendMessageHandler(
		&SpyCredentialsProvider{
			err: errors.New("test-error"),
		},
		&SpyStreamManager{},
	)
	event := handlers.Event{
		Cmd:       "twitch-send-message",
		RequestID: "test-request-id",
		Payload: map[string]interface{}{
			"user_type": "streamer",
			"message":   "test-message",
		},
	}

	handler.HandleEvent(event, spySession)

	expected := handlers.Event{
		Cmd:       "twitch-send-message",
		RequestID: "test-request-id",
		Error:     handlers.UnknownError,
	}
	expect(spySession.sendCalledWith).To.Equal(expected)
}
