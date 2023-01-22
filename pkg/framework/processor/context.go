package processor

import "context"

type taskKey struct{}

func SetTask(ctx context.Context, task Task) context.Context {
	return context.WithValue(ctx, taskKey{}, task)
}

func GetTask(ctx context.Context) Task {
	task, _ := ctx.Value(taskKey{}).(Task)
	return task
}
