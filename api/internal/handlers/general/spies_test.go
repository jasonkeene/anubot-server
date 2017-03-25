package general_test

import "github.com/jasonkeene/anubot-server/api/internal/handlers"

type SpySession struct {
	handlers.Session
	sendCalledWith handlers.Event
}

func (s *SpySession) Send(e handlers.Event) error {
	s.sendCalledWith = e
	return nil
}
