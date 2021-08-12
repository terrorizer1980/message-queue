package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/infosum/statsd"
	"github.com/mullvad/message-queue/pubsub"
	"github.com/mullvad/message-queue/queue"

	"github.com/jamiealquiza/envy"
	"github.com/mullvad/message-queue/api"
)

var (
	metrics *statsd.Client
	p       *pubsub.PubSub
	q       *queue.Queue
)

func main() {
	// Set up commandline flags
	listen := flag.String("listen", ":8080", "listen address")
	bufferSize := flag.Int("buffer-size", 100, "client buffer size")
	redisSentinelService := flag.String("redis-sentinel-service", "", "redis sentinel service name")
	redisSentinelAddrs := flag.String("redis-sentinel-addresses", "", "comma-delimited list of redis sentinel addresses, may contain authentication details")
	redisServerAddress := flag.String("redis-server-address", "", "address for the redis server to connect to redis directly (without redis sentinel)")
	redisPassword := flag.String("redis-server-password", "", "password for the redis servers managed by redis sentinel")
	channels := flag.String("channels", "", "comma-delimited list of channels to listen and broadcast to")
	statsdAddress := flag.String("statsd-address", "127.0.0.1:8125", "statsd address to send metrics to")

	// Parse environment variables
	envy.Parse("MQ")

	// Parse commandline flags
	flag.Parse()

	if *redisSentinelAddrs == "" && *redisServerAddress == "" {
		log.Fatalf("either '-redis-sentinel-addresses' or '-redis-server-address' is required")
	}

	if *redisSentinelAddrs != "" && *redisServerAddress != "" {
		log.Fatalf("'-redis-sentinel-addresses' and '-redis-server-address' are incompatible")
	}

	if *redisSentinelAddrs != "" && *redisSentinelService == "" {
		log.Fatalf("'-redis-sentinel-service' is required when using redis sentinel")
	}

	if *redisPassword == "" {
		log.Fatalf("'-redis-server-password' is required")
	}

	redisSentinelAddrList := strings.Split(*redisSentinelAddrs, ",")

	if *channels == "" {
		log.Fatalf("no channels configured")
	}

	channelList := strings.Split(*channels, ",")

	log.Printf("starting message-queue")

	// Initialize metrics
	var err error
	metrics, err = statsd.New(statsd.TagsFormat(statsd.Datadog), statsd.Prefix("mq"), statsd.Address(*statsdAddress))
	if err != nil {
		log.Fatal("Error initializing metrics: ", err)
	}
	defer metrics.Close()

	// Set up context for shutting down
	shutdownCtx, shutdown := context.WithCancel(context.Background())
	defer shutdown()

	// Set up the pubsub listener
	if *redisSentinelService != "" {
		p, err = pubsub.NewWithSentinel(*redisSentinelService, redisSentinelAddrList, *redisPassword)
	} else {
		p, err = pubsub.New(*redisServerAddress, *redisPassword)
	}
	if err != nil {
		log.Fatal("error initializing pubsub: ", err)
	}

	// Set up the queue
	q = queue.New(shutdownCtx, *bufferSize)

	// Set up the message passing from redis pubsub to the queue
	err = setupChannels(shutdownCtx, channelList)
	if err != nil {
		log.Fatal("error initializing queue: ", err)
	}

	// Create a ticker for metrics
	ticker := time.NewTicker(time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				collectMetrics()
			case <-shutdownCtx.Done():
				ticker.Stop()
				return
			}
		}
	}()

	// Start and listen on http
	api := api.New(q)

	server := &http.Server{
		Addr:    *listen,
		Handler: api.Router(),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Println("shutting down the http server", err)
			shutdown()
		}
	}()

	log.Printf("http server listening on %s", *listen)

	// Wait for shutdown or error
	err = waitForInterrupt(shutdownCtx)
	log.Println("shutting down", err)

	// Shut down http server
	serverCtx, serverCancel := context.WithTimeout(context.Background(), time.Second*30)
	defer serverCancel()
	if err := server.Shutdown(serverCtx); err != nil {
		log.Println("error shutting down", err)
	}
}

func setupChannels(ctx context.Context, channels []string) error {
	for _, channel := range channels {
		in, err := p.Subscribe(channel)
		if err != nil {
			return err
		}

		out, err := q.CreateChannel(channel)
		if err != nil {
			return err
		}

		go channelWorker(ctx, in, out)
	}

	return nil
}

func channelWorker(ctx context.Context, in <-chan []byte, out chan<- []byte) {
	defer func() {
		close(out)
	}()

	for {
		select {
		case msg, open := <-in:
			if !open {
				return
			}

			out <- msg
		case <-ctx.Done():
			return
		}
	}
}

func collectMetrics() {
	metrics.Gauge("subscribers", q.SubscriberCount())
}

func waitForInterrupt(ctx context.Context) error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-c:
		return fmt.Errorf("received signal %s", sig)
	case <-ctx.Done():
		return errors.New("canceled")
	}
}
