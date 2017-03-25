package general_test

import (
	"testing"

	"github.com/a8m/expect"
	"github.com/jasonkeene/anubot-server/api/internal/handlers"
	"github.com/jasonkeene/anubot-server/api/internal/handlers/general"
)

func TestPingHandler(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{}
	event := handlers.Event{
		Cmd:       "ping",
		RequestID: "test-request-id",
	}
	general.PingHandler(event, spySession)

	expected := handlers.Event{
		Cmd:       "pong",
		RequestID: "test-request-id",
	}
	expect(spySession.sendCalledWith).To.Equal(expected)
}

func TestMethodsHandler(t *testing.T) {
	expect := expect.New(t)

	spySession := &SpySession{}
	methods := map[string]handlers.EventHandler{
		"test-handler-2": nil,
		"test-handler-1": nil,
	}
	handler := general.NewMethodsHandler(methods)
	event := handlers.Event{
		Cmd:       "methods",
		RequestID: "test-request-id",
	}
	handler.HandleEvent(event, spySession)

	expected := handlers.Event{
		Cmd:       "methods",
		RequestID: "test-request-id",
		Payload: []string{
			"test-handler-1",
			"test-handler-2",
		},
	}
	expect(spySession.sendCalledWith).To.Equal(expected)
}
