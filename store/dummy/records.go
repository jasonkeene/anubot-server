package dummy

import (
	"errors"
	"strconv"
	"time"

	"github.com/jasonkeene/anubot-server/store"
	"github.com/jasonkeene/anubot-server/stream"
)

type nonceRecord struct {
	userID  string
	tu      store.TwitchUser
	created time.Time
}

type users map[string]userRecord

func (u users) lookup(username string) (string, userRecord, bool) {
	for id, creds := range u {
		if creds.username == username {
			return id, creds, true
		}
	}
	return "", userRecord{}, false
}

func (u users) exists(username string) bool {
	_, _, exists := u.lookup(username)
	return exists
}

type userRecord struct {
	username         string
	password         string
	streamerUsername string
	streamerOD       store.OauthData
	streamerID       int
	botUsername      string
	botOD            store.OauthData
	botID            int
}

func messageKey(msg stream.RXMessage) (string, error) {
	switch msg.Type {
	case stream.Twitch:
		return "twitch:" + strconv.Itoa(msg.Twitch.OwnerID), nil
	case stream.Discord:
		return "discord:" + msg.Discord.OwnerID, nil
	default:
		return "", errors.New("invalid message type")
	}
}
