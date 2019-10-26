package pubsub_test

import (
	"sync"
	"testing"

	"github.com/mediocregopher/radix/v3"
	"github.com/mullvad/message-queue/pubsub"
)

const (
	address = "127.0.0.1:6379"
	channel = "test"
	message = "foobar"
)

func TestPubSub(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration tests")
	}

	p, err := pubsub.New(address)
	if err != nil {
		t.Fatal(err)
	}

	defer p.Shutdown()

	ch, err := p.Subscribe(channel)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		msg := <-ch
		if string(msg) != message {
			t.Error("invalid message")
		}
		wg.Done()
	}()

	sendMessage(t)

	wg.Wait()
}

func sendMessage(t *testing.T) {
	t.Helper()

	conn, err := radix.Dial("tcp", address)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	conn.Do(radix.Cmd(nil, "PUBLISH", channel, message))
}
