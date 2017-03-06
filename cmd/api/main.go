package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"

	"github.com/fluffle/goirc/logging/golog"
	"github.com/spf13/viper"

	"github.com/jasonkeene/anubot-server/api"
	"github.com/jasonkeene/anubot-server/dispatch"
	"github.com/jasonkeene/anubot-server/store"
	"github.com/jasonkeene/anubot-server/stream"
	"github.com/jasonkeene/anubot-server/twitch"
	"github.com/jasonkeene/anubot-server/twitch/oauth"
)

func init() {
	golog.Init()
}

func main() {
	// load config
	v := viper.New()
	v.SetEnvPrefix("anubot")
	v.AutomaticEnv()

	// setup twitch api client
	twitchClient := twitch.New(
		v.GetString("twitch_api_url"),
		v.GetString("twitch_oauth_client_id"),
	)

	// create store
	backend := v.GetString("store_backend")
	var st store.Store
	switch backend {
	case "postgres":
		key, err := base64.RawStdEncoding.DecodeString(v.GetString("encryption_key"))
		if err != nil {
			log.Panicf("unable to decode encryption key: %s", err)
		}
		st, err = store.NewPostgres(v.GetString("store_postgres_url"), key)
		if err != nil {
			log.Panicf("unable to open postgres database: %s", err)
		}
	case "bolt":
		var err error
		st, err = store.NewBolt(v.GetString("store_bolt_path"))
		if err != nil {
			log.Panicf("unable to create bolt database: %s", err)
		}
	case "dummy":
		log.Panicf("dummy store backend is not wired up")
	default:
		log.Panicf("unknown store backend: %s", backend)
	}

	// create message dispatcher
	dispatch.Start()

	// setup puller to store messages
	puller, err := store.NewPuller(st)
	if err != nil {
		log.Panicf("pull not able to connect, got err: %s", err)
	}
	go puller.Start()

	// create bot manager
	// TODO: consider who is responsible for making sure a given user's bot is
	// running
	//botManager := bot.NewManager()

	// create stream manager
	streamManager := stream.NewManager(twitchClient)

	mux := http.NewServeMux()

	// wire up oauth handler
	doneHandler := oauth.NewDoneHandler(
		v.GetString("twitch_oauth_client_id"),
		v.GetString("twitch_oauth_client_secret"),
		v.GetString("twitch_oauth_redirect_uri"),
		st,
		twitchClient,
	)
	mux.Handle("/twitch_oauth/done", doneHandler)

	// setup websocket API server
	api := api.New(
		streamManager,
		st,
		twitchClient,
		doneHandler,
		v.GetString("twitch_oauth_client_id"),
	)
	mux.Handle("/api", api)

	// bind websocket API
	v.SetDefault("port", 8080)
	port := v.GetInt("port")

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
		// TODO: consider timeouts
	}

	certFile := v.GetString("tls_cert_file")
	keyFile := v.GetString("tls_key_file")
	if certFile != "" && keyFile != "" {
		fmt.Println("listening for tls on port", port)
		err = server.ListenAndServeTLS(certFile, keyFile)
		if err != nil {
			log.Panic("ListenAndServeTLS: " + err.Error())
		}
		return
	}

	fmt.Println("listening on port", port)
	err = server.ListenAndServe()
	if err != nil {
		log.Panic("ListenAndServe: " + err.Error())
	}
}
