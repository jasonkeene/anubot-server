package api

import (
	"sort"
)

// pingHandler responds with pong data frame used for testing connectivity.
func pingHandler(e event, s *session) {
	resp, send := setup(e, s)
	defer send()

	resp.Cmd = "pong"
	resp.Error = nil
}

// methodsHandler responds with a list of methods the API supports.
func methodsHandler(e event, s *session) {
	resp, send := setup(e, s)
	defer send()

	methods := make([]string, 0, len(eventHandlers))
	for m := range eventHandlers {
		methods = append(methods, m)
	}
	sort.Slice(methods, func(i, j int) bool {
		return methods[i] < methods[j]
	})

	resp.Payload = methods
	resp.Error = nil
}
