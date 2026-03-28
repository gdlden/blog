package data

import (
	"blog/internal/biz"
	"context"
	"errors"
	"fmt"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	jwtv5 "github.com/golang-jwt/jwt/v5"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type Price struct {
	gorm.Model
	Name      string `gorm:"comment:'商品名称'"`
	Price     string `gorm:"comment:'单价'"`
	PriceDate string `gorm:"comment:'价格日期'"`
	UserID    string `gorm:"comment:'用户ID`
}
type priceRepo struct {
	data *Data
	log  *log.Helper
}

func NewPriceRepo(data *Data, logger log.Logger) biz.PriceRep {
	return &priceRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}
func (r *priceRepo) Save(ctx context.Context, p *biz.Price) uint {
	var price Price
	price.Name = p.Name
	price.Price = p.Price
	price.PriceDate = p.PriceDate

	token, ok := jwt.FromContext(ctx)
	if !ok {
		log.Info("未登录")
	}
	userInfoMap := token.(jwtv5.MapClaims)
	userId := userInfoMap["userId"]

	price.UserID = userId.(string)
	err := r.data.db.Where(&Price{Name: price.Name}).First(&price)
	if !errors.Is(err.Error, gorm.ErrRecordNotFound) {
		return 0
	}

	error := r.data.db.Create(&price)
	if error != nil {
		fmt.Println(error)
	}
	return price.ID
}
func (r *priceRepo) Update(ctx context.Context, p *biz.Price) (*biz.Price, error) {
	return p, nil
}
func (r *priceRepo) FindByID(ctx context.Context, id int64) (*biz.Price, error) {
	return &biz.Price{}, nil
}
func (r *priceRepo) ListByHello(ctx context.Context, str string) ([]*biz.Price, error) {
	return make([]*biz.Price, 0), nil
}
func (r *priceRepo) ListAll(ctx context.Context) ([]*biz.Price, error) {
	return make([]*biz.Price, 0), nil
}
