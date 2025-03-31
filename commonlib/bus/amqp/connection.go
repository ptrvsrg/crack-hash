package amqp

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"atomicgo.dev/robin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"sync/atomic"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	timeout = time.Second
)

var (
	ErrUrlsIsEmpty = errors.New("urls is empty")
)

type (
	Config struct {
		URI      string
		Username string
		Password string
		Prefetch int
	}

	ClusterConfig struct {
		URIs     []string
		Username string
		Password string
		Prefetch int
	}

	// Connection amqp.Connection wrapper
	Connection struct {
		conn *amqp.Connection

		opts     amqp.Config
		balancer *robin.Loadbalancer[string]
		logger   zerolog.Logger
		prefetch int

		// Reconnect
		reconnectLock sync.RWMutex
		reconnect     atomic.Bool

		// Watcher lifecycle
		wg     sync.WaitGroup
		cancel context.CancelFunc

		// State
		closed atomic.Bool
	}

	// Channel amqp.Channel wapper
	Channel struct {
		ch *amqp.Channel

		conn   *Connection
		logger zerolog.Logger

		// Reconnect
		reconnectLock sync.RWMutex
		reconnect     atomic.Bool

		// Watcher lifecycle
		wg     sync.WaitGroup
		cancel context.CancelFunc

		// State
		closed atomic.Bool
	}
)

// Dial wrap amqp.Dial, dial and get a reconnect connection
func Dial(ctx context.Context, cfg Config) (*Connection, error) {
	// Connect to RabbitMQ
	opts := amqp.Config{
		SASL: []amqp.Authentication{
			&amqp.PlainAuth{
				Username: cfg.Username,
				Password: cfg.Password,
			},
		},
	}

	balancer := robin.NewLoadbalancer([]string{cfg.URI})

	origConn, err := amqp.DialConfig(balancer.Next(), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to dial amqp: %w", err)
	}

	// Create context for watcher
	ctx, cancel := context.WithCancel(ctx)

	conn := &Connection{
		conn: origConn,

		opts:     opts,
		balancer: balancer,
		logger: log.With().
			Str("component", "amqp-connection").
			Str("mode", "standalone").
			Logger(),
		prefetch: cfg.Prefetch,

		reconnectLock: sync.RWMutex{},
		reconnect:     atomic.Bool{},

		wg:     sync.WaitGroup{},
		cancel: cancel,

		closed: atomic.Bool{},
	}

	go conn.runWatcher(ctx)

	return conn, nil
}

// DialCluster with reconnect
func DialCluster(ctx context.Context, cfg ClusterConfig) (*Connection, error) {
	if len(cfg.URIs) == 0 {
		return nil, ErrUrlsIsEmpty
	}

	balancer := robin.NewLoadbalancer(cfg.URIs)

	logger := log.With().
		Str("component", "amqp-connection").
		Str("mode", "cluster").
		Logger()

	// Connect to one from RabbitMQ node
	opts := amqp.Config{
		SASL: []amqp.Authentication{
			&amqp.PlainAuth{
				Username: cfg.Username,
				Password: cfg.Password,
			},
		},
	}

	var (
		origConn *amqp.Connection
		err      error
		joinErr  error
	)
	for i := 0; i < len(cfg.URIs); i++ {
		uri := balancer.Next()
		logger.Debug().Str("node", uri).Msg("dial amqp")

		origConn, err = amqp.DialConfig(uri, opts)
		if err == nil {
			break
		}

		logger.Error().Err(err).Str("node", uri).Msg("failed to dial amqp")
		joinErr = errors.Join(joinErr, err)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to dial amqp: %w", joinErr)
	}

	// Create context for watcher
	ctx, cancel := context.WithCancel(ctx)

	conn := &Connection{
		conn: origConn,

		opts:     opts,
		balancer: balancer,
		logger:   logger,
		prefetch: cfg.Prefetch,

		reconnectLock: sync.RWMutex{},
		reconnect:     atomic.Bool{},

		wg:     sync.WaitGroup{},
		cancel: cancel,

		closed: atomic.Bool{},
	}

	go conn.runWatcher(ctx)

	return conn, nil
}

// runWatcher watch connection state for reconnection
func (c *Connection) runWatcher(ctx context.Context) {
	c.logger.Info().Msg("connection watcher started")

	c.wg.Add(1)
	defer c.wg.Done()

	for {
		select {
		case <-ctx.Done():
			c.logger.Info().Msg("connection watcher stopped")
			return

		case reason, ok := <-c.GetConnection().NotifyClose(make(chan *amqp.Error)):
			if !ok {
				c.logger.Info().Msg("connection watcher stopped")
				return
			}

			// lock connection for reconnect
			c.reconnectLock.Lock()
			c.reconnect.Store(true)

			c.logger.Error().Err(reason).Msg("connection closed, try to reconnect")

			// reconnect
			for {
				c.logger.Debug().Dur("timeout", timeout).Msg("try reconnect")
				time.Sleep(timeout)

				// check closed
				if c.IsClosed() {
					break
				}

				// try next node
				uri := c.balancer.Next()
				c.logger.Debug().Str("node", uri).Msg("dial amqp")

				origConn, err := amqp.DialConfig(uri, c.opts)
				if err != nil {
					c.logger.Error().Err(err).Str("node", uri).Msg("failed to reconnect")
					continue
				}

				// set new amqp.Connection
				c.conn = origConn
				break
			}

			// unlock connection for reconnect
			c.reconnect.Store(false)
			c.reconnectLock.Unlock()
			c.logger.Info().Msg("connection reconnected")
		}
	}
}

// GetConnection get amqp.Connection
func (c *Connection) GetConnection() *amqp.Connection {
	c.reconnectLock.RLock()
	defer c.reconnectLock.RUnlock()

	return c.conn
}

// Channel wrap amqp.Connection.Channel, get a auto reconnect channel
func (c *Connection) Channel(ctx context.Context) (*Channel, error) {
	// Open a channel
	origCh, err := c.GetConnection().Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	// set prefetch
	if err := origCh.Qos(c.prefetch, 0, false); err != nil {
		_ = origCh.Close()
		return nil, fmt.Errorf("failed to set prefetch: %w", err)
	}

	// Create context for watcher
	ctx, cancel := context.WithCancel(ctx)

	ch := &Channel{
		ch: origCh,

		conn:   c,
		logger: log.With().Str("component", "amqp-channel").Logger(),

		reconnectLock: sync.RWMutex{},
		reconnect:     atomic.Bool{},

		wg:     sync.WaitGroup{},
		cancel: cancel,

		closed: atomic.Bool{},
	}

	go ch.runWatcher(ctx)

	return ch, nil
}

// IsReconnect indicate reconnect
func (c *Connection) IsReconnect() bool {
	return c.reconnect.Load()
}

// IsClosed indicate closed by developer
func (c *Connection) IsClosed() bool {
	return c.closed.Load()
}

// Close ensure closed flag set
func (c *Connection) Close() error {
	if c.IsClosed() {
		return amqp.ErrClosed
	}

	c.closed.Store(true)
	c.cancel()
	c.wg.Wait()

	if err := c.GetConnection().Close(); err != nil {
		return fmt.Errorf("failed to close amqp connection: %w", err)
	}

	return nil
}

// runWatcher run channel watcher
func (ch *Channel) runWatcher(ctx context.Context) {
	ch.logger.Info().Msg("channel watcher started")

	ch.wg.Add(1)
	defer ch.wg.Done()

	for {
		select {
		case <-ctx.Done():
			ch.logger.Info().Msg("channel watcher stopped")
			return

		case reason := <-ch.GetChannel().NotifyClose(make(chan *amqp.Error)):
			// lock channel for reconnect
			ch.reconnectLock.Lock()
			ch.reconnect.Store(true)
			ch.logger.Error().Err(reason).Msg("channel closed, try to reconnect")

			// reconnect if not closed by developer
			for {
				ch.logger.Debug().Dur("timeout", timeout).Msg("try reconnect")
				time.Sleep(timeout)

				// check closed
				if ch.IsClosed() {
					break
				}

				// open a new channel
				origCh, err := ch.conn.GetConnection().Channel()
				if err != nil {
					ch.logger.Error().Err(err).Msg("failed to reconnect")
					continue
				}

				// set prefetch
				if err := origCh.Qos(ch.conn.prefetch, 0, false); err != nil {
					_ = origCh.Close()
					ch.logger.Error().Err(err).Msg("failed to set prefetch")
					continue
				}

				// set new amqp.Channel
				ch.ch = origCh
				break
			}

			// unlock channel for reconnect
			ch.reconnect.Store(false)
			ch.reconnectLock.Unlock()
			ch.logger.Info().Msg("channel reconnected")
		}
	}
}

// GetChannel get amqp.Channel
func (ch *Channel) GetChannel() *amqp.Channel {
	ch.reconnectLock.RLock()
	defer ch.reconnectLock.RUnlock()

	return ch.ch
}

// IsReconnect indicate reconnect
func (ch *Channel) IsReconnect() bool {
	return ch.reconnect.Load()
}

// IsClosed indicate closed by developer
func (ch *Channel) IsClosed() bool {
	return ch.closed.Load()
}

// Close ensure closed flag set
func (ch *Channel) Close() error {
	if ch.IsClosed() {
		return amqp.ErrClosed
	}

	ch.closed.Store(true)
	ch.cancel()
	ch.wg.Wait()

	if err := ch.GetChannel().Close(); err != nil {
		return fmt.Errorf("failed to close amqp channel: %w", err)
	}

	return nil
}

// Publish wrap amqp.Channel.Publish
func (ch *Channel) Publish(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	if err := ch.GetChannel().PublishWithContext(ctx, exchange, key, mandatory, immediate, msg); err != nil {
		return fmt.Errorf("failed to publish a message: %w", err)
	}

	return nil
}

// Consume wrap amqp.Channel.Consume, the returned delivery will end only when channel closed by developer
func (ch *Channel) Consume(
	ctx context.Context, queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table,
) <-chan amqp.Delivery {
	deliveries := make(chan amqp.Delivery)
	go ch.runConsumer(ctx, deliveries, queue, consumer, autoAck, exclusive, noLocal, noWait, args)
	return deliveries
}

// runConsumer run channel consumer
func (ch *Channel) runConsumer(
	ctx context.Context, deliveries chan<- amqp.Delivery, queue, consumer string, autoAck, exclusive, noLocal,
	noWait bool, args amqp.Table,
) {
	for {
		d, err := ch.GetChannel().ConsumeWithContext(ctx, queue, consumer, autoAck, exclusive, noLocal, noWait, args)
		if err != nil {
			ch.logger.Error().Err(err).Msg("failed to consume")
			time.Sleep(timeout)
			continue
		}

		for msg := range d {
			deliveries <- msg
		}

		// sleep before IsClose call. closed flag may not set before sleep.
		time.Sleep(timeout)

		if ch.IsClosed() {
			close(deliveries)
			break
		}
	}
}
