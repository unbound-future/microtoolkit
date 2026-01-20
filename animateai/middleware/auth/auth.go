package auth

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/AnimateAIPlatform/animate-ai/common/consts"
	"github.com/AnimateAIPlatform/animate-ai/internal/gateway/service"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	hzconsts "github.com/cloudwego/hertz/pkg/protocol/consts"
)

// Auth 鉴权中间件
// 从请求头获取鉴权信息（Basic Auth），验证用户和密码，并将用户信息存储到 context
func Auth() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		// 获取 Authorization header
		authHeader := string(c.Request.Header.Peek("Authorization"))
		if authHeader == "" {
			hlog.CtxErrorf(ctx, "Missing Authorization header")
			c.JSON(hzconsts.StatusUnauthorized, map[string]interface{}{
				"error": "Unauthorized: Missing Authorization header",
			})
			c.Abort()
			return
		}

		// 解析 Basic Auth
		// 格式: "Basic base64(username:password)"
		if !strings.HasPrefix(authHeader, "Basic ") {
			hlog.CtxErrorf(ctx, "Invalid Authorization header format, expected Basic auth")
			c.JSON(hzconsts.StatusUnauthorized, map[string]interface{}{
				"error": "Unauthorized: Invalid Authorization header format",
			})
			c.Abort()
			return
		}

		// 解码 base64
		encoded := strings.TrimPrefix(authHeader, "Basic ")
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			hlog.CtxErrorf(ctx, "Failed to decode Authorization header: %v", err)
			c.JSON(hzconsts.StatusUnauthorized, map[string]interface{}{
				"error": "Unauthorized: Invalid Authorization header",
			})
			c.Abort()
			return
		}

		// 解析用户名和密码
		// 格式: "username:password"
		credentials := string(decoded)
		parts := strings.SplitN(credentials, ":", 2)
		if len(parts) != 2 {
			hlog.CtxErrorf(ctx, "Invalid credentials format")
			c.JSON(hzconsts.StatusUnauthorized, map[string]interface{}{
				"error": "Unauthorized: Invalid credentials format",
			})
			c.Abort()
			return
		}

		userName := parts[0]
		password := parts[1]

		hlog.CtxInfof(ctx, "Auth middleware: authenticating user: %s", userName)

		// 验证用户和密码
		userService := service.NewUserService()
		user, err := userService.Login(ctx, userName, password)
		if err != nil {
			hlog.CtxErrorf(ctx, "Authentication failed for user %s: %v", userName, err)
			c.JSON(hzconsts.StatusUnauthorized, map[string]interface{}{
				"error": "Unauthorized: Invalid username or password",
			})
			c.Abort()
			return
		}

		// 验证成功，将用户信息存储到 context
		ctx = context.WithValue(ctx, consts.UserIDKey, user.ID)
		ctx = context.WithValue(ctx, consts.UserAccountIDKey, user.AccountID)
		ctx = context.WithValue(ctx, consts.UserNameKey, user.UserName)

		hlog.CtxInfof(ctx, "Authentication successful: userID=%d, accountID=%s, userName=%s",
			user.ID, user.AccountID, user.UserName)

		// 继续处理请求
		c.Next(ctx)
	}
}
