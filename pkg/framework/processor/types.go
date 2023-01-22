package processor

import (
	"context"

	"github.com/antonpriyma/otus-highload/pkg/utils"
)

var (
	ErrTaskNotFound      = utils.NewTypedError("processor_task_not_found", "task not found")
	ErrEndOfTasks        = utils.NewTypedError("processor_end_of_tasks", "end of tasks")
	ErrDeleteTask        = utils.NewTypedError("processor_delete_task", "delete task")
	ErrRetryTask         = utils.NewTypedError("processor_retry_task", "retry task")
	ErrTaskUnrecoverable = utils.NewTypedError("processor_task_unrecoverable", "task unrecoverable")
)

type Task interface {
	Process(ctx context.Context) error
	Ack(ctx context.Context) error
	Delete(ctx context.Context) error
	Defer(ctx context.Context)

	Key() string
	Type() string
}

type TaskWithUser interface {
	Task
	UserID() string
}

type TaskGetter interface {
	Get(ctx context.Context) (Task, error)
}

type (
	ProcessFunc    func(ctx context.Context) error
	MiddlewareFunc func(next ProcessFunc) ProcessFunc
	ErrorHandler   func(ctx context.Context, err error)
)

type MiddlewareList []MiddlewareFunc

func (ml MiddlewareList) Prepend(middlewares ...MiddlewareFunc) MiddlewareList {
	ret := make([]MiddlewareFunc, 0, len(ml)+len(middlewares))

	ret = append(ret, middlewares...)
	ret = append(ret, ml...)

	return ret
}

func (ml MiddlewareList) Apply(f ProcessFunc) ProcessFunc {
	for i := len(ml) - 1; i >= 0; i-- {
		middleware := ml[i]
		f = middleware(f)
	}

	return f
}
