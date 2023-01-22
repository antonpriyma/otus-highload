package async

import (
	"context"
	"time"

	"github.com/antonpriyma/otus-highload/pkg/debug"
	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/log"
	"github.com/antonpriyma/otus-highload/pkg/stat"
	"github.com/antonpriyma/otus-highload/pkg/stat/loggerstat"
)

type Task func()

type Executor interface {
	Do(Task)
}

type Config struct {
	PoolSize        int           `mapstructure:"pool_size"`
	QueueSize       int           `mapstructure:"queue_size"`
	LockWarnTimeout time.Duration `mapstructure:"lock_warn_timeout"`
}

func NewExecutor(cfg Config, logger log.Logger, registry stat.Registry) Executor {
	ex := executor{
		Config: cfg,
		Logger: logger.WithField("layer", "async executor"),

		Queue: make(chan Task, cfg.QueueSize),
		Pool:  make(chan struct{}, cfg.PoolSize),
	}
	stat.NewRegistrar(registry.ForSubsystem("async_executor")).MustRegister(&ex.statSender)

	for i := 0; i < cfg.PoolSize; i++ {
		ex.returnSlot()
	}

	go ex.run()

	return ex
}

type executor struct {
	Config Config
	Logger log.Logger

	Queue chan Task
	Pool  chan struct{}

	statSender
}

type statSender struct {
	AsyncPutTotal stat.CounterCtor `labels:"status"`
}

func (ex executor) getSlot() {
	select {
	case <-ex.Pool:
		return
	default:
	}

	ticker := time.NewTicker(ex.Config.LockWarnTimeout)
	defer ticker.Stop()

	start := time.Now()
	for {
		select {
		case <-ex.Pool:
			return
		case <-ticker.C:
			ex.Logger.Warnf("async pool is locked for %d seconds", time.Since(start).Seconds())
		}
	}
}

func (ex executor) returnSlot() {
	ex.Pool <- struct{}{}
}

func (ex executor) run() {
	for {
		ex.getSlot()
		t := ex.getTask()

		go func() {
			defer func() {
				ex.returnSlot()

				r := errors.RecoverError(recover())
				if r == nil {
					return
				}

				ex.Logger.
					WithError(r).
					WithField("stack", debug.ErrorStackTrace(r)).
					Error("panic during perform a task")
			}()

			t()
		}()
	}
}

func (ex executor) getTask() Task {
	return <-ex.Queue
}

func (ex executor) Do(t Task) {
	ctx := loggerstat.InitStatForCtx(context.Background())

	select {
	case ex.Queue <- t:
		ex.AsyncPutTotal.Counter(ctx).WithLabels(stat.Labels{"async_status": "ok"}).Add(1)
	default:
		ex.Logger.Error("failed to put async task: queue is full")
		ex.AsyncPutTotal.Counter(ctx).WithLabels(stat.Labels{"async_status": "fail"}).Add(1)
	}
}
