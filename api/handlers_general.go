package api

import "log"

// pingHandler responds with pong data frame used for testing connectivity.
func pingHandler(e event, s *session) {
	err := s.Send(event{
		Cmd:       "pong",
		RequestID: e.RequestID,
	})
	if err != nil {
		log.Printf("unable to tx: %s", err)
	}
}

// methodsHandler responds with a list of methods the API supports.
func methodsHandler(e event, s *session) {
	methods := []string{}
	for m := range eventHandlers {
		methods = append(methods, m)
	}
	err := s.Send(event{
		Cmd:       "methods",
		RequestID: e.RequestID,
		Payload:   methods,
	})
	if err != nil {
		log.Printf("unable to tx: %s", err)
	}
}
