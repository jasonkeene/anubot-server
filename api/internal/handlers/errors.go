package handlers

// Error represents an error that occurred with the API.
type Error struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

// Error returns the textual representation of the error.
func (e Error) Error() string {
	return e.Text
}

var (
	// UnknownError is the default error message.
	UnknownError = &Error{
		Code: 0,
		Text: "unknown error",
	}
	// InvalidCommand occurs when an event is received that has a command that
	// is not recognized by the API.
	InvalidCommand = &Error{
		Code: 1,
		Text: "invalid command",
	}
	// InvalidPayload occurs when an event is received and the payload is not
	// in the correct format.
	InvalidPayload = &Error{
		Code: 2,
		Text: "invalid payload data",
	}
	// UsernameTaken is only ever returned when attempting to register a user
	// and the username is already taken by another user.
	UsernameTaken = &Error{
		Code: 3,
		Text: "username has already been taken",
	}
	// AuthenticationError occurs when the user typed the wrong username or
	// password or the user is not authenticated and is trying to access an
	// endpoint that requires authentication.
	AuthenticationError = &Error{
		Code: 4,
		Text: "authentication error",
	}
	// TwitchAuthenticationError occurs when the user has not authenticated
	// with Twitch but is trying to access an endpoint that requires Twitch
	// authentication.
	TwitchAuthenticationError = &Error{
		Code: 5,
		Text: "authentication error with twitch",
	}
	// TwitchOauthStartOrderError occurs when the user requests to start the
	// oauth flow for the bot before their streamer user has completed the
	// flow.
	TwitchOauthStartOrderError = &Error{
		Code: 6,
		Text: "unable to start oauth flow for bot, streamer not finished",
	}
	// BTTVUnavailable occurs when BTTV's API is down.
	BTTVUnavailable = &Error{
		Code: 7,
		Text: "unable to gather emoji from bttv api",
	}
)
