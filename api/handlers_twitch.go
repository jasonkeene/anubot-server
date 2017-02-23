package api

import (
	"log"

	"github.com/jasonkeene/anubot-server/store"
	"github.com/jasonkeene/anubot-server/stream"
	"github.com/jasonkeene/anubot-server/twitch/oauth"
)

// twitchOauthStartHandler responds with a URL to start the Twitch oauth flow.
// The streamer user is required to be the first to begin the oauth flow,
// followed by the bot user.
func twitchOauthStartHandler(e event, s *session) {
	resp, send := setup(e, s)
	defer send()

	tus, ok := e.Payload.(string)
	if !ok {
		resp.Error = invalidPayload
		return
	}

	userID, _ := s.Authenticated()

	var tu store.TwitchUser
	switch tus {
	case "streamer":
		tu = store.Streamer
	case "bot":
		tu = store.Bot
		creds, err := s.Store().TwitchCredentials(userID)
		if err != nil {
			return
		}
		if !creds.StreamerAuthenticated {
			resp.Error = twitchOauthStartOrderError
			return
		}
	default:
		resp.Error = invalidPayload
		return
	}

	nonce := s.api.nonceGen()
	err := s.Store().StoreOauthNonce(userID, tu, nonce)
	if err != nil {
		log.Printf("got an err trying to store oauth nonce: %s", err)
		return
	}
	url := oauth.URL(s.TwitchOauthClientID(), userID, tu, nonce)

	s.api.twitchOauthCallbacks.RegisterCompletionCallback(nonce, func() {
		resp := event{
			Cmd:     "twitch-oauth-complete",
			Payload: tus,
		}
		err := s.Send(resp)
		if err != nil {
			log.Printf("unable to tx: %s", err)
		}
	})
	resp.Payload = url
	resp.Error = nil
}

// twitchClearAuthHandler clears all auth data for the user.
func twitchClearAuthHandler(e event, s *session) {
	resp, send := setup(e, s)
	defer send()

	err := s.Store().TwitchClearAuth(s.userID)
	if err != nil {
		return
	}
	resp.Error = nil
}

// twitchUserDetailsHandler provides information on the Twitch streamer and
// bot users.
func twitchUserDetailsHandler(e event, s *session) {
	resp, send := setup(e, s)
	defer send()

	p := map[string]interface{}{
		"streamer_authenticated": false,
		"streamer_username":      "",
		"streamer_status":        "",
		"streamer_game":          "",

		"bot_authenticated": false,
		"bot_username":      "",
	}

	creds, err := s.Store().TwitchCredentials(s.userID)
	if err != nil {
		log.Printf("unable to authenticate: %s", err)
		return
	}
	if !creds.StreamerAuthenticated {
		resp.Payload = p
		resp.Error = nil
		return
	}

	status, game, err := s.api.twitchClient.StreamInfo(creds.StreamerUsername)
	if err != nil {
		log.Printf("unable to fetch stream info for user: %s: %s", creds.StreamerUsername, err)
		return
	}
	p["streamer_authenticated"] = creds.StreamerAuthenticated
	p["streamer_username"] = creds.StreamerUsername
	p["streamer_status"] = status
	p["streamer_game"] = game

	if !creds.BotAuthenticated {
		resp.Payload = p
		resp.Error = nil
		return
	}

	p["bot_authenticated"] = creds.BotAuthenticated
	p["bot_username"] = creds.BotUsername
	resp.Payload = p
	resp.Error = nil
}

// twitchGamesHandler returns the available games.
func twitchGamesHandler(e event, s *session) {
	resp, send := setup(e, s)
	defer send()

	resp.Payload = s.api.twitchClient.Games()
	resp.Error = nil
}

// twitchStreamMessagesHandler writes chat messages to websocket connection.
func twitchStreamMessagesHandler(e event, s *session) {
	creds, err := s.Store().TwitchCredentials(s.userID)
	if err != nil {
		log.Printf("unable to get creds: %s", err)
		return
	}

	recent, err := s.Store().FetchRecentMessages(s.userID)
	if err == nil {
		for _, msg := range recent {
			if msg.Type != stream.Twitch {
				continue
			}
			if msg.Twitch.OwnerID == creds.BotTwitchUserID && !userMessage(&msg, creds.StreamerUsername) {
				continue
			}

			err = s.Send(event{
				Cmd: "chat-message",
				Payload: message{
					Type: msg.Type,
					Twitch: &twitchMessage{
						Cmd:    msg.Twitch.Line.Cmd,
						Nick:   msg.Twitch.Line.Nick,
						Target: msg.Twitch.Line.Args[0],
						Body:   msg.Twitch.Line.Args[1],
						Time:   msg.Twitch.Line.Time,
						Tags:   msg.Twitch.Line.Tags,
					},
				},
			})
			if err != nil {
				log.Printf("unable to tx: %s", err)
			}
		}
	}

	s.api.streamManager.ConnectTwitch(
		creds.StreamerUsername,
		"oauth:"+creds.StreamerPassword,
		"#"+creds.StreamerUsername,
	)
	s.api.streamManager.ConnectTwitch(
		creds.BotUsername,
		"oauth:"+creds.BotPassword,
		"#"+creds.StreamerUsername,
	)

	mw, err := newMessageWriter(
		creds.StreamerUsername,
		"twitch:"+creds.StreamerUsername,
		"twitch:"+creds.BotUsername,
		s.api.subEndpoints,
		s,
	)
	if err != nil {
		log.Printf("unable to stream messages: %s", err)
		return
	}
	go mw.startStreamer()
	go mw.startBot()
}

// twitchSendMessageHandler accepts messages to send via Twitch chat.
func twitchSendMessageHandler(e event, s *session) {
	resp, send := setup(e, s)
	defer send()

	data, ok := e.Payload.(map[string]interface{})
	if !ok {
		resp.Error = invalidPayload
		return
	}
	userType, ok := data["user_type"].(string)
	if !ok {
		resp.Error = invalidPayload
		return
	}
	message, ok := data["message"].(string)
	if !ok {
		resp.Error = invalidPayload
		return
	}

	creds, err := s.Store().TwitchCredentials(s.userID)
	if err != nil {
		log.Printf("unable to get creds: %s", err)
		return
	}

	var username, password string
	switch userType {
	case "streamer":
		username, password = creds.StreamerUsername, creds.StreamerPassword
	case "bot":
		username, password = creds.BotUsername, creds.BotPassword
	default:
		resp.Error = invalidTwitchUserType
		return
	}
	s.api.streamManager.ConnectTwitch(
		username,
		"oauth:"+password,
		"#"+creds.StreamerUsername,
	)
	s.api.streamManager.Send(stream.TXMessage{
		Type: stream.Twitch,
		Twitch: &stream.TXTwitch{
			Username: username,
			To:       "#" + creds.StreamerUsername,
			Message:  message,
		},
	})
	resp.Error = nil
}

// twitchUpdateChatDescriptionHandler updates the chat description for Twitch.
func twitchUpdateChatDescriptionHandler(e event, s *session) {
	resp, send := setup(e, s)
	defer send()

	payload, ok := e.Payload.(map[string]interface{})
	if !ok {
		resp.Error = invalidPayload
		return
	}
	status, ok := payload["status"].(string)
	if !ok {
		resp.Error = invalidPayload
		return
	}
	game, ok := payload["game"].(string)
	if !ok {
		resp.Error = invalidPayload
		return
	}

	creds, err := s.Store().TwitchCredentials(s.userID)
	if err != nil {
		log.Printf("unable to get creds: %s", err)
		return
	}

	err = s.api.twitchClient.UpdateDescription(
		status,
		game,
		creds.StreamerUsername,
		creds.StreamerPassword,
	)
	if err != nil {
		log.Println("unable to update chat description, got error:", err)
		return
	}
	resp.Error = nil
}

// twitchAuthenticateWrapper wraps a handler and makes sure the user attached
// to the session is properly authenticated with twitch.
func twitchAuthenticateWrapper(f eventHandler) eventHandler {
	return func(e event, s *session) {
		userID, _ := s.Authenticated()

		creds, err := s.Store().TwitchCredentials(userID)

		if err != nil {
			err := s.Send(event{
				Cmd:       e.Cmd,
				RequestID: e.RequestID,
				Error:     unknownError,
			})
			if err != nil {
				log.Printf("unable to tx: %s", err)
			}
			return
		}
		if creds.StreamerAuthenticated && creds.BotAuthenticated {
			f(e, s)
			return
		}
		err = s.Send(event{
			Cmd:       e.Cmd,
			RequestID: e.RequestID,
			Error:     twitchAuthenticationError,
		})
		if err != nil {
			log.Printf("unable to tx: %s", err)
		}
	}
}
