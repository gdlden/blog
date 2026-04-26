package server

import (
	debtV1 "blog/api/debt/v1"
	v1 "blog/api/helloworld/v1"
	ocrv1 "blog/api/ocr/v1"
	postV1 "blog/api/post/v1"
	priceV1 "blog/api/price/v1"
	userV1 "blog/api/user/v1"
	"blog/internal/conf"
	"blog/internal/service"
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/http"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, greeter *service.GreeterService,
	postService *service.PostService,
	userService *service.UserService,
	ocrService *service.AiocrService,
	priceService *service.PriceService,
	debtService *service.DebtService,
	detailService *service.DebtDetailService,
	logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			tracing.Server(),
			logging.Server(logger),
			selector.Server(
				jwt.Server(func(token *jwtv5.Token) (interface{}, error) {
					return []byte("dfsdsjikldsfkdfjdkls"), nil
				}),
			).
				// Path("/post/add/v1").
				Match(NewWhiteListMatcher()).Build(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	v1.RegisterGreeterHTTPServer(srv, greeter)
	postV1.RegisterPostHTTPServer(srv, postService)
	userV1.RegisterUserHTTPServer(srv, userService)
	ocrv1.RegisterAiocrHTTPServer(srv, ocrService)
	priceV1.RegisterPriceHTTPServer(srv, priceService)
	debtV1.RegisterDebtHTTPServer(srv, debtService)
	debtV1.RegisterDebtDetailHTTPServer(srv, detailService)
	srv.Route("/").POST("/debtDetail/ocr/v1", detailService.RecognizeDebtDetailOCRHTTP)
	return srv
}
func NewWhiteListMatcher() selector.MatchFunc {

	whiteList := make(map[string]struct{})
	whiteList["/post.v1.Post/CreatePost"] = struct{}{}
	whiteList["/post.v1.Post/GetPostPage"] = struct{}{}
	whiteList["/user.v1.User/CreateUser"] = struct{}{}
	whiteList["/shop.interface.v1.ShopInterface/Register"] = struct{}{}
	whiteList["/user.v1.User/UserLogin"] = struct{}{}
	whiteList["/ocr.v1.Aiocr/Ocr"] = struct{}{}
	whiteList["/api.price.Price/CreatePrice"] = struct{}{}
	return func(ctx context.Context, operation string) bool {
		if _, ok := whiteList[operation]; ok {
			return false
		}
		return true
	}
}
