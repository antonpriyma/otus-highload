package processorpool

import (
	"context"
	"time"

	"github.com/antonpriyma/otus-highload/pkg/debug"
	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/framework/processor"
	"github.com/antonpriyma/otus-highload/pkg/log"
	"github.com/antonpriyma/otus-highload/pkg/stat"
	"github.com/antonpriyma/otus-highload/pkg/stat/loggerstat"
)

type job struct {
	ctx  context.Context
	task processor.Task

	TaskTimeout  time.Duration
	Logger       log.Logger
	Stat         stats
	Middlewares  processor.MiddlewareList
	ErrorHandler processor.ErrorHandler
}

func (p *pool) newJob(ctx context.Context, task processor.Task) *job {
	j := job{
		ctx:  makeThreadContext(ctx, "worker"),
		task: task,

		TaskTimeout:  p.Config.TaskTimeout,
		Logger:       p.Logger.WithFields(taskLogFields(task)),
		Stat:         p.Stat,
		Middlewares:  p.App.Middlewares,
		ErrorHandler: p.App.ErrorHandler,
	}

	j.Middlewares = j.Middlewares.Prepend(j.basicMiddleware, j.errorHandlerMiddleware, j.statMiddleware)

	return &j
}

func (j job) basicMiddleware(next processor.ProcessFunc) processor.ProcessFunc {
	return func(ctx context.Context) error {
		ctx = processor.SetTask(ctx, j.task)
		ctx = log.AddCtxFields(ctx, taskLogFields(j.task))
		ctx = loggerstat.InitStatForCtx(ctx)

		j.Logger.ForCtx(ctx).Debug("job started")
		defer j.Logger.ForCtx(ctx).Debug("job finished")

		return next(ctx)
	}
}

func (j job) statMiddleware(next processor.ProcessFunc) processor.ProcessFunc {
	return func(ctx context.Context) (err error) {
		typ := j.task.Type()

		timer := j.Stat.JobDuration.Timer(ctx).Start()
		defer func() {
			timer.WithLabels(stat.Labels{"type": typ}).Stop()
			j.Stat.JobStatus.Counter(ctx).WithLabels(stat.Labels{
				"status": stat.TypedErrorLabel(ctx, err),
				"type":   typ,
			}).Add(1)

			loggerstat.PrintStat(ctx, j.Logger)
		}()

		return next(ctx)
	}
}

func (j job) errorHandlerMiddleware(next processor.ProcessFunc) processor.ProcessFunc {
	return func(ctx context.Context) error {
		err := next(ctx)
		if err != nil {
			j.ErrorHandler(ctx, err)
		}

		return nil
	}
}

func (j *job) Work() {
	ctx := j.ctx

	defer func() {
		r := errors.RecoverError(recover())
		if r == nil {
			return
		}

		// dumb panic check
		j.Logger.ForCtx(ctx).
			WithError(r).
			WithField("stack", debug.ErrorStackTrace(r)).
			Error("job unexpected panic catched")
	}()

	err := j.Middlewares.Apply(j.processTask)(ctx)
	if err != nil {
		j.Logger.ForCtx(ctx).WithError(err).Error("error unhandled by error handling middleware")
		return
	}
}

func (j *job) processTask(ctx context.Context) (err error) {
	limitedCtx := ctx
	if j.TaskTimeout > 0 {
		var cancel context.CancelFunc
		limitedCtx, cancel = context.WithTimeout(ctx, j.TaskTimeout)
		defer cancel()
	}

	defer j.task.Defer(ctx)

	err = j.task.Process(limitedCtx)
	if errors.Is(err, processor.ErrDeleteTask, processor.ErrTaskUnrecoverable) {
		errDelete := j.task.Delete(ctx)
		if errDelete != nil {
			return errors.Wrapf(err, "deleting task failed (%s)", errDelete)
		}

		return errors.Wrap(err, "deleting task")
	}
	if err != nil {
		return errors.Wrap(err, "failed to process Task")
	}

	err = j.task.Ack(ctx)
	return errors.Wrap(err, "failed to acknowledge task")
}

func taskLogFields(task processor.Task) log.Fields {
	return log.Fields{
		"type": task.Type(),
		"key":  task.Key(),
	}
}
