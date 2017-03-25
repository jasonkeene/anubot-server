package twitch_test

import (
	"flag"
	"io/ioutil"
	"log"
	"testing"
)

func init() {
	flag.Parse()
	if !testing.Verbose() {
		log.SetOutput(ioutil.Discard)
	}
}
