package client

import (
	"fmt"
	"net/http"
	"time"

	"golang.org/x/net/http2"
	"resty.dev/v3"

	"github.com/ptrvsrg/crack-hash/commonlib/http/client/loadbalancer"
)

type Option func(*resty.Client) error

func New(opts ...Option) (*resty.Client, error) {
	client := resty.New()

	transport, ok := client.Transport().(*http.Transport)
	if ok {
		if err := http2.ConfigureTransport(transport); err == nil {
			client.SetTransport(transport)
		}
	}

	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	return client, nil
}

func WithRetries(retries int, minWaitTime, maxWaitTime time.Duration) Option {
	return func(c *resty.Client) error {
		c.SetRetryCount(retries)
		c.SetRetryWaitTime(minWaitTime)
		c.SetRetryMaxWaitTime(maxWaitTime)

		return nil
	}
}

func WithLoadBalancer(urls []string, opts ...loadbalancer.Option) Option {
	return func(c *resty.Client) error {
		lb, err := loadbalancer.NewRoundRobin(urls, opts...)
		if err != nil {
			return fmt.Errorf("failed to setup load balancer: %w", err)
		}

		c.SetLoadBalancer(lb)
		return nil
	}
}
