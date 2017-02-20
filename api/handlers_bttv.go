package api

import "log"

// bttvEmojiHandler returns emoji from BTTV. If the user has authenticated
// their streamer user with Twitch it will also include channel specific
// emoji.
func bttvEmojiHandler(e event, s *session) {
	resp, send := setup(e, s)
	defer send()

	streamerAuthenticated, err := s.Store().TwitchStreamerAuthenticated(s.userID)
	if err != nil {
		log.Printf("error authenticating twitch streamer: %s", err)
		return
	}

	var streamerUsername string
	if streamerAuthenticated {
		streamerUsername, _, _, err = s.Store().TwitchStreamerCredentials(s.userID)
		if err != nil {
			log.Printf("error getting twitch streamer creds: %s", err)
			streamerUsername = ""
		}
	}

	payload, err := s.api.bttvClient.Emoji(streamerUsername)
	if err != nil {
		log.Printf("error getting bttv emoji: %s", err)
		resp.Error = bttvUnavailable
		return
	}

	resp.Payload = payload
	resp.Error = nil
}
