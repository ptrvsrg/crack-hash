package loadbalancer

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
	"golang.org/x/sync/errgroup"
	"resty.dev/v3"

	"github.com/ptrvsrg/crack-hash/commonlib/http/types"
)

const (
	HostStateInActive HostState = iota
	HostStateActive
)

type (
	HostState int

	Host struct {
		url   string
		state HostState
	}

	RoundRobin struct {
		hosts  []*Host
		lock   *sync.RWMutex
		logger zerolog.Logger

		// Health checks
		healthPath    string
		healthTimeout time.Duration
		healthDelay   time.Duration
		healthRetries int
		healthGroup   *errgroup.Group
		healthCtx     context.Context
		healthCancel  context.CancelFunc
	}

	Option func(*RoundRobin)
)

func WithHealthChecks(path string, timeout, delay time.Duration, retries int) Option {
	return func(rr *RoundRobin) {
		rr.healthPath = path
		rr.healthTimeout = timeout
		rr.healthDelay = delay
		rr.healthRetries = retries

		rr.healthCtx, rr.healthCancel = context.WithCancel(context.Background())
		rr.healthGroup, _ = errgroup.WithContext(rr.healthCtx)
	}
}

func NewRoundRobin(urls []string, opts ...Option) (*RoundRobin, error) {
	hosts := make([]*Host, len(urls))
	for i, u := range urls {
		if _, err := url.Parse(u); err != nil {
			return nil, fmt.Errorf("failed to parse URL: %w", err)
		}

		hosts[i] = &Host{
			url:   u,
			state: HostStateActive,
		}
	}

	rr := &RoundRobin{
		hosts: hosts,
		lock:  new(sync.RWMutex),
		logger: log.With().
			Str("component", "load-balancer").
			Str("strategy", "round-robin").
			Logger(),
	}

	for _, opt := range opts {
		opt(rr)
	}

	if rr.healthGroup != nil {
		go rr.healthCheck()
	}

	return rr, nil
}

func (rr *RoundRobin) Next() (string, error) {
	rr.lock.Lock()
	defer rr.lock.Unlock()

	var best *Host
	for _, h := range rr.hosts {
		if h.state == HostStateActive {
			best = h
			break
		}
	}

	if best == nil {
		return "", resty.ErrNoActiveHost
	}

	return best.url, nil
}

func (rr *RoundRobin) Feedback(_ *resty.RequestFeedback) {
	// no-op
}

func (rr *RoundRobin) CountActiveHosts() int {
	rr.lock.RLock()
	defer rr.lock.RUnlock()

	return lo.CountBy(
		rr.hosts, func(host *Host) bool {
			return host.state == HostStateActive
		},
	)
}

func (rr *RoundRobin) Close() error {
	rr.lock.Lock()
	defer rr.lock.Unlock()

	if rr.healthCancel != nil {
		rr.healthCancel()
	}

	if rr.healthGroup != nil {
		err := rr.healthGroup.Wait()
		if err != nil {
			return fmt.Errorf("failed to wait for health check: %w", err)
		}
	}

	return nil
}

func (rr *RoundRobin) healthCheck() {
	client := resty.New()

	for _, host := range rr.hosts {
		healthUrl := fmt.Sprintf("%s%s", host.url, rr.healthPath)

		rr.healthGroup.Go(
			func() error {
				for {
					select {
					case <-rr.healthCtx.Done():
						return nil

					default:
						// Timeout
						timer := time.NewTimer(rr.healthDelay)
						isCtxDone := false

						select {
						case <-rr.healthCtx.Done():
							isCtxDone = true
							break
						case <-timer.C:
							break
						}

						timer.Stop()
						if isCtxDone {
							return nil
						}

						// Health check start
						rr.logger.Debug().Str("url", healthUrl).Msg("health check started")

						errOutput := &types.ErrorOutput{}
						timeoutCtx, cancel := context.WithTimeout(context.Background(), rr.healthTimeout)

						resp, err := client.R().
							SetContext(timeoutCtx).
							SetError(errOutput).
							SetRetryCount(rr.healthRetries).
							Get(healthUrl)
						cancel()

						if err != nil {
							rr.logger.Error().Err(err).Str("url", healthUrl).Msg("health check failed")
							rr.changeState(host, HostStateInActive)
							continue
						}

						if resp.IsError() {
							err := errors.New(errOutput.Message) // nolint
							rr.logger.Error().Err(err).Str("url", healthUrl).Msg("health check failed")
							rr.changeState(host, HostStateInActive)
							continue
						}

						if host.state == HostStateInActive {
							rr.changeState(host, HostStateActive)
						}

						rr.logger.Debug().Str("url", healthUrl).Msg("health check finished")
					}
				}
			},
		)
	}
}

func (rr *RoundRobin) changeState(host *Host, state HostState) {
	defer rr.lock.Unlock()
	rr.lock.Lock()

	host.state = state
}
