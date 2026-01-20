package util

import (
	"context"

	"github.com/AnimateAIPlatform/animate-ai/common/consts"
)

func GetTraceID(ctx context.Context) string {
	v := ctx.Value(consts.ServerTraceIDKey)
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
