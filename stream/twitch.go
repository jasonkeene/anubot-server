package stream

import (
	"crypto/tls"
	"errors"
	"log"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/fluffle/goirc/client"
)

const (
	defaultTwitchHost = "irc.chat.twitch.tv"
	defaultTwitchPort = 443
)

var (
	twitchHost         = defaultTwitchHost
	twitchPort         = defaultTwitchPort
	insecureSkipVerify = false
	flood              = false

	capString string
)

func init() {
	caps := []string{
		"twitch.tv/tags",
		"twitch.tv/commands",
		"twitch.tv/membership",
	}
	capString = strings.Join(caps, " ")
}

type twitchConn struct {
	d  chan dispatchMessage
	c  *client.Conn
	u  string
	id int
}

func (c *twitchConn) send(m TXMessage) {
	c.c.Privmsg(m.Twitch.To, m.Twitch.Message)
}

func (c *twitchConn) close() error {
	disconnected := make(chan struct{})
	c.c.HandleFunc("DISCONNECTED", func(conn *client.Conn, line *client.Line) {
		close(disconnected)
	})
	log.Printf("twitchConn.close: disconnecting from twitch for user: %s", c.u)
	c.c.Quit()
	log.Printf("twitchConn.close: waiting for disconnect event from twitch for user: %s", c.u)
	<-disconnected
	log.Printf("twitchConn.close: disconnected from twitch for user: %s", c.u)
	return nil
}

func connectTwitch(u, p, c string, d chan dispatchMessage, twitch TwitchUserIDFetcher) (*twitchConn, error) {
	userID, err := twitch.UserID(u)
	if err != nil {
		return nil, err
	}

	cfg := client.NewConfig(u)
	cfg.Me.Name = u
	cfg.Me.Ident = "anubot"
	cfg.Pass = p
	cfg.Flood = flood
	cfg.SSL = true
	cfg.SSLConfig = &tls.Config{
		ServerName:         twitchHost,
		InsecureSkipVerify: insecureSkipVerify,
	}
	cfg.Server = net.JoinHostPort(twitchHost, strconv.Itoa(twitchPort))
	tc := &twitchConn{
		d:  d,
		c:  client.Client(cfg),
		u:  u,
		id: userID,
	}

	connected := make(chan struct{})
	tc.c.HandleFunc("CONNECTED", func(conn *client.Conn, line *client.Line) {
		close(connected)
	})
	tc.c.HandleFunc("PRIVMSG", tc.dispatchMessage)
	tc.c.HandleFunc("ACTION", tc.dispatchMessage)
	tc.c.HandleFunc("WHISPER", tc.dispatchMessage)

	log.Printf("connectTwitch: connecting to twitch for user: %s", u)
	if err := tc.c.Connect(); err != nil {
		log.Printf("connectTwitch: connection to twitch failed for user: %s: %s", u, err)
		return nil, err
	}
	log.Printf("connectTwitch: connection to twitch established for user: %s", u)
	select {
	case <-connected:
	case <-time.After(3 * time.Second):
		log.Printf("connectTwitch: did not receive CONNECTED event")
		return nil, errors.New("did not receive CONNECTED event")
	}
	log.Printf("connectTwitch: received connection event from twitch for user: %s", u)
	tc.c.Join(c)
	log.Printf("connectTwitch: joined channel: %s on twitch for user: %s", c, u)

	tc.c.Raw("CAP REQ :" + capString)
	log.Printf("connectTwitch: requested capabilities on twitch for user: %s", u)

	return tc, nil
}

func (c *twitchConn) dispatchMessage(conn *client.Conn, line *client.Line) {
	topic := "twitch:" + c.u
	msg := RXMessage{
		Type: Twitch,
		Twitch: &RXTwitch{
			OwnerID: c.id,
			Line:    line,
		},
	}
	select {
	case c.d <- dispatchMessage{
		topic: topic,
		msg:   msg,
	}:
	default:
		log.Println("twitchConn.dispatchMessage: unable to dispatch message")
	}
}
