package processorpool

import (
	"context"
	"time"

	"github.com/antonpriyma/otus-highload/pkg/context/reqid"
	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/framework/processor"
	"github.com/antonpriyma/otus-highload/pkg/framework/processor/middleware"
	"github.com/antonpriyma/otus-highload/pkg/log"
	"github.com/antonpriyma/otus-highload/pkg/stat"
	"github.com/antonpriyma/otus-highload/pkg/stat/loggerstat"
	"github.com/antonpriyma/otus-highload/pkg/utils"

	"github.com/ivpusic/grpool"
)

type Pool interface {
	Run(ctx context.Context)
	Graceful(ctx context.Context) error
	Stop(ctx context.Context) error
}

type AppConfig struct {
	Task         processor.TaskGetter
	Middlewares  processor.MiddlewareList
	ErrorHandler processor.ErrorHandler
}

type PoolConfig struct {
	MaxWorkers         int           `mapstructure:"max_workers"`
	QueueLimit         int           `mapstructure:"queue_limit"`
	SleepOnNoTask      time.Duration `mapstructure:"sleep_on_no_task"`
	SleepOnTaskGetFail time.Duration `mapstructure:"sleep_on_task_get_fail"`
	TaskTimeout        time.Duration `mapstructure:"task_timeout"`
	PrintIterationStat bool          `mapstructure:"print_iteration_stat"`
}

func New(
	app AppConfig,
	cfg PoolConfig,
	logger log.Logger,
	registry stat.Registry,
) Pool {
	if app.ErrorHandler == nil {
		app.ErrorHandler = middleware.LogErrorHandler(logger)
	}

	p := pool{
		App:     app,
		Config:  cfg,
		Logger:  logger,
		Workers: grpool.NewPool(cfg.MaxWorkers, cfg.QueueLimit),

		gracefulInitedCh:   make(chan struct{}),
		gracefulFinishedCh: make(chan struct{}),
	}

	stat.NewRegistrar(registry.ForSubsystem("task_processor")).MustRegister(&p.Stat)

	return &p
}

type pool struct {
	App     AppConfig
	Config  PoolConfig
	Logger  log.Logger
	Workers *grpool.Pool
	Stat    stats

	gracefulInitedCh   chan struct{}
	gracefulFinishedCh chan struct{}
}

type stats struct {
	JobDuration stat.TimerCtor   `labels:"method,type"`
	JobStatus   stat.CounterCtor `labels:"method,status,type"`
	GetDuration stat.TimerCtor   `labels:"method,status"`
}

func (p *pool) enqueueTask(ctx context.Context, task processor.Task) {
	j := p.newJob(ctx, task)
	p.Workers.WaitCount(1)

	p.Workers.JobQueue <- func() {
		j.Work()
		p.Workers.JobDone()
	}
}

func (p *pool) processTask(inputCtx context.Context) (err error) {
	ctx := loggerstat.InitDummyForCtx(inputCtx)
	if p.Config.PrintIterationStat {
		ctx = loggerstat.InitStatForCtx(inputCtx)
	}

	timer := p.Stat.GetDuration.Timer(ctx).Start()
	defer func() {
		timer.WithLabels(stat.Labels{"status": stat.TypedErrorLabel(ctx, err)}).Stop()
	}()

	task, err := p.App.Task.Get(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to select task")
	}

	p.enqueueTask(inputCtx, task)

	loggerstat.PrintStat(ctx, p.Logger)

	return nil
}

func (p *pool) Run(ctx context.Context) {
	ctx = makeThreadContext(ctx, "pool_iteration")
	p.Logger.ForCtx(ctx).Info("starting loop")

	for {
		if utils.IsSignalChanClosed(p.gracefulInitedCh) {
			p.Logger.ForCtx(ctx).Info("stopping loop")
			break
		}

		ctx = log.AddCtxFields(ctx, log.Fields{"job_id": reqid.GenerateRequestID()})

		err := p.processTask(ctx)
		if errors.Is(err, processor.ErrEndOfTasks) {
			p.Logger.ForCtx(ctx).WithError(err).Warn("end of tasks, exiting")
			break
		}
		if errors.Is(err, processor.ErrTaskNotFound) {
			p.Logger.ForCtx(ctx).WithError(err).Debugf("trigger skip iteration, sleep for %s", p.Config.SleepOnNoTask)
			utils.SleepOrChannelClose(p.gracefulInitedCh, p.Config.SleepOnNoTask)
			continue
		}
		if err != nil {
			p.Logger.ForCtx(ctx).WithError(err).Error("error during processor iteration")
			utils.SleepOrChannelClose(p.gracefulInitedCh, p.Config.SleepOnTaskGetFail)
		}
	}

	p.Workers.WaitAll()
	close(p.gracefulFinishedCh)
}

func (p *pool) Graceful(ctx context.Context) error {
	close(p.gracefulInitedCh)

	select {
	case <-ctx.Done():
		return errors.New("failed to graceful shutdown")
	case <-p.gracefulFinishedCh:
		return nil
	}
}

func (p *pool) Stop(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return errors.New("failed to shutdown")
	case <-p.gracefulFinishedCh:
		return nil
	}
}

func makeThreadContext(ctx context.Context, typ string) context.Context {
	fields := log.GetCtxFields(ctx)
	fields["thread"] = typ

	return log.SetCtxFields(ctx, fields)
}
