package batchgetter

import (
	"context"
	"sync"

	"github.com/antonpriyma/otus-highload/pkg/errors"
	"github.com/antonpriyma/otus-highload/pkg/framework/processor"
)

type BatchTaskGetter interface {
	GetBatch(ctx context.Context) ([]processor.Task, error)
}

func New(batched BatchTaskGetter) processor.TaskGetter {
	return &batchedWrapForTaskGetter{
		wg:          &sync.WaitGroup{},
		BatchGetter: batched,
	}
}

type batchedWrapForTaskGetter struct {
	BatchGetter  BatchTaskGetter
	batchedTasks []processor.Task
	wg           *sync.WaitGroup
}

func (b *batchedWrapForTaskGetter) Get(ctx context.Context) (processor.Task, error) {
	if len(b.batchedTasks) != 0 {
		return b.popTask(), nil
	}

	b.wg.Wait()
	tasks, err := b.BatchGetter.GetBatch(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to select tasks")
	}
	if len(tasks) == 0 {
		return nil, errors.Transform(
			errors.Errorf("there are no tasks for process"),
			processor.ErrTaskNotFound,
		)
	}

	b.setTasks(tasks)
	return b.popTask(), nil
}

func (b *batchedWrapForTaskGetter) popTask() processor.Task {
	currentTask := b.batchedTasks[0]
	b.setTasks(b.batchedTasks[1:])

	b.wg.Add(1)
	return &wrappedTask{
		Task: currentTask,
		wg:   b.wg,
	}
}

func (b *batchedWrapForTaskGetter) setTasks(tasks []processor.Task) {
	b.batchedTasks = tasks
}

type wrappedTask struct {
	processor.Task
	wg *sync.WaitGroup
}

func (w *wrappedTask) Defer(ctx context.Context) {
	defer w.wg.Done()
	w.Task.Defer(ctx)
}

func (w *wrappedTask) UserID() string {
	taskWithUser, ok := w.Task.(processor.TaskWithUser)
	if !ok {
		return ""
	}

	return taskWithUser.UserID()
}
