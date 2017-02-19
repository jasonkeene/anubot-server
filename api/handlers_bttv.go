package api

import (
	"log"

	"github.com/jasonkeene/anubot-server/bttv"
)

// bttvEmojiHandler returns emoji from BTTV. If the user has authenticated
// their streamer user with Twitch it will also include channel specific
// emoji.
func bttvEmojiHandler(e event, s *session) {
	streamerAuthenticated, err := s.Store().TwitchStreamerAuthenticated(s.userID)
	if err != nil {
		log.Printf("error authenticating twitch streamer: %s", err)
		err = s.Send(event{
			Cmd:       e.Cmd,
			RequestID: e.RequestID,
			Error:     unknownError,
		})
		if err != nil {
			log.Printf("unable to tx: %s", err)
		}
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

	payload, err := bttv.Emoji(streamerUsername)
	if err != nil {
		log.Printf("error getting bttv emoji: %s", err)
		err = s.Send(event{
			Cmd:       e.Cmd,
			RequestID: e.RequestID,
			Error:     bttvUnavailable,
		})
		if err != nil {
			log.Printf("unable to tx: %s", err)
		}
		return
	}

	err = s.Send(event{
		Cmd:       e.Cmd,
		RequestID: e.RequestID,
		Payload:   payload,
	})
	if err != nil {
		log.Printf("unable to tx: %s", err)
	}
}
