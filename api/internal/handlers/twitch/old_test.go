package twitch_test

// func TestTwitchSendMessage(t *testing.T) {
// 	expect := expect.New(t)

// 	server, cleanup := setupServer()
// 	defer cleanup()
// 	client, cleanup := setupClient(server.url)
// 	defer cleanup()
// 	authenticate(server, client)

// 	authenticateTwitch(server)
// 	server.mockStore.TwitchCredentialsOutput.Creds <- store.TwitchCredentials{
// 		StreamerUsername: "streamer-user",
// 		StreamerPassword: "streamer-pass",
// 		BotUsername:      "bot-user",
// 		BotPassword:      "bot-pass",
// 	}
// 	server.mockStore.TwitchCredentialsOutput.Err <- nil

// 	sendMessageReq := event{
// 		Cmd:       "twitch-send-message",
// 		RequestID: requestID(),
// 		Payload: map[string]string{
// 			"user_type": "streamer",
// 			"message":   "test-streamer-message",
// 		},
// 	}
// 	client.SendEvent(sendMessageReq)
// 	expectedResp := event{
// 		Cmd:       "twitch-send-message",
// 		RequestID: sendMessageReq.RequestID,
// 	}
// 	expect(client.ReadEvent()).To.Equal(expectedResp)
// 	expect(<-server.mockStreamManager.ConnectTwitchInput.User).To.Equal("streamer-user")
// 	expect(<-server.mockStreamManager.ConnectTwitchInput.Pass).To.Equal("oauth:streamer-pass")
// 	expect(<-server.mockStreamManager.ConnectTwitchInput.Channel).To.Equal("#streamer-user")
// 	expect(<-server.mockStreamManager.SendInput.Msg).To.Equal(stream.TXMessage{
// 		Type: stream.Twitch,
// 		Twitch: &stream.TXTwitch{
// 			Username: "streamer-user",
// 			To:       "#streamer-user",
// 			Message:  "test-streamer-message",
// 		},
// 	})

// 	authenticateTwitch(server)
// 	server.mockStore.TwitchCredentialsOutput.Creds <- store.TwitchCredentials{
// 		StreamerUsername: "streamer-user",
// 		StreamerPassword: "streamer-pass",
// 		BotUsername:      "bot-user",
// 		BotPassword:      "bot-pass",
// 	}
// 	server.mockStore.TwitchCredentialsOutput.Err <- nil

// 	sendMessageReq = event{
// 		Cmd:       "twitch-send-message",
// 		RequestID: requestID(),
// 		Payload: map[string]string{
// 			"user_type": "bot",
// 			"message":   "test-bot-message",
// 		},
// 	}
// 	client.SendEvent(sendMessageReq)
// 	expectedResp = event{
// 		Cmd:       "twitch-send-message",
// 		RequestID: sendMessageReq.RequestID,
// 	}
// 	expect(client.ReadEvent()).To.Equal(expectedResp)
// 	expect(<-server.mockStreamManager.ConnectTwitchInput.User).To.Equal("bot-user")
// 	expect(<-server.mockStreamManager.ConnectTwitchInput.Pass).To.Equal("oauth:bot-pass")
// 	expect(<-server.mockStreamManager.ConnectTwitchInput.Channel).To.Equal("#streamer-user")
// 	expect(<-server.mockStreamManager.SendInput.Msg).To.Equal(stream.TXMessage{
// 		Type: stream.Twitch,
// 		Twitch: &stream.TXTwitch{
// 			Username: "bot-user",
// 			To:       "#streamer-user",
// 			Message:  "test-bot-message",
// 		},
// 	})

// 	authenticateTwitch(server)
// 	server.mockStore.TwitchCredentialsOutput.Creds <- store.TwitchCredentials{
// 		StreamerUsername: "streamer-user",
// 		StreamerPassword: "streamer-pass",
// 		BotUsername:      "bot-user",
// 		BotPassword:      "bot-pass",
// 	}
// 	server.mockStore.TwitchCredentialsOutput.Err <- nil

// 	sendMessageReq = event{
// 		Cmd:       "twitch-send-message",
// 		RequestID: requestID(),
// 		Payload: map[string]string{
// 			"user_type": "wrong-uer-type",
// 			"message":   "test-message",
// 		},
// 	}
// 	expectedResp = event{
// 		Cmd:       "twitch-send-message",
// 		RequestID: sendMessageReq.RequestID,
// 		Error: &eventErr{
// 			Code: 7,
// 			Text: "you specified an invalid user type",
// 		},
// 	}
// 	client.SendEvent(sendMessageReq)
// 	expect(client.ReadEvent()).To.Equal(expectedResp)
// }
