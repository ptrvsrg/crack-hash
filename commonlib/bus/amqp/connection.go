package amqp

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"sync/atomic"

	"github.com/streadway/amqp"
)

var (
	ErrUrlsIsEmpty = errors.New("urls is empty")

	BackoffPolicy = []time.Duration{
		1 * time.Second,
		2 * time.Second,
		5 * time.Second,
		10 * time.Second,
		15 * time.Second,
		20 * time.Second,
		25 * time.Second,
	}
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
		*amqp.Connection

		logger   zerolog.Logger
		prefetch int

		reconnect atomic.Bool
		closed    atomic.Bool
		rw        sync.RWMutex
		cancel    context.CancelFunc
		wg        sync.WaitGroup
	}

	// Channel amqp.Channel wapper
	Channel struct {
		*amqp.Channel

		logger zerolog.Logger

		closed atomic.Bool
		rw     sync.RWMutex
		cancel context.CancelFunc
		wg     sync.WaitGroup
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

	origConn, err := amqp.DialConfig(cfg.URI, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to dial amqp: %w", err)
	}

	// Create context for watcher
	ctx, cancel := context.WithCancel(ctx)

	conn := &Connection{
		Connection: origConn,

		logger: log.With().
			Str("component", "amqp-connection").
			Str("mode", "standalone").
			Logger(),
		prefetch: cfg.Prefetch,

		rw:        sync.RWMutex{},
		cancel:    cancel,
		wg:        sync.WaitGroup{},
		reconnect: atomic.Bool{},
		closed:    atomic.Bool{},
	}

	// Start watcher
	go func() {
		conn.logger.Info().Msg("connection watcher started")

		conn.wg.Add(1)
		defer conn.wg.Done()

		for {
			select {
			case <-ctx.Done():
				conn.logger.Info().Msg("connection watcher stopped")
				return

			default:
				// check if connection is closed
				reason, ok := <-conn.getConnection().NotifyClose(make(chan *amqp.Error))
				if !ok {
					time.Sleep(time.Second)
					continue
				}

				conn.logger.Error().Err(reason).Msg("connection closed")

				// set flag
				conn.reconnect.Store(true)

				// reconnect if not closed by developer
				for _, timeout := range BackoffPolicy {
					conn.logger.Debug().Dur("timeout", timeout).Msg("try reconnect")

					origConn, err := amqp.DialConfig(cfg.URI, opts)
					if err != nil {
						conn.logger.Error().Err(err).Msg("failed to reconnect")
						time.Sleep(timeout)
						continue
					}

					// set new amqp.Connection and reset flag
					conn.setConnection(origConn)
					conn.reconnect.Store(false)

					conn.logger.Info().Msg("reconnect success")

					break
				}
			}
		}
	}()

	return conn, nil
}

// DialCluster with reconnect
func DialCluster(ctx context.Context, cfg ClusterConfig) (*Connection, error) {
	if len(cfg.URIs) == 0 {
		return nil, ErrUrlsIsEmpty
	}

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

	nodeSequence := 0

	var (
		origConn *amqp.Connection
		err      error
	)
	for i := 0; i < len(cfg.URIs); i++ {
		logger.Debug().Str("node", cfg.URIs[nodeSequence]).Msg("dial amqp")

		origConn, err = amqp.DialConfig(cfg.URIs[nodeSequence], opts)
		if err != nil {
			nodeSequence = next(cfg.URIs, nodeSequence)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to dial amqp: %w", err)
	}

	// Create context for watcher
	ctx, cancel := context.WithCancel(ctx)

	conn := &Connection{
		Connection: origConn,

		logger:   logger,
		prefetch: cfg.Prefetch,

		rw:        sync.RWMutex{},
		cancel:    cancel,
		closed:    atomic.Bool{},
		reconnect: atomic.Bool{},
	}

	// Start watcher
	go func() {
		conn.logger.Info().Msg("connection watcher started")

		conn.wg.Add(1)
		defer conn.wg.Done()

		errCh := conn.getConnection().NotifyClose(make(chan *amqp.Error))

		for {
			select {
			case <-ctx.Done():
				conn.logger.Info().Msg("connection watcher stopped")
				return

			case reason, ok := <-errCh:
				if !ok {
					errCh = conn.getConnection().NotifyClose(make(chan *amqp.Error))
					time.Sleep(time.Second)
					continue
				}

				conn.logger.Error().Err(reason).Msg("connection closed")

				// set flag
				conn.reconnect.Store(true)

				// reconnect if not closed by developer
				for _, timeout := range BackoffPolicy {
					conn.logger.Debug().Dur("timeout", timeout).Msg("try reconnect")

					// try next node
					nodeSequence = next(cfg.URIs, nodeSequence)
					logger.Debug().Str("node", cfg.URIs[nodeSequence]).Msg("dial amqp")

					origConn, err := amqp.DialConfig(cfg.URIs[nodeSequence], opts)
					if err != nil {
						conn.logger.Error().Err(err).Msg("failed to reconnect")
						time.Sleep(timeout)
						continue
					}

					// set new amqp.Connection and reset flag
					conn.setConnection(origConn)
					conn.reconnect.Store(false)

					conn.logger.Info().Msg("reconnect success")

					break
				}
			}
		}
	}()

	return conn, nil
}

// getConnection get amqp.Connection
func (c *Connection) getConnection() *amqp.Connection {
	c.rw.RLock()
	defer c.rw.RUnlock()

	return c.Connection
}

// setConnection set amqp.Connection
func (c *Connection) setConnection(conn *amqp.Connection) {
	c.rw.Lock()
	defer c.rw.Unlock()

	c.Connection = conn
}

// Channel wrap amqp.Connection.Channel, get a auto reconnect channel
func (c *Connection) Channel(ctx context.Context) (*Channel, error) {
	// Open a channel
	origCh, err := c.getConnection().Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a ch: %w", err)
	}

	// set prefetch
	if err := origCh.Qos(c.prefetch, 0, false); err != nil {
		return nil, fmt.Errorf("failed to set prefetch: %w", err)
	}

	// Create context for watcher
	ctx, cancel := context.WithCancel(ctx)

	ch := &Channel{
		Channel: origCh,

		logger: log.With().Str("component", "amqp-channel").Logger(),

		rw:     sync.RWMutex{},
		cancel: cancel,
		wg:     sync.WaitGroup{},
		closed: atomic.Bool{},
	}

	// Start watcher
	go func() {
		ch.logger.Info().Msg("channel watcher started")

		ch.wg.Add(1)
		defer ch.wg.Done()

		errCh := ch.getChannel().NotifyClose(make(chan *amqp.Error))

		for {
			select {
			case <-ctx.Done():
				ch.logger.Info().Msg("channel watcher stopped")
				return

			case reason, ok := <-errCh:
				if !ok {
					errCh = ch.getChannel().NotifyClose(make(chan *amqp.Error))
					time.Sleep(time.Second)
					continue
				}

				ch.logger.Error().Err(reason).Msg("channel closed")

				// wait for reconnect connection
				for c.reconnect.Load() {
					ch.logger.Debug().Msg("wait for reconnect connection")
					time.Sleep(time.Second)
				}

				// reconnect if not closed by developer
				for _, timeout := range BackoffPolicy {
					ch.logger.Debug().Dur("timeout", timeout).Msg("try reconnect")

					// open a new channel
					origCh, err := c.getConnection().Channel()
					if err != nil {
						ch.logger.Error().Err(err).Msg("failed to reconnect")
						time.Sleep(timeout)
						continue
					}

					// set prefetch
					if err := origCh.Qos(c.prefetch, 0, false); err != nil {
						ch.logger.Error().Err(err).Msg("failed to set prefetch")
						time.Sleep(timeout)
						continue
					}

					// set new amqp.Channel
					ch.setChannel(origCh)

					ch.logger.Info().Msg("reconnect success")

					break
				}
			}
		}

	}()

	return ch, nil
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

	if err := c.getConnection().Close(); err != nil {
		return fmt.Errorf("failed to close amqp connection: %w", err)
	}

	return nil
}

// Next element index of slice
func next(s []string, lastSeq int) int {
	length := len(s)
	if length == 0 || lastSeq == length-1 {
		return 0
	} else if lastSeq < length-1 {
		return lastSeq + 1
	} else {
		return -1
	}
}

// getChannel get amqp.Channel
func (ch *Channel) getChannel() *amqp.Channel {
	ch.rw.RLock()
	defer ch.rw.RUnlock()

	return ch.Channel
}

// setChannel set amqp.Channel
func (ch *Channel) setChannel(conn *amqp.Channel) {
	ch.rw.Lock()
	defer ch.rw.Unlock()

	ch.Channel = conn
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

	if err := ch.getChannel().Close(); err != nil {
		return fmt.Errorf("failed to close amqp channel: %w", err)
	}

	return nil
}

// Consume wrap amqp.Channel.Consume, the returned delivery will end only when channel closed by developer
func (ch *Channel) Consume(
	queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table,
) (<-chan amqp.Delivery, error) {
	deliveries := make(chan amqp.Delivery)

	go func() {
		for {
			d, err := ch.getChannel().Consume(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
			if err != nil {
				ch.logger.Error().Err(err).Msg("failed to consume")
				time.Sleep(time.Second)
				continue
			}

			for msg := range d {
				deliveries <- msg
			}

			// sleep before IsClose call. closed flag may not set before sleep.
			time.Sleep(time.Second)

			if ch.IsClosed() {
				break
			}
		}
	}()

	return deliveries, nil
}
