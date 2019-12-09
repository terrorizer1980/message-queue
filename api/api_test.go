package api_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/mullvad/message-queue/queue"

	"github.com/mullvad/message-queue/api"
	"nhooyr.io/websocket"
)

const (
	subProtocol = "message-queue-v1"
	testMessage = "foobar"
	channel     = "test"
)

func TestAPI(t *testing.T) {
	queueCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	queue := queue.New(queueCtx, 100)

	ch, err := queue.CreateChannel(channel)
	if err != nil {
		t.Fatal(err)
	}

	a := api.New(queue)
	a.PingTimeout = time.Millisecond * 10
	a.PingInterval = a.PingTimeout / 4

	server := httptest.NewServer(a.Router())
	defer server.Close()

	parsedURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("recieve message", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		c, _, err := websocket.Dial(ctx, fmt.Sprintf("ws://%s/channel/%s", parsedURL.Host, channel), &websocket.DialOptions{
			HTTPClient:   server.Client(),
			Subprotocols: []string{subProtocol},
		})

		if err != nil {
			t.Fatal(err)
		}

		defer c.Close(websocket.StatusNormalClosure, "")

		ch <- []byte(testMessage)

		_, message, err := c.Read(ctx)
		if err != nil {
			t.Fatalf(err.Error())
		}

		if string(message) != testMessage {
			t.Errorf("wrong message: %s", message)
		}
	})

	t.Run("ping timeout", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		c, _, err := websocket.Dial(ctx, fmt.Sprintf("ws://%s/channel/%s", parsedURL.Host, channel), &websocket.DialOptions{
			HTTPClient:   server.Client(),
			Subprotocols: []string{subProtocol},
		})

		if err != nil {
			t.Fatal(err)
		}

		defer c.Close(websocket.StatusNormalClosure, "")

		time.Sleep(a.PingTimeout * 2)

		// The connection should now have been closed due to ping timeouts
		_, _, err = c.Read(ctx)
		if err == nil {
			t.Error("no error")
		}
	})

	t.Run("invalid subprotocol", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		c, _, err := websocket.Dial(ctx, fmt.Sprintf("ws://%s/channel/%s", parsedURL.Host, channel), &websocket.DialOptions{
			HTTPClient:   server.Client(),
			Subprotocols: []string{"invalid"},
		})

		if err != nil {
			t.Fatal(err)
		}

		defer c.Close(websocket.StatusNormalClosure, "")

		_, _, err = c.Read(ctx)
		if err == nil {
			t.Error("no error")
		}
	})

	t.Run("invalid channel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		_, _, err := websocket.Dial(ctx, fmt.Sprintf("ws://%s/channel/invalid", parsedURL.Host), &websocket.DialOptions{
			HTTPClient:   server.Client(),
			Subprotocols: []string{"invalid"},
		})

		if err == nil {
			t.Error("no error")
		}
	})
}
