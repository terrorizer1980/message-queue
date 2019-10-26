package queue_test

import (
	"context"
	"testing"

	"github.com/mullvad/message-queue/queue"
)

func TestCreateChannel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	q := queue.New(ctx, 100)

	t.Run("error on duplicate channel", func(t *testing.T) {
		_, err := q.CreateChannel("duplicate")
		if err != nil {
			t.Fatal(err)
		}

		_, err = q.CreateChannel("duplicate")
		if err == nil {
			t.Fatal("No error")
		}
	})
}

func TestSubscribe(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	q := queue.New(ctx, 100)

	t.Run("error on invalid channel", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		_, err := q.Subscribe(ctx, "nonexistent")
		if err == nil {
			t.Fatal("No error")
		}
	})

	t.Run("subscribe and receive messages", func(t *testing.T) {
		channel, err := q.CreateChannel("test")
		if err != nil {
			t.Fatal(err)
		}

		sub, err := q.Subscribe(ctx, "test")
		if err != nil {
			t.Fatal(err)
		}

		channel <- "test"

		message := <-sub
		if message != "test" {
			t.Fatalf("wrong message: %s", message)
		}
	})

	t.Run("subscriber with ended context", func(t *testing.T) {
		channel, err := q.CreateChannel("context")
		if err != nil {
			t.Fatal(err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel the context
		sub, err := q.Subscribe(ctx, "context")
		if err != nil {
			t.Fatal(err)
		}

		channel <- "test"

		_, open := <-sub
		if open {
			t.Fatal("channel not closed")
		}
	})
}

func TestClose(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	q := queue.New(ctx, 0)

	t.Run("test closed broadcast channel", func(t *testing.T) {
		channel, err := q.CreateChannel("close")
		if err != nil {
			t.Fatal(err)
		}

		sub, err := q.Subscribe(ctx, "close")
		if err != nil {
			t.Fatal(err)
		}

		close(channel)

		_, open := <-sub
		if open {
			t.Fatal("channel not closed")
		}

		// Try recreating the channel
		_, err = q.CreateChannel("close")
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("test subscriber that has gone away", func(t *testing.T) {
		channel, err := q.CreateChannel("subscriber")
		if err != nil {
			t.Fatal(err)
		}

		// Ignore the resulting channel
		_, err = q.Subscribe(ctx, "subscriber")
		if err != nil {
			t.Fatal(err)
		}

		channel <- "test"
	})
}
