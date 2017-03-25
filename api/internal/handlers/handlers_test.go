package handlers_test

import (
	"testing"

	"github.com/a8m/expect"
	"github.com/jasonkeene/anubot-server/api/internal/handlers"
)

func TestSetup(t *testing.T) {
	expect := expect.New(t)

	originalEvent := handlers.Event{
		Cmd:       "test-command",
		RequestID: "test-request-id",
	}
	spySession := &SpySession{}

	event, send := handlers.Setup(originalEvent, spySession)

	expect(event.Cmd).To.Equal(originalEvent.Cmd)
	expect(event.RequestID).To.Equal(originalEvent.RequestID)
	expect(event.Error).To.Equal(handlers.UnknownError)

	send()
	expect(spySession.sendCalledWith).To.Equal(*event)
}

type SpySession struct {
	handlers.Session
	sendCalledWith handlers.Event
}

func (s *SpySession) Send(e handlers.Event) (err error) {
	s.sendCalledWith = e
	return nil
}
