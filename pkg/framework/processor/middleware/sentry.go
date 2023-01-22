package middleware

import (
	"context"

	"github.com/antonpriyma/otus-highload/pkg/clients/sentry"
	"github.com/antonpriyma/otus-highload/pkg/context/reqid"
	"github.com/antonpriyma/otus-highload/pkg/framework/processor"
)

func NewSentryTaskException(processorName string) func(context.Context) sentry.Exception {
	return func(ctx context.Context) sentry.Exception {
		task := processor.GetTask(ctx)
		if task == nil {
			return sentry.Exception{
				RequestID: reqid.GetRequestID(ctx),
				Path:      processorName + "unknown",
				User: sentry.User{
					ID: "unknown",
				},
				CustomTags: map[string]string{
					"type": "unknown",
				},
			}
		}

		userID := ""
		taskWithUser, ok := task.(processor.TaskWithUser)
		if ok {
			userID = taskWithUser.UserID()
		}

		return sentry.Exception{
			RequestID: reqid.GetRequestID(ctx),
			Path:      processorName + "_task",
			User: sentry.User{
				ID: userID,
			},
			CustomTags: map[string]string{
				"type": task.Type(),
				"key":  task.Key(),
			},
		}
	}
}
