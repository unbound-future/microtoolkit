package retry

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/AnimateAIPlatform/animate-ai/common/client"
	"github.com/AnimateAIPlatform/animate-ai/common/consts"
	"github.com/AnimateAIPlatform/animate-ai/common/model"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol"
)

func Retry() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		headers := make(map[string]string)
		ctx = context.WithValue(ctx, consts.UpstreamHeadersKey, headers)
		ctx = context.WithValue(ctx, consts.UpstreamResponseKey, model.Error(0, "", "", ""))
		c.Next(ctx)
		errorResponse, ok := ctx.Value(consts.UpstreamResponseKey).(*model.ErrorResponse)
		if !ok {
			hlog.CtxErrorf(ctx, "Failed to get upstream response from context")
			return
		}
		if errorResponse.Code == 0 || errorResponse.Code == 200 {
			// 无需重试
			return
		}
		retryInfo, err := ParseRetryInfo(map[string]string{
			"code": strconv.Itoa(errorResponse.Code),
			"msg":  errorResponse.Msg,
		})
		// 没有匹配到重试规则，不重试
		if err != nil {
			hlog.CtxErrorf(ctx, "No matching retry rule found: %v, do not retry", err)
			retryInfo = map[string]string{
				consts.MaxRetry:      "0",
				consts.RetryInterval: "0",
			}
		}
		retryTimesStr := c.Request.Header.Get(string(consts.ServerRetryCountKey))
		if retryTimesStr == "" {
			retryTimesStr = "0"
		}
		retryTimes, err := strconv.Atoi(retryTimesStr)
		if err != nil {
			hlog.CtxErrorf(ctx, "Failed to parse retry times: %v", err)
			return
		}
		maxRetry, err := strconv.Atoi(retryInfo[consts.MaxRetry])
		if err != nil {
			hlog.CtxErrorf(ctx, "Failed to parse max retry: %v", err)
			return
		}
		retryInterval, err := strconv.Atoi(retryInfo[consts.RetryInterval])
		if err != nil {
			hlog.CtxErrorf(ctx, "Failed to parse retry interval: %v", err)
			return
		}
		// 判断是否需要重试
		if retryTimes < maxRetry {
			hlog.CtxInfof(ctx, "Retrying request, attempt %d", retryTimes+1)
			// 等待指定的时间间隔
			time.Sleep(time.Duration(retryInterval) * time.Second)
			// 重新发起请求
			c.Request.Header.Set(string(consts.ServerRetryCountKey), strconv.Itoa(retryTimes+1))
			c.Request.Header.Set(string(consts.ServerTraceIDKey), errorResponse.ErrorInfo.LogID)
			res := protocol.AcquireResponse()
			err = client.GetClient().Do(ctx, &c.Request, res)
			if err != nil {
				hlog.CtxErrorf(ctx, "Error retrying request: %v", err)
				return
			}
			if res.StatusCode() == 200 {
				// 设置响应头
				res.Header.VisitAll(func(key, value []byte) {
					c.Response.Header.Add(string(key), string(value))
				})
				// 设置状态码
				c.Response.SetStatusCode(res.StatusCode())

				// 流式转发响应体，Hertz 会自动边读边写
				c.SetBodyStream(res.BodyStream(), res.Header.ContentLength())
			} else {
				resBody := model.Error(0, "", "", "")
				if err := json.Unmarshal(res.Body(), resBody); err != nil {
					hlog.CtxErrorf(ctx, "Failed to unmarshal error response: %v", err)
					return
				}
				newErrorResponse := model.Error(res.StatusCode(), resBody.Msg, errorResponse.ErrorInfo.LogID, "")
				c.JSON(newErrorResponse.Code, newErrorResponse)
			}
		} else {
			for k, v := range headers {
				c.Response.Header.Add(k, v)
			}
			c.JSON(errorResponse.Code, errorResponse)
			hlog.CtxInfof(ctx, "Max retry attempts reached (%s), not retrying", retryInfo["max-retry"])
		}
	}
}
