package twitch

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jasonkeene/anubot-server/api/internal/handlers"
	"github.com/jasonkeene/anubot-server/stream"

	"github.com/pebbe/zmq4"
)

// Message is the structure that is written out to the websocket connection.
type Message struct {
	Type    stream.Type     `json:"type"`
	Twitch  *TMessage       `json:"twitch"`
	Discord *DiscordMessage `json:"discord"`
}

// TMessage represents a Twitch message.
type TMessage struct {
	Cmd    string            `json:"cmd"`
	Nick   string            `json:"nick"`
	Target string            `json:"target"`
	Body   string            `json:"body"`
	Time   time.Time         `json:"time"`
	Tags   map[string]string `json:"tags"`
}

// DiscordMessage represents a Discord message.
type DiscordMessage struct{}

type messageWriter struct {
	streamerUsername string
	streamerSub      *zmq4.Socket
	botSub           *zmq4.Socket
	s                handlers.Session
	requestID        string
}

func newMessageWriter(
	streamerUsername string,
	streamerTopic string,
	botTopic string,
	subEndpoints []string,
	s handlers.Session,
	requestID string,
) (*messageWriter, error) {
	streamerSub, err := zmq4.NewSocket(zmq4.SUB)
	if err != nil {
		return nil, err
	}
	err = streamerSub.SetSubscribe(streamerTopic)
	if err != nil {
		return nil, err
	}

	botSub, err := zmq4.NewSocket(zmq4.SUB)
	if err != nil {
		return nil, err
	}
	err = botSub.SetSubscribe(botTopic)
	if err != nil {
		return nil, err
	}

	for _, endpoint := range subEndpoints {
		err = streamerSub.Connect(endpoint)
		if err != nil {
			return nil, err
		}
		err = botSub.Connect(endpoint)
		if err != nil {
			return nil, err
		}
	}

	return &messageWriter{
		streamerUsername: streamerUsername,
		streamerSub:      streamerSub,
		botSub:           botSub,
		s:                s,
		requestID:        requestID,
	}, nil
}

// StartStreamer reads messages off of the streamer sub socket and writes them
// to the session's ws conn.
func (mw *messageWriter) StartStreamer() {
	for {
		ms, err := readMessage(mw.streamerSub)
		if err != nil {
			log.Printf("got err reading from streamer socket: %s", err)
			continue
		}
		err = mw.WriteMessage(ms)
		if err != nil {
			log.Printf("got error when writing to ws conn, aborting: %s", err)
			return
		}
	}
}

// StartBot reads messages off of the bot sub socket and writes them to the
// session's ws conn.
func (mw *messageWriter) StartBot() {
	for {
		ms, err := readMessage(mw.botSub)
		if err != nil {
			log.Printf("got err reading from streamer socket: %s", err)
			continue
		}
		if !UserMessage(ms, mw.streamerUsername) {
			continue
		}
		err = mw.WriteMessage(ms)
		if err != nil {
			log.Printf("got error when writing to ws conn, aborting: %s", err)
			return
		}
	}
}

func readMessage(sub *zmq4.Socket) (*stream.RXMessage, error) {
	rb, err := sub.RecvMessageBytes(0)
	if err != nil {
		return nil, fmt.Errorf("messages not read, got err: %s", err)
	}
	if len(rb) < 2 {
		return nil, fmt.Errorf("received message bytes had invalid length: %#v", rb)
	}

	var ms stream.RXMessage
	err = json.Unmarshal(rb[1], &ms)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal, got err: %s", err)
	}
	return &ms, nil
}

// UserMessage returns true if the message was sent from the user, otherwise it
// returns false.
func UserMessage(ms *stream.RXMessage, username string) bool {
	if ms.Type != stream.Twitch {
		return false
	}
	return ms.Twitch.Line.Nick == username
}

func (mw *messageWriter) WriteMessage(ms *stream.RXMessage) error {
	p := Message{
		Type: ms.Type,
	}
	switch ms.Type {
	case stream.Twitch:
		p.Twitch = &TMessage{
			Cmd:    ms.Twitch.Line.Cmd,
			Nick:   ms.Twitch.Line.Nick,
			Target: ms.Twitch.Line.Args[0],
			Body:   ms.Twitch.Line.Args[1],
			Time:   ms.Twitch.Line.Time,
			Tags:   ms.Twitch.Line.Tags,
		}
	case stream.Discord:
		// TODO: add support for discord messages
		return nil
	default:
		log.Println("got unknown message type while reading from sub sock")
		return nil
	}
	e := handlers.Event{
		Cmd:       "chat-message",
		RequestID: mw.requestID,
		Payload:   p,
	}
	return mw.s.Send(e)
}
