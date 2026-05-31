package server

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

// CORS 中间件 — 处理跨域请求
func corsMiddleware() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			if tr, ok := transport.FromServerContext(ctx); ok {
				tr.ReplyHeader().Set("Access-Control-Allow-Origin", "*")
				tr.ReplyHeader().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				tr.ReplyHeader().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				tr.ReplyHeader().Set("Access-Control-Max-Age", "86400")
			}
			return handler(ctx, req)
		}
	}
}
