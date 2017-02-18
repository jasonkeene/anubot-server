package api

import (
	"anubot/store"
	"anubot/stream"
	"anubot/twitch/oauth"
	"log"
)

// twitchOauthStartHandler responds with a URL to start the Twitch oauth flow.
// The streamer user is required to be the first to begin the oauth flow,
// followed by the bot user.
func twitchOauthStartHandler(e event, s *session) {
	tus, ok := e.Payload.(string)
	if !ok {
		s.Send(event{
			Cmd:       "twitch-oauth-start",
			RequestID: e.RequestID,
			Error:     invalidPayload,
		})
		return
	}

	userID, _ := s.Authenticated()

	var tu store.TwitchUser
	switch tus {
	case "streamer":
		tu = store.Streamer
	case "bot":
		tu = store.Bot
		if !s.Store().TwitchStreamerAuthenticated(userID) {
			s.Send(event{
				Cmd:       "twitch-oauth-start",
				RequestID: e.RequestID,
				Error:     twitchOauthStartOrderError,
			})
			return
		}
	default:
		s.Send(event{
			Cmd:       "twitch-oauth-start",
			RequestID: e.RequestID,
			Error:     invalidPayload,
		})
		return
	}

	url, err := oauth.URL(s.TwitchOauthClientID(), userID, tu, s.Store())
	if err != nil {
		log.Printf("got an err trying to create oauth url: %s", err)
		s.Send(event{
			Cmd:       "twitch-oauth-start",
			RequestID: e.RequestID,
			Error:     unknownError,
		})
		return
	}

	s.Send(event{
		Cmd:       "twitch-oauth-start",
		RequestID: e.RequestID,
		Payload:   url,
	})
}

// twitchClearAuthHandler clears all auth data for the user.
func twitchClearAuthHandler(e event, s *session) {
	s.Store().TwitchClearAuth(s.userID)
	s.Send(event{
		Cmd:       "twitch-clear-auth",
		RequestID: e.RequestID,
	})
}

// twitchUserDetailsHandler provides information on the Twitch streamer and
// bot users.
func twitchUserDetailsHandler(e event, s *session) {
	p := map[string]interface{}{
		"streamer_authenticated": false,
		"streamer_username":      "",
		"streamer_status":        "",
		"streamer_game":          "",

		"bot_authenticated": false,
		"bot_username":      "",
	}
	resp := event{
		Cmd:       "twitch-user-details",
		RequestID: e.RequestID,
		Payload:   p,
	}
	defer s.Send(resp)

	streamerAuthenticated := s.Store().TwitchStreamerAuthenticated(s.userID)
	if !streamerAuthenticated {
		return
	}
	streamerUsername, _, _ := s.Store().TwitchStreamerCredentials(s.userID)
	status, game, err := s.api.twitchClient.StreamInfo(streamerUsername)
	if err != nil {
		log.Printf("unable to fetch stream info for user %s: %s",
			streamerUsername, err)
		return
	}
	p["streamer_authenticated"] = streamerAuthenticated
	p["streamer_username"] = streamerUsername
	p["streamer_status"] = status
	p["streamer_game"] = game

	botAuthenticated := s.Store().TwitchBotAuthenticated(s.userID)
	if !botAuthenticated {
		return
	}

	p["bot_authenticated"] = botAuthenticated
	p["bot_username"], _, _ = s.Store().TwitchBotCredentials(s.userID)
}

// twitchGamesHandler returns the available games.
func twitchGamesHandler(e event, s *session) {
	s.Send(event{
		Cmd:       e.Cmd,
		RequestID: e.RequestID,
		Payload:   s.api.twitchClient.Games(),
	})
}

// twitchStreamMessagesHandler writes chat messages to websocket connection.
func twitchStreamMessagesHandler(e event, s *session) {
	streamerUsername, streamerPassword, _ := s.Store().TwitchStreamerCredentials(s.userID)
	botUsername, botPassword, botID := s.Store().TwitchBotCredentials(s.userID)

	recent, err := s.Store().FetchRecentMessages(s.userID)
	if err == nil {
		for _, msg := range recent {
			if msg.Type != stream.Twitch {
				continue
			}
			if msg.Twitch.OwnerID == botID && !userMessage(&msg, streamerUsername) {
				continue
			}

			s.Send(event{
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
		}
	}

	s.api.streamManager.ConnectTwitch(streamerUsername, "oauth:"+streamerPassword, "#"+streamerUsername)
	s.api.streamManager.ConnectTwitch(botUsername, "oauth:"+botPassword, "#"+streamerUsername)

	mw, err := newMessageWriter(
		streamerUsername,
		"twitch:"+streamerUsername,
		"twitch:"+botUsername,
		s.api.subEndpoints,
		s.ws,
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
	data, ok := e.Payload.(map[string]interface{})
	if !ok {
		s.Send(event{
			Cmd:   e.Cmd,
			Error: invalidPayload,
		})
		return
	}
	userType, ok := data["user_type"].(string)
	if !ok {
		s.Send(event{
			Cmd:   e.Cmd,
			Error: invalidPayload,
		})
		return
	}
	message, ok := data["message"].(string)
	if !ok {
		s.Send(event{
			Cmd:   e.Cmd,
			Error: invalidPayload,
		})
		return
	}

	streamerUsername, streamerPassword, _ := s.Store().TwitchStreamerCredentials(s.userID)
	var username, password string
	switch userType {
	case "streamer":
		username, password = streamerUsername, streamerPassword
	case "bot":
		username, password, _ = s.Store().TwitchBotCredentials(s.userID)
	default:
		s.Send(event{
			Cmd:   e.Cmd,
			Error: invalidTwitchUserType,
		})
		return
	}
	s.api.streamManager.ConnectTwitch(username, "oauth:"+password, "#"+username)
	s.api.streamManager.Send(stream.TXMessage{
		Type: stream.Twitch,
		Twitch: &stream.TXTwitch{
			Username: username,
			To:       "#" + streamerUsername,
			Message:  message,
		},
	})
}

// twitchUpdateChatDescriptionHandler updates the chat description for Twitch.
func twitchUpdateChatDescriptionHandler(e event, s *session) {
	payload, ok := e.Payload.(map[string]interface{})
	if !ok {
		s.Send(event{
			Cmd:   e.Cmd,
			Error: invalidPayload,
		})
		return
	}
	status, ok := payload["status"].(string)
	if !ok {
		s.Send(event{
			Cmd:   e.Cmd,
			Error: invalidPayload,
		})
		return
	}
	game, ok := payload["game"].(string)
	if !ok {
		s.Send(event{
			Cmd:   e.Cmd,
			Error: invalidPayload,
		})
		return
	}

	user, pass, _ := s.Store().TwitchStreamerCredentials(s.userID)
	err := s.api.twitchClient.UpdateDescription(status, game, user, pass)
	if err != nil {
		log.Println("unable to update chat description, got error:", err)
		s.Send(event{
			Cmd:   e.Cmd,
			Error: unknownError,
		})
	}
}

// twitchAuthenticateWrapper wraps a handler and makes sure the user attached
// to the session is properly authenticated with twitch.
func twitchAuthenticateWrapper(f handlerFunc) handlerFunc {
	return func(e event, s *session) {
		userID, _ := s.Authenticated()
		if s.Store().TwitchAuthenticated(userID) {
			f(e, s)
			return
		}
		s.Send(event{
			Cmd:       e.Cmd,
			RequestID: e.RequestID,
			Error:     twitchAuthenticationError,
		})
	}
}
