package general

import (
	"sort"

	"github.com/jasonkeene/anubot-server/api/internal/handlers"
)

// PingHandler responds with a pong event for testing connectivity.
func PingHandler(e handlers.Event, s handlers.Session) {
	resp, send := handlers.Setup(e, s)
	defer send()

	resp.Cmd = "pong"
	resp.Error = nil
}

// MethodsHandler responds with a list of commands the API supports.
type MethodsHandler struct {
	handlers map[string]handlers.EventHandler
}

// NewMethodsHandler returns a new MethodsHandler.
func NewMethodsHandler(handlers map[string]handlers.EventHandler) *MethodsHandler {
	return &MethodsHandler{
		handlers: handlers,
	}
}

// HandleEvent responds with a list of commands the API supports.
func (h *MethodsHandler) HandleEvent(e handlers.Event, s handlers.Session) {
	resp, send := handlers.Setup(e, s)
	defer send()

	methods := make([]string, 0, len(h.handlers))
	for m := range h.handlers {
		methods = append(methods, m)
	}
	sort.Slice(methods, func(i, j int) bool {
		return methods[i] < methods[j]
	})

	resp.Payload = methods
	resp.Error = nil
}
