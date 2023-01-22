package dwh

import (
	"context"
	"os"
	"time"

	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/log"
	"github.com/antonpriyma/otus-highload/pkg/stat"
	"github.com/antonpriyma/otus-highload/pkg/utils"
	"github.com/segmentio/kafka-go"
)

var (
	ErrMarshalMessage = utils.NewTypedError("dwh_marshal_fail", "failed to marshal dwh message")
	ErrSendDWH        = utils.NewTypedError("dwh_send_fail", "failed to send message to dwh")
)

type kafkaStat struct {
	SendDuration        stat.TimerCtor `labels:"status"`
	SendCompactDuration stat.TimerCtor `labels:"status"`
}

type kafkaClient struct {
	Writer   *kafka.Writer
	Hostname string
	Stat     kafkaStat
}

type KafkaClientConfig struct {
	Enabled bool `mapstructure:"enabled"`

	Brokers         []string      `mapstructure:"brokers"`
	Topic           string        `mapstructure:"topic"`
	MaxSendAttempts int           `mapstructure:"max_send_attempts"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	KeepAlive       time.Duration `mapstructure:"keep_alive"`

	Async              bool          `mapstructure:"async"`
	AsyncQueueCapacity int           `mapstructure:"async_queue_capacity"`
	BatchSize          int           `mapstructure:"batch_size"`
	BatchTimeout       time.Duration `mapstructure:"batch_timeout"`
}

func NewKafkaClient(
	cfg KafkaClientConfig,
	registry stat.Registry,
	logger log.Logger,
) (Client, error) {
	logger = logger.WithField("lib", "dwh")

	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:         cfg.Brokers,
		Topic:           cfg.Topic,
		MaxAttempts:     cfg.MaxSendAttempts,
		QueueCapacity:   cfg.AsyncQueueCapacity,
		BatchSize:       cfg.BatchSize,
		BatchTimeout:    cfg.BatchTimeout,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		IdleConnTimeout: cfg.KeepAlive,
		Async:           cfg.Async,
		Logger:          logger,
		ErrorLogger:     logger,
	})

	hostname, err := os.Hostname()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get hostname")
	}

	c := kafkaClient{
		Writer:   w,
		Hostname: hostname,
	}
	stat.NewRegistrar(registry.ForSubsystem("kafka")).MustRegister(&c.Stat)

	return c, nil
}

func (k kafkaClient) SendMessage(ctx context.Context, msg Message) (err error) {
	tm := k.Stat.SendDuration.Timer(ctx).Start()
	defer func() {
		tm.WithLabels(stat.Labels{"status": stat.TypedErrorLabel(ctx, err)}).Stop()
	}()

	rawMsg, err := msg.MarshalDWH()
	if err != nil {
		return errors.Transform(err, ErrMarshalMessage)
	}

	err = k.Writer.WriteMessages(ctx, kafka.Message{
		Value: rawMsg,
		Time:  time.Now(),
	})
	if err != nil {
		return errors.Transform(err, ErrSendDWH)
	}

	return nil
}

func (k kafkaClient) SendCompactMessage(ctx context.Context, msg CompactMessage) (err error) {
	tm := k.Stat.SendCompactDuration.Timer(ctx).Start()
	defer func() {
		tm.WithLabels(stat.Labels{"status": stat.TypedErrorLabel(ctx, err)}).Stop()
	}()

	dwhMsg, err := MarshalCompactMessage(msg, k.Hostname)
	if err != nil {
		return errors.Transform(err, ErrMarshalMessage)
	}

	err = k.Writer.WriteMessages(ctx, kafka.Message{
		Value: []byte(dwhMsg),
		Time:  time.Now(),
	})
	if err != nil {
		return errors.Transform(err, ErrSendDWH)
	}

	return nil
}
