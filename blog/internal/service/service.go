package service

import "github.com/google/wire"

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewGreeterService, NewPostService, NewUserService, NewAiocrService, NewPriceService, NewDebtService, NewDebtDetailService)
