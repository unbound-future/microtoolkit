package cors

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
)

// CORS 跨域中间件 - 允许所有跨域请求
func CORS() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		origin := string(c.Request.Header.Peek("Origin"))

		// 允许所有源
		if origin != "" {
			c.Response.Header.Set("Access-Control-Allow-Origin", origin)
			c.Response.Header.Set("Access-Control-Allow-Credentials", "true")
		} else {
			c.Response.Header.Set("Access-Control-Allow-Origin", "*")
		}
		c.Response.Header.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Response.Header.Set("Access-Control-Allow-Headers", strings.Join([]string{
			"Content-Type",
			"Authorization",
			"X-Requested-With",
			"Accept",
			"Origin",
		}, ", "))
		c.Response.Header.Set("Access-Control-Max-Age", "3600")

		// 处理预检请求
		if string(c.Request.Method()) == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next(ctx)
	}
}
