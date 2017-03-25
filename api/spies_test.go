package api_test

import (
	"github.com/fluffle/goirc/client"
	"github.com/jasonkeene/anubot-server/api"
	"github.com/jasonkeene/anubot-server/store"
	"github.com/jasonkeene/anubot-server/stream"
	"github.com/jasonkeene/anubot-server/twitch"
)

type SpyStreamManager struct {
	api.StreamManager
}

func (s *SpyStreamManager) ConnectTwitch(user, pass, channel string) {}

type SpyStore struct {
	api.Store

	userID        string
	authenticated bool
	err           error

	creds store.TwitchCredentials
}

func (s *SpyStore) AuthenticateUser(username, password string) (userID string, authenticated bool, err error) {
	return s.userID, s.authenticated, s.err
}

func (s *SpyStore) TwitchCredentials(userID string) (creds store.TwitchCredentials, err error) {
	return s.creds, nil
}

func (s *SpyStore) TwitchClearAuth(userID string) (err error) {
	return nil
}

func (s *SpyStore) FetchRecentMessages(userID string) (msgs []stream.RXMessage, err error) {
	return []stream.RXMessage{{
		Type: stream.Twitch,
		Twitch: &stream.RXTwitch{
			Line: &client.Line{
				Args: []string{"", ""},
			},
		},
	}}, nil
}

type SpyTwitchClient struct {
	api.TwitchClient
}

func (s *SpyTwitchClient) Games() (games []twitch.Game) {
	return nil
}
