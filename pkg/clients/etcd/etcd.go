package etcd

import (
	"context"
	"strings"
	"time"

	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/log"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Client struct {
	client     *clientv3.Client
	eventsChan chan struct{ Key, Value string }
	logger     log.Logger
}

type Config struct {
	Endpoints   []string      `mapstructure:"endpoints"`
	DialTimeout time.Duration `mapstructure:"dial_timeout"`
}

func NewClient(cfg Config, logger log.Logger) (Client, error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   cfg.Endpoints,
		DialTimeout: cfg.DialTimeout,
	})
	if err != nil {
		return Client{}, errors.Wrap(err, "failed to init etcd client")
	}

	return Client{
		client:     client,
		eventsChan: make(chan struct{ Key, Value string }),
		logger:     logger,
	}, nil
}

func (c *Client) WatchKeysByPrefix(prefix string) {
	c.logger.Info("start watch changes from etcd")
	defer c.logger.Info("stop watch changes from etcd")

	eventChan := c.client.Watch(context.Background(), prefix, clientv3.WithPrefix())
	for {
		event := <-eventChan

		for _, evt := range event.Events {
			c.eventsChan <- struct {
				Key,
				Value string
			}{
				Key:   strings.TrimPrefix(string(evt.Kv.Key), prefix),
				Value: string(evt.Kv.Value),
			}
		}
	}
}

func (c Client) GetKeysByPrefix(prefix string) error {
	resp, err := c.client.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		return errors.Wrapf(err, "failed to get keys from etcd by prefix: %s", prefix)
	}

	for _, kv := range resp.Kvs {
		c.eventsChan <- struct {
			Key,
			Value string
		}{
			Key:   strings.TrimPrefix(string(kv.Key), prefix),
			Value: string(kv.Value),
		}
	}

	return nil
}

func (c Client) Chan() <-chan struct{ Key, Value string } {
	return c.eventsChan
}

func (c Client) Close() {
	if err := c.client.Close(); err != nil {
		c.logger.WithError(err).Error("failed to close connection to etcd")
	}
}
