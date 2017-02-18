package store

import (
	"anubot/stream"
	"encoding/json"
	"log"

	"github.com/pebbe/zmq4"
)

// MessageStorer stores messages.
type MessageStorer interface {
	StoreMessage(msg stream.RXMessage) (err error)
}

// Puller pulls messages from dispatch and stores them.
type Puller struct {
	store         MessageStorer
	pull          *zmq4.Socket
	pullEndpoints []string
	stop          chan struct{}
	done          chan struct{}
}

// Option is used to configure a Puller.
type Option func(*Puller)

// WithPullEndpoints allows you to override the default pull endpoints.
func WithPullEndpoints(endpoints []string) Option {
	return func(p *Puller) {
		p.pullEndpoints = endpoints
	}
}

// NewPuller returns a new puller.
func NewPuller(store MessageStorer, opts ...Option) (*Puller, error) {
	p := &Puller{
		store:         store,
		pullEndpoints: []string{"inproc://dispatch-push"},
		stop:          make(chan struct{}),
		done:          make(chan struct{}),
	}
	for _, opt := range opts {
		opt(p)
	}
	err := p.setupSockets()
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Puller) setupSockets() error {
	pull, err := zmq4.NewSocket(zmq4.PULL)
	if err != nil {
		return err
	}
	for _, endpoint := range p.pullEndpoints {
		err = pull.Connect(endpoint)
		if err != nil {
			return err
		}
	}
	p.pull = pull
	return nil
}

// Start reads messages from pull socket and stores them. It needs to run in
// its own goroutine.
func (p *Puller) Start() {
	defer close(p.done)

	for {
		select {
		case <-p.stop:
			return
		default:
		}

		rb, err := p.pull.RecvBytes(0)
		if err != nil {
			log.Printf("messages not read, got err: %s", err)
			continue
		}
		var ms stream.RXMessage
		err = json.Unmarshal(rb, &ms)
		if err != nil {
			log.Printf("could not unmarshal, got err: %s", err)
			continue
		}

		err = p.store.StoreMessage(ms)
		if err != nil {
			log.Printf("could not store message, got err: %s", err)
			continue
		}
	}
}

// Stop signals to the goroutine reading messages to stop. It returns a
// function that can be used to block until reading has finished.
func (p *Puller) Stop() (wait func()) {
	close(p.stop)
	return func() {
		<-p.done
	}
}
