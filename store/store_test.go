package store_test

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/a8m/expect"
	"github.com/fluffle/goirc/client"
	"github.com/jasonkeene/anubot-server/store"
	"github.com/jasonkeene/anubot-server/stream"
)

func TestThatBackendsComplyWithAllStoreMethods(t *testing.T) {
	var (
		_ store.Store = &store.Postgres{}
		_ store.Store = &store.Bolt{}
		_ store.Store = &store.Dummy{}
	)
}

func TestThatRegisteringAUserReservesThatUsername(t *testing.T) {
	expect := expect.New(t)

	backends, cleanup := setupBackends(t)
	defer cleanup()

	for _, b := range backends {
		_, err := b.RegisterUser("test-user", "test-pass")
		expect(err).To.Be.Nil()
		_, err = b.RegisterUser("test-user", "test-pass")
		expect(err).To.Equal(store.ErrUsernameTaken)
	}
}

func TestThatUsersCanAuthenticate(t *testing.T) {
	expect := expect.New(t)

	backends, cleanup := setupBackends(t)
	defer cleanup()

	for _, b := range backends {
		expectedUserID, err := b.RegisterUser("test-user", "test-pass")
		expect(err).To.Be.Nil()

		userID, authenticated, err := b.AuthenticateUser("test-user", "bad-pass")
		expect(err).To.Be.Nil()
		expect(userID).To.Equal("")
		expect(authenticated).To.Be.False()

		userID, authenticated, err = b.AuthenticateUser("test-user", "test-pass")
		expect(err).To.Be.Nil()
		expect(userID).To.Equal(expectedUserID)
		expect(authenticated).To.Be.True()
	}
}

func TestThatTwitchOauthFlowWorks(t *testing.T) {
	expect := expect.New(t)

	backends, cleanup := setupBackends(t)
	defer cleanup()

	for _, b := range backends {
		userID, err := b.RegisterUser("test-user", "test-pass")
		expect(err).To.Be.Nil()

		_, err = b.OauthNonce(userID, store.Streamer)
		expect(err).Not.To.Be.Nil()

		streamerNonce := "streamer-nonce"
		err = b.StoreOauthNonce(userID, store.Streamer, streamerNonce)
		expect(err).To.Be.Nil()

		actualStreamerNonce, err := b.OauthNonce(userID, store.Streamer)
		expect(actualStreamerNonce).To.Equal(streamerNonce)
		expect(err).To.Be.Nil()

		ok, err := b.OauthNonceExists(streamerNonce)
		expect(err).To.Be.Nil()
		expect(ok).To.Be.Ok()

		streamerOD := store.OauthData{
			AccessToken:  "test-streamer-access-token",
			RefreshToken: "test-streamer-refresh-token",
			Scope:        []string{"test-streamer-scope"},
		}
		err = b.FinishOauthNonce(streamerNonce, "test-streamer-user", 12345, streamerOD)
		expect(err).To.Be.Nil()

		ok, _ = b.OauthNonceExists(streamerNonce)
		expect(ok).Not.To.Be.Ok()

		_, err = b.OauthNonce(userID, store.Bot)
		expect(err).Not.To.Be.Nil()

		botNonce := "bot-nonce"
		err = b.StoreOauthNonce(userID, store.Bot, botNonce)
		expect(err).To.Be.Nil()

		actualBotNonce, err := b.OauthNonce(userID, store.Bot)
		expect(actualBotNonce).To.Equal(botNonce)
		expect(err).To.Be.Nil()

		ok, err = b.OauthNonceExists(botNonce)
		expect(err).To.Be.Nil()
		expect(ok).To.Be.Ok()

		botOD := store.OauthData{
			AccessToken:  "test-bot-access-token",
			RefreshToken: "test-bot-refresh-token",
			Scope:        []string{"test-bot-scope"},
		}
		err = b.FinishOauthNonce(botNonce, "test-bot-user", 54321, botOD)
		expect(err).To.Be.Nil()

		ok, _ = b.OauthNonceExists(botNonce)
		expect(ok).Not.To.Be.Ok()

		creds, err := b.TwitchCredentials(userID)
		expect(err).To.Be.Nil()
		expectedCreds := store.TwitchCredentials{
			StreamerAuthenticated: true,
			StreamerUsername:      "test-streamer-user",
			StreamerPassword:      "test-streamer-access-token",
			StreamerTwitchUserID:  12345,
			BotAuthenticated:      true,
			BotUsername:           "test-bot-user",
			BotPassword:           "test-bot-access-token",
			BotTwitchUserID:       54321,
		}
		expect(creds).To.Equal(expectedCreds)

		err = b.TwitchClearAuth(userID)
		expect(err).To.Be.Nil()

		creds, err = b.TwitchCredentials(userID)
		expectedCreds = store.TwitchCredentials{
			StreamerAuthenticated: false,
			StreamerUsername:      "",
			StreamerPassword:      "",
			StreamerTwitchUserID:  0,
			BotAuthenticated:      false,
			BotUsername:           "",
			BotPassword:           "",
			BotTwitchUserID:       0,
		}
		expect(err).To.Be.Nil()
		expect(creds).To.Equal(expectedCreds)
	}
}

func TestThatYouCanStoreMessages(t *testing.T) {
	expect := expect.New(t)

	backends, cleanup := setupBackends(t)
	defer cleanup()

	for _, b := range backends {
		userID, err := b.RegisterUser("test-user", "test-pass")
		expect(err).To.Be.Nil()

		od := store.OauthData{
			AccessToken:  "test-access-token",
			RefreshToken: "test-refresh-token",
			Scope:        []string{"test-scope"},
		}
		streamerNonce := "streamer-nonce"
		err = b.StoreOauthNonce(userID, store.Streamer, streamerNonce)
		expect(err).To.Be.Nil()
		err = b.FinishOauthNonce(streamerNonce, "test-streamer-user", 12345, od)
		expect(err).To.Be.Nil()

		botNonce := "bot-nonce"
		err = b.StoreOauthNonce(userID, store.Bot, botNonce)
		expect(err).To.Be.Nil()
		err = b.FinishOauthNonce(botNonce, "test-bot-user", 54321, od)
		expect(err).To.Be.Nil()

		msg1 := stream.RXMessage{
			Type: stream.Twitch,
			Twitch: &stream.RXTwitch{
				OwnerID: 12345,
				Line: &client.Line{
					Raw: "test-message-1",
				},
			},
		}
		err = b.StoreMessage(msg1)
		expect(err).To.Be.Nil().Else.FailNow()
		msg2 := stream.RXMessage{
			Type: stream.Twitch,
			Twitch: &stream.RXTwitch{
				OwnerID: 12345,
				Line: &client.Line{
					Raw: "test-message-2",
				},
			},
		}
		err = b.StoreMessage(msg2)
		expect(err).To.Be.Nil().Else.FailNow()

		messages, err := b.FetchRecentMessages(userID)
		expect(err).To.Be.Nil().Else.FailNow()
		expect(len(messages)).To.Equal(2).Else.FailNow()
		expect(messages[0].Twitch.Line.Raw).To.Equal("test-message-1")
		expect(messages[1].Twitch.Line.Raw).To.Equal("test-message-2")
	}
}

func setupBackends(t *testing.T) ([]store.Store, func()) {
	bolt, cleanup := setupBolt(t)
	stores := []store.Store{
		bolt,
		store.NewDummy(),
	}
	cleanups := []func(){
		cleanup,
	}
	if os.Getenv("ANUBOT_TEST_POSTGRES") != "" {
		pg, cleanup := setupPostgres(t)
		stores = append(stores, pg)
		cleanups = append(cleanups, cleanup)
	}
	return stores, func() {
		for _, cleanup := range cleanups {
			cleanup()
		}
	}
}

func setupBolt(t *testing.T) (*store.Bolt, func()) {
	path, cleanup := tempFile(t)
	b, err := store.NewBolt(path)
	if err != nil {
		t.Fatal(err)
	}
	return b, func() {
		err := b.Close()
		if err != nil {
			log.Println("unable to close boltdb")
			t.FailNow()
		}
		cleanup()
	}
}

func setupPostgres(t *testing.T) (*store.Postgres, func()) {
	pg, err := store.NewPostgres(
		os.Getenv("ANUBOT_TEST_POSTGRES"),
		randomKey(),
	)
	if err != nil {
		t.Fatal(err)
	}
	err = pg.Ping()
	if err != nil {
		t.Fatal(err)
	}
	return pg, func() {
		err := truncatePostgres()
		if err != nil {
			log.Printf("unable to truncate postgres: %s", err)
			t.Fail()
		}
		err = pg.Close()
		if err != nil {
			log.Printf("unable to close postgres: %s", err)
			t.FailNow()
		}
	}
}

func truncatePostgres() error {
	db, err := sql.Open("postgres", os.Getenv("ANUBOT_TEST_POSTGRES"))
	if err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	tables := []string{
		"message",
		"nonce",
		"user",
	}
	for _, table := range tables {
		_, err := tx.Exec(`TRUNCATE TABLE "` + table + `" CASCADE`)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func tempFile(t *testing.T) (string, func()) {
	tf, err := ioutil.TempFile("", "")
	if err != nil {
		fmt.Println("could not obtain a temporary file")
		t.FailNow()
	}
	return tf.Name(), func() {
		err := os.Remove(tf.Name())
		if err != nil {
			log.Println("unable to remove temp file")
			t.FailNow()
		}
	}
}
