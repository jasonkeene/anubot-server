package store

import (
	"database/sql"
	"encoding/json"
	"errors"

	"golang.org/x/crypto/chacha20poly1305"

	// Import pq driver for registration side effects.
	_ "github.com/lib/pq"
	uuid "github.com/satori/go.uuid"

	"github.com/jasonkeene/anubot-server/stream"
)

// Postgres is a store backend for the postgres database.
type Postgres struct {
	db  *sql.DB
	key []byte
}

// NewPostgres creates a new postgres store.
func NewPostgres(url string, key []byte) (*Postgres, error) {
	_, err := chacha20poly1305.New(key)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	return &Postgres{
		db:  db,
		key: key,
	}, nil
}

// Ping verifies a connection to the database is still alive,
// establishing a connection if necessary.
func (p *Postgres) Ping() (err error) {
	return p.db.Ping()
}

// Close closes the underlying sql.DB.
func (p *Postgres) Close() (err error) {
	return p.db.Close()
}

// RegisterUser registers a new user returning the user ID.
func (p *Postgres) RegisterUser(username, password string) (userID string, err error) {
	hash, err := Hash(password)
	if err != nil {
		return "", err
	}

	tx, err := p.db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`SELECT COUNT(*) AS n FROM "user" WHERE username=$1`)
	if err != nil {
		return "", err
	}
	defer stmt.Close()
	var n int
	err = stmt.QueryRow(username).Scan(&n)
	if err != nil {
		if err != sql.ErrNoRows {
			return "", err
		}
	}
	if n != 0 {
		return "", ErrUsernameTaken
	}

	id := uuid.NewV4().String()
	istmt, err := tx.Prepare(`INSERT INTO "user" (user_id, username, password_hash) VALUES ($1, $2, $3)`)
	if err != nil {
		return "", err
	}
	defer istmt.Close()
	_, err = istmt.Exec(id, username, hash)
	if err != nil {
		return "", err
	}

	err = tx.Commit()
	if err != nil {
		return "", err
	}
	return id, nil
}

// AuthenticateUser checks to see if the given user credentials are valid. If
// they are the user ID is returned with a bool to indicate success.
func (p *Postgres) AuthenticateUser(username, password string) (userID string, success bool, err error) {
	tx, err := p.db.Begin()
	if err != nil {
		return "", false, err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`SELECT user_id, password_hash FROM "user" WHERE username=$1`)
	if err != nil {
		return "", false, err
	}
	defer stmt.Close()

	var passwordHash string
	err = stmt.QueryRow(username).Scan(&userID, &passwordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", false, nil
		}
		return "", false, err
	}

	success, err = Verify(password, passwordHash)
	if err != nil || !success {
		return "", false, err
	}
	return userID, true, nil
}

// StoreOauthNonce stores the oauth nonce.
func (p *Postgres) StoreOauthNonce(userID string, tu TwitchUser, nonce string) (err error) {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	istmt, err := tx.Prepare(`INSERT INTO nonce (user_id, twitch_user, nonce) VALUES ($1, $2, $3)`)
	if err != nil {
		return err
	}
	defer istmt.Close()

	_, err = istmt.Exec(userID, tu.String(), nonce)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// OauthNonceExists tells you if the provided nonce was recently created and
// not yet finished.
func (p *Postgres) OauthNonceExists(nonce string) (exists bool, err error) {
	tx, err := p.db.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`SELECT COUNT(*) AS n FROM nonce WHERE nonce=$1`)
	if err != nil {
		return false, err
	}
	defer stmt.Close()

	var n int
	err = stmt.QueryRow(nonce).Scan(&n)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	if n != 1 {
		return false, errors.New("invalid count of nonce")
	}
	return true, nil
}

// FinishOauthNonce completes the oauth flow, removing the nonce and storing
// the oauth data.
func (p *Postgres) FinishOauthNonce(nonce string, username string, twitchUserID int, od OauthData) (err error) {
	odJSON, err := json.Marshal(od)
	if err != nil {
		return err
	}
	box, err := Encrypt(odJSON, p.key)
	if err != nil {
		return err
	}

	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`SELECT user_id, twitch_user FROM nonce WHERE nonce=$1`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	var (
		userID     string
		twitchUser string
	)
	err = stmt.QueryRow(nonce).Scan(&userID, &twitchUser)
	if err != nil {
		return err
	}

	var ustmt *sql.Stmt
	switch twitchUser {
	case "Streamer":
		var err error
		ustmt, err = tx.Prepare(`UPDATE "user" SET streamer_id=$2, streamer_username=$3, streamer_oauth_data=$4 WHERE user_id=$1`)
		if err != nil {
			return err
		}
		defer ustmt.Close()
	case "Bot":
		var err error
		ustmt, err = tx.Prepare(`UPDATE "user" SET bot_id=$2, bot_username=$3, bot_oauth_data=$4 WHERE user_id=$1`)
		if err != nil {
			return err
		}
		defer ustmt.Close()
	}

	_, err = ustmt.Exec(userID, twitchUserID, username, box)
	if err != nil {
		return err
	}

	dstmt, err := tx.Prepare(`DELETE FROM nonce WHERE nonce=$1`)
	if err != nil {
		return err
	}
	defer dstmt.Close()
	_, err = dstmt.Exec(nonce)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// TwitchCredentials gives you the twitch credentials for a given users.
func (p *Postgres) TwitchCredentials(userID string) (creds TwitchCredentials, err error) {
	tx, err := p.db.Begin()
	if err != nil {
		return TwitchCredentials{}, err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`SELECT streamer_id, streamer_username, streamer_oauth_data, bot_id, bot_username, bot_oauth_data FROM "user" WHERE user_id=$1`)
	if err != nil {
		return TwitchCredentials{}, err
	}
	defer stmt.Close()

	var (
		streamerID        int
		streamerUsername  string
		streamerOauthData string
		botID             int
		botUsername       string
		botOauthData      string
	)
	err = stmt.QueryRow(userID).Scan(
		&streamerID,
		&streamerUsername,
		&streamerOauthData,
		&botID,
		&botUsername,
		&botOauthData,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return TwitchCredentials{}, nil
		}
		return TwitchCredentials{}, err
	}

	if len(streamerOauthData) == 0 {
		return TwitchCredentials{}, nil
	}
	streamerODPlain, err := Decrypt(streamerOauthData, p.key)
	if err != nil {
		return TwitchCredentials{}, err
	}
	var streamerOD OauthData
	err = json.Unmarshal(streamerODPlain, &streamerOD)
	if err != nil {
		return TwitchCredentials{}, err
	}

	creds.StreamerAuthenticated = streamerOD.AccessToken != ""
	creds.StreamerUsername = streamerUsername
	creds.StreamerPassword = streamerOD.AccessToken
	creds.StreamerTwitchUserID = streamerID

	if len(botOauthData) == 0 {
		return creds, nil
	}
	botODPlain, err := Decrypt(botOauthData, p.key)
	if err != nil {
		return TwitchCredentials{}, err
	}
	var botOD OauthData
	err = json.Unmarshal(botODPlain, &botOD)
	if err != nil {
		return TwitchCredentials{}, err
	}

	creds.BotAuthenticated = botOD.AccessToken != ""
	creds.BotUsername = botUsername
	creds.BotPassword = botOD.AccessToken
	creds.BotTwitchUserID = botID

	err = tx.Commit()
	if err != nil {
		return TwitchCredentials{}, err
	}
	return creds, nil
}

// TwitchClearAuth removes all the auth data for twitch for the user.
func (p *Postgres) TwitchClearAuth(userID string) (err error) {
	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`UPDATE "user" SET streamer_id=$2, streamer_username=$3, streamer_oauth_data=$4, bot_id=$5, bot_username=$6, bot_oauth_data=$7 WHERE user_id=$1`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID, 0, "", "", 0, "", "")
	if err != nil {
		return err
	}

	return tx.Commit()
}

// StoreMessage stores a message for a given user for later searching and
// scrollback history.
func (p *Postgres) StoreMessage(msg stream.RXMessage) (err error) {
	message, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	tx, err := p.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	switch msg.Type {
	case stream.Twitch:
		if msg.Twitch == nil {
			return errors.New("invalid twitch message")
		}
		stmt, err := tx.Prepare(`INSERT INTO message (source, message, twitch_owner_id) VALUES ($1, $2, $3)`)
		if err != nil {
			return err
		}
		defer stmt.Close()
		_, err = stmt.Exec(msg.Type.String(), message, msg.Twitch.OwnerID)
		if err != nil {
			return err
		}
	case stream.Discord:
		if msg.Discord == nil {
			return errors.New("invalid discord message")
		}
		stmt, err := tx.Prepare(`INSERT INTO message (source, message, discord_owner_id) VALUES ($1, $2, $3)`)
		if err != nil {
			return err
		}
		defer stmt.Close()
		_, err = stmt.Exec(msg.Type.String(), message, msg.Discord.OwnerID)
		if err != nil {
			return err
		}
	default:
		return errors.New("invalid message type")
	}

	return tx.Commit()
}

// FetchRecentMessages gets the recent messages for the user's channel.
func (p *Postgres) FetchRecentMessages(userID string) (msgs []stream.RXMessage, err error) {
	tx, err := p.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`SELECT streamer_id FROM "user" WHERE user_id=$1`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	var twitchStreamerID int
	err = stmt.QueryRow(userID).Scan(&twitchStreamerID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	mstmt, err := tx.Prepare(`SELECT message FROM message WHERE source='Twitch' AND twitch_owner_id=$1 LIMIT 500`)
	if err != nil {
		return nil, err
	}
	defer mstmt.Close()
	rows, err := mstmt.Query(twitchStreamerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []stream.RXMessage
	for rows.Next() {
		var messageBytes []byte
		err := rows.Scan(&messageBytes)
		if err != nil {
			return nil, err
		}

		var message stream.RXMessage
		err = json.Unmarshal(messageBytes, &message)
		if err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return messages, nil
}

// QueryMessages allows the user to search for messages that match a
// search string.
func (p *Postgres) QueryMessages(userID string, search string) (msgs []stream.RXMessage, err error) {
	panic("not implemented")
}
