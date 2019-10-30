package api

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/mullvad/message-queue/api/handler"
	"github.com/mullvad/message-queue/queue"
	"nhooyr.io/websocket"
)

const subProtocol = "message-queue-v1"

// API is a http API
type API struct {
	Queue        *queue.Queue
	PingTimeout  time.Duration
	PingInterval time.Duration
}

// New returns a new instance of the API with default settings
func New(q *queue.Queue) *API {
	return &API{
		Queue:        q,
		PingTimeout:  time.Minute,
		PingInterval: time.Second * 15,
	}
}

// Router returns a http router
func (a *API) Router() http.Handler {
	router := mux.NewRouter()

	router.NotFoundHandler = handler.Handler(a.notFoundHandler)

	// Redirect trailing slashes
	router.StrictSlash(true)

	router.Handle("/channel/{channel}", handler.Handler(a.handleChannel))

	return handler.Recovery(router)
}

var notFoundError = &handler.Error{
	Message: "page not found",
	Code:    http.StatusNotFound,
}

func (a *API) notFoundHandler(w http.ResponseWriter, r *http.Request) *handler.Error {
	return notFoundError
}

func (a *API) handleChannel(w http.ResponseWriter, r *http.Request) *handler.Error {
	vars := mux.Vars(r)
	channel := vars["channel"]

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	ch, err := a.Queue.Subscribe(ctx, channel)
	if err != nil {
		return handler.BadRequest("invalid channel")
	}

	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		Subprotocols: []string{subProtocol},
	})

	if err != nil {
		log.Println("error upgrading connection to websocket", err)
		return handler.InternalServerError()
	}

	defer c.Close(websocket.StatusInternalError, "the sky is falling")

	if c.Subprotocol() != subProtocol {
		c.Close(websocket.StatusPolicyViolation, "client must speak the correct subprotocol")
		return nil
	}

	// Set up reader to handle pings, but terminate the connection if we receieve any messages
	ctx = c.CloseRead(ctx)

	// Close the connection if no successful ping has occured within the ping timeout
	pingTimer := time.AfterFunc(a.PingTimeout, cancel)
	defer pingTimer.Stop()
	pingTicker := time.NewTicker(a.PingInterval)
	defer pingTicker.Stop()

	// Subscribe to the queue
	for {
		select {
		case <-ctx.Done():
			c.Close(websocket.StatusNormalClosure, "")
			return nil
		case <-pingTicker.C:
			func() {
				pingCtx, cancel := context.WithTimeout(ctx, a.PingTimeout)
				defer cancel()

				err := c.Ping(pingCtx)

				// Reset ping timer on successful ping response
				if err == nil {
					if !pingTimer.Stop() {
						<-pingTimer.C
					}

					pingTimer.Reset(a.PingTimeout)
				}
			}()
		case msg, open := <-ch:
			// Channel has been closed, close the connection
			if !open {
				c.Close(websocket.StatusInternalError, "something went wrong")
				return nil
			}

			err := c.Write(ctx, websocket.MessageText, msg)
			if err != nil {
				log.Println("error sending message", err)
			}
		}
	}
}
