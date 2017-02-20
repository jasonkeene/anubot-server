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
		authenticated, err := s.Store().TwitchStreamerAuthenticated(userID)
		if err != nil {
			return
		}
		if !authenticated {
			resp.Error = twitchOauthStartOrderError
			return
		}
	default:
		resp.Error = invalidPayload
		return
	}

	url, err := oauth.URL(s.TwitchOauthClientID(), userID, tu, s.Store())
	if err != nil {
		log.Printf("got an err trying to create oauth url: %s", err)
		return
	}

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

	streamerAuthenticated, err := s.Store().TwitchStreamerAuthenticated(s.userID)
	if err != nil {
		log.Printf("unable to authenticate streamer: %s", err)
		return
	}
	if !streamerAuthenticated {
		return
	}
	streamerUsername, _, _, err := s.Store().TwitchStreamerCredentials(s.userID)
	if err != nil {
		log.Printf("unable to get streamer creds: %s", err)
		return
	}
	status, game, err := s.api.twitchClient.StreamInfo(streamerUsername)
	if err != nil {
		log.Printf("unable to fetch stream info for user: %s: %s", streamerUsername, err)
		return
	}
	p["streamer_authenticated"] = streamerAuthenticated
	p["streamer_username"] = streamerUsername
	p["streamer_status"] = status
	p["streamer_game"] = game

	botAuthenticated, err := s.Store().TwitchBotAuthenticated(s.userID)
	if err != nil {
		log.Printf("unable to authenticate bot: %s", err)
		return
	}
	if !botAuthenticated {
		resp.Payload = p
		return
	}

	botUsername, _, _, err := s.Store().TwitchBotCredentials(s.userID)
	if err != nil {
		log.Printf("unable to get bot creds: %s", err)
		return
	}
	p["bot_authenticated"] = botAuthenticated
	p["bot_username"] = botUsername
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
	streamerUsername, streamerPassword, _, err := s.Store().TwitchStreamerCredentials(s.userID)
	if err != nil {
		log.Printf("unable to get streamer creds: %s", err)
		return
	}
	botUsername, botPassword, botID, err := s.Store().TwitchBotCredentials(s.userID)
	if err != nil {
		log.Printf("unable to get bot creds: %s", err)
		return
	}

	recent, err := s.Store().FetchRecentMessages(s.userID)
	if err == nil {
		for _, msg := range recent {
			if msg.Type != stream.Twitch {
				continue
			}
			if msg.Twitch.OwnerID == botID && !userMessage(&msg, streamerUsername) {
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

	s.api.streamManager.ConnectTwitch(streamerUsername, "oauth:"+streamerPassword, "#"+streamerUsername)
	s.api.streamManager.ConnectTwitch(botUsername, "oauth:"+botPassword, "#"+streamerUsername)

	mw, err := newMessageWriter(
		streamerUsername,
		"twitch:"+streamerUsername,
		"twitch:"+botUsername,
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

	streamerUsername, streamerPassword, _, err := s.Store().TwitchStreamerCredentials(s.userID)
	if err != nil {
		log.Printf("unable to get streamer creds: %s", err)
		return
	}
	botUsername, botPassword, _, err := s.Store().TwitchBotCredentials(s.userID)
	if err != nil {
		log.Printf("unable to get bot creds: %s", err)
		return
	}

	var username, password string
	switch userType {
	case "streamer":
		username, password = streamerUsername, streamerPassword
	case "bot":
		username, password = botUsername, botPassword
	default:
		resp.Error = invalidTwitchUserType
		return
	}
	s.api.streamManager.ConnectTwitch(username, "oauth:"+password, "#"+streamerUsername)
	s.api.streamManager.Send(stream.TXMessage{
		Type: stream.Twitch,
		Twitch: &stream.TXTwitch{
			Username: username,
			To:       "#" + streamerUsername,
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

	user, pass, _, err := s.Store().TwitchStreamerCredentials(s.userID)
	if err != nil {
		log.Printf("unable to get streamer creds: %s", err)
		return
	}

	err = s.api.twitchClient.UpdateDescription(status, game, user, pass)
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
		authenticated, err := s.Store().TwitchAuthenticated(userID)
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
		if authenticated {
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
