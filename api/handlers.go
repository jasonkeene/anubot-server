package api

import "log"

// eventHandlers is a map of event commands to their handlers.
var eventHandlers map[string]eventHandler

// eventHandler is a func that can handle events from a websocket connection.
type eventHandler func(event, *session)

// setup creates a generic response and a func that can be used to send that
// response.
func setup(e event, s *session) (*event, func()) {
	resp := &event{
		Cmd:       e.Cmd,
		RequestID: e.RequestID,
		Error:     unknownError,
	}
	return resp, func() {
		err := s.Send(*resp)
		if err != nil {
			log.Printf("unable to tx: %s", err)
		}
	}
}

func init() {
	eventHandlers = make(map[string]eventHandler)

	// public
	{
		// general
		eventHandlers["ping"] = pingHandler
		eventHandlers["methods"] = methodsHandler

		// authentication
		eventHandlers["register"] = registerHandler
		eventHandlers["authenticate"] = authenticateHandler
		eventHandlers["logout"] = logoutHandler
	}

	// authenticated
	{
		// twitch oauth
		eventHandlers["twitch-oauth-start"] = authenticateWrapper(
			twitchOauthStartHandler,
		)
		eventHandlers["twitch-clear-auth"] = authenticateWrapper(
			twitchClearAuthHandler,
		)

		// user information
		eventHandlers["twitch-user-details"] = authenticateWrapper(
			twitchUserDetailsHandler,
		)

		// twitch
		eventHandlers["twitch-games"] = authenticateWrapper(
			twitchGamesHandler,
		)

		// bttv
		eventHandlers["bttv-emoji"] = authenticateWrapper(
			bttvEmojiHandler,
		)
	}

	// twitch authenticated
	{
		// twitch chat
		eventHandlers["twitch-stream-messages"] = authenticateWrapper(
			twitchAuthenticateWrapper(
				twitchStreamMessagesHandler,
			),
		)
		eventHandlers["twitch-send-message"] = authenticateWrapper(
			twitchAuthenticateWrapper(
				twitchSendMessageHandler,
			),
		)
		eventHandlers["twitch-update-chat-description"] = authenticateWrapper(
			twitchAuthenticateWrapper(
				twitchUpdateChatDescriptionHandler,
			),
		)
	}
}
