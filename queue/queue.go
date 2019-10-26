package queue

import (
	"context"
	"fmt"
	"sync"
)

// Queue is a message queue with multiple channels, where every message is broadcast to all subscribers of a channel
// It will automatically close the channels for and remove subscribers whose context has ended, or buffer is full
type Queue struct {
	channels   map[string]channel
	mutex      sync.Mutex
	ctx        context.Context
	bufferSize int // The message buffer size for each subscriber to a channel
}

type channel struct {
	queue       <-chan interface{}
	subscribers map[subscriber]struct{}
}

type subscriber struct {
	channel chan<- interface{}
	context context.Context
}

// New creates a new queue
func New(ctx context.Context, bufferSize int) *Queue {
	return &Queue{
		channels:   make(map[string]channel),
		ctx:        ctx,
		bufferSize: bufferSize,
	}
}

// CreateChannel creates a new queue channel and returns a channel for broadcasting to it
func (q *Queue) CreateChannel(channelName string) (chan<- interface{}, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	_, ok := q.channels[channelName]
	if ok {
		return nil, fmt.Errorf("%q: channel already exists", channelName)
	}

	ch := make(chan interface{})

	c := channel{
		queue:       ch,
		subscribers: make(map[subscriber]struct{}),
	}

	q.channels[channelName] = c

	go q.worker(channelName, &c)

	return ch, nil
}

func (q *Queue) worker(channelName string, c *channel) {
	defer q.cleanup(channelName, c)

	for {
		select {
		case message, open := <-c.queue:
			// The channel has been closed, exit
			if !open {
				return
			}

			func() {
				q.mutex.Lock()
				defer q.mutex.Unlock()

				for subscriber := range c.subscribers {
					// Check the subscribers context
					select {
					// If it's done, remove the subscriber
					case <-subscriber.context.Done():
						c.removeSubscriber(subscriber)
					default:
						// Otherwise, try to write to the subscribers channel
						select {
						case subscriber.channel <- message:
						// If the write fails (buffer is full), remove the subscriber
						default:
							c.removeSubscriber(subscriber)
						}
					}
				}
			}()
		case <-q.ctx.Done():
			return
		}
	}
}

func (q *Queue) cleanup(channelName string, c *channel) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	for subscriber := range c.subscribers {
		c.removeSubscriber(subscriber)
	}

	delete(q.channels, channelName)
}

func (c *channel) removeSubscriber(s subscriber) {
	delete(c.subscribers, s)
	close(s.channel)
}

// Subscribe subscribes to a queue channel, and returns a channel for receiving messages
// Consumers should check whether the channel is closed, as the queue may terminate subscriptions at any time
// Returns an error if the given channel doesn't exist
func (q *Queue) Subscribe(context context.Context, channelName string) (<-chan interface{}, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	_, ok := q.channels[channelName]
	if !ok {
		return nil, fmt.Errorf("%q: channel doesn't exist", channelName)
	}

	channel := make(chan interface{}, q.bufferSize)

	newChannel := q.channels[channelName]
	s := subscriber{
		context: context,
		channel: channel,
	}
	newChannel.subscribers[s] = struct{}{}
	q.channels[channelName] = newChannel

	return channel, nil
}
