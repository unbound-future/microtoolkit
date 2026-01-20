package logger

import (
	"context"
	"strings"
	"time"

	"github.com/AnimateAIPlatform/animate-ai/common/consts"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/segmentio/ksuid"
)

func AccessLog() app.HandlerFunc {
	var traceID string
	return func(ctx context.Context, c *app.RequestContext) {
		if c.Request.Header.Get(string(consts.ServerTraceIDKey)) != "" {
			traceID = c.Request.Header.Get(string(consts.ServerTraceIDKey))
		} else if c.Request.Header.Get(string(consts.ClientTraceIDKey)) != "" {
			traceID = c.Request.Header.Get(string(consts.ClientTraceIDKey))
		} else {
			traceID = ksuid.New().String()
		}
		c.Response.Header.Set(string(consts.ServerTraceIDKey), traceID)
		ctx = context.WithValue(ctx, consts.ServerTraceIDKey, traceID)
		start := time.Now()

		// 获取所有请求头信息
		var requestHeaders []string
		c.Request.Header.VisitAll(func(key, value []byte) {
			requestHeaders = append(requestHeaders, string(key)+":"+string(value))
		})
		allRequestHeaders := strings.Join(requestHeaders, "; ")

		// 获取查询参数
		var queryParams []string
		c.Request.URI().QueryArgs().VisitAll(func(key, value []byte) {
			queryParams = append(queryParams, string(key)+"="+string(value))
		})
		allQueryParams := strings.Join(queryParams, "&")

		// 获取POST表单参数
		var formParams []string
		c.Request.PostArgs().VisitAll(func(key, value []byte) {
			formParams = append(formParams, string(key)+"="+string(value))
		})
		allFormParams := strings.Join(formParams, "&")

		// 获取所有参数（包括multipart表单）
		var allParams []string
		// 合并查询参数和表单参数
		allParams = append(allParams, queryParams...)
		allParams = append(allParams, formParams...)
		allRequestParams := strings.Join(allParams, "&")

		// 打印请求信息
		hlog.CtxInfof(ctx, "request_url=%s", c.Request.URI().PathOriginal())
		hlog.CtxInfof(ctx, "request_method=%s", c.Request.Header.Method())
		hlog.CtxInfof(ctx, "request_headers=%s", allRequestHeaders)
		hlog.CtxInfof(ctx, "request_query_params=%s", allQueryParams)
		hlog.CtxInfof(ctx, "request_form_params=%s", allFormParams)
		hlog.CtxInfof(ctx, "request_all_params=%s", allRequestParams)
		hlog.CtxInfof(ctx, "request_body=%s", string(c.Request.Body()))
		hlog.CtxInfof(ctx, "request_content_type=%s", string(c.Request.Header.ContentType()))
		hlog.CtxInfof(ctx, "request_user_agent=%s", string(c.Request.Header.UserAgent()))
		hlog.CtxInfof(ctx, "request_referer=%s", string(c.Request.Header.Peek("Referer")))
		hlog.CtxInfof(ctx, "request_remote_addr=%s", c.ClientIP())
		hlog.CtxInfof(ctx, "request_uri=%s", string(c.Request.URI().FullURI()))
		c.Next(ctx)
		end := time.Now()
		latency := end.Sub(start).Microseconds()
		// 获取所有响应头信息
		var responseHeaders []string
		c.Response.Header.VisitAll(func(key, value []byte) {
			responseHeaders = append(responseHeaders, string(key)+":"+string(value))
		})
		allHeaders := strings.Join(responseHeaders, "; ")

		// 打印响应信息
		hlog.CtxInfof(ctx, "response_status=%d", c.Response.StatusCode())
		hlog.CtxInfof(ctx, "response_cost=%d", latency)
		hlog.CtxInfof(ctx, "response_method=%s", c.Request.Header.Method())
		hlog.CtxInfof(ctx, "response_full_path=%s", c.Request.URI().PathOriginal())
		hlog.CtxInfof(ctx, "response_client_ip=%s", c.ClientIP())
		hlog.CtxInfof(ctx, "response_host=%s", c.Request.Host())
		hlog.CtxInfof(ctx, "response_headers=%s", allHeaders)
		hlog.CtxInfof(ctx, "response_content_type=%s", string(c.Response.Header.ContentType()))
		hlog.CtxInfof(ctx, "response_content_length=%d", c.Response.Header.ContentLength())

		// 打印响应body
		responseBody := string(c.Response.Body())
		hlog.CtxInfof(ctx, "response_body=%s", responseBody)

		// 打印trace信息
		hlog.CtxInfof(ctx, "trace_id=%s", traceID)
	}
}
