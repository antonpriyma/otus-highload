package cmnlabelsstat

import (
	"context"

	"github.com/antonpriyma/otus-highload/pkg/stat"
)

type contextKey struct{}

func AddCommonLabels(ctx context.Context, labels stat.Labels) context.Context {
	return context.WithValue(ctx, contextKey{}, stat.MergeLabels(getCommonLabels(ctx), labels))
}

func CopyCommonLabels(target, source context.Context) context.Context {
	return AddCommonLabels(target, getCommonLabels(source))
}

func getCommonLabels(ctx context.Context) stat.Labels {
	if val := ctx.Value(contextKey{}); val != nil {
		return val.(stat.Labels)
	}

	return stat.Labels{}
}
