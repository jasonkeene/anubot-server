package bttv_test

import (
	"github.com/jasonkeene/anubot-server/api/internal/handlers"
	"github.com/jasonkeene/anubot-server/store"
)

type SpySession struct {
	handlers.Session
	sendCalledWith handlers.Event
	userID         string
	authenticated  bool
}

func (s *SpySession) Send(e handlers.Event) error {
	s.sendCalledWith = e
	return nil
}

func (s *SpySession) Authenticated() (userID string, authenticated bool) {
	return s.userID, s.authenticated
}

type SpyTwitchCredentialsProvider struct {
	calledWith string
	creds      store.TwitchCredentials
	err        error
}

func (s *SpyTwitchCredentialsProvider) TwitchCredentials(userID string) (creds store.TwitchCredentials, err error) {
	s.calledWith = userID
	return s.creds, s.err
}

type SpyEmojiProvider struct {
	calledWith string
	emoji      map[string]string
	err        error
}

func (s *SpyEmojiProvider) Emoji(channel string) (emoji map[string]string, err error) {
	s.calledWith = channel
	return s.emoji, s.err
}
