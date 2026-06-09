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
	UserID    string `gorm:"comment:'用户ID'"`
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
	var price Price
	result := r.data.db.First(&price, p.ID)
	if result.Error != nil {
		return nil, result.Error
	}
	price.Name = p.Name
	price.Price = p.Price
	price.PriceDate = p.PriceDate
	if err := r.data.db.Save(&price).Error; err != nil {
		return nil, err
	}
	return &biz.Price{
		ID:        price.ID,
		Name:      price.Name,
		Price:     price.Price,
		PriceDate: price.PriceDate,
	}, nil
}
func (r *priceRepo) FindByID(ctx context.Context, id int64) (*biz.Price, error) {
	var price Price
	result := r.data.db.First(&price, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &biz.Price{
		ID:        price.ID,
		Name:      price.Name,
		Price:     price.Price,
		PriceDate: price.PriceDate,
	}, nil
}
func (r *priceRepo) Delete(ctx context.Context, id int64) error {
	result := r.data.db.Delete(&Price{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *priceRepo) ListByHello(ctx context.Context, str string) ([]*biz.Price, error) {
	return make([]*biz.Price, 0), nil
}
func (r *priceRepo) ListAll(ctx context.Context) ([]*biz.Price, error) {
	var prices []Price
	result := r.data.db.Order("created_at DESC").Find(&prices)
	if result.Error != nil {
		return nil, result.Error
	}
	list := make([]*biz.Price, 0, len(prices))
	for _, p := range prices {
		list = append(list, &biz.Price{
			ID:        p.ID,
			Name:      p.Name,
			Price:     p.Price,
			PriceDate: p.PriceDate,
		})
	}
	return list, nil
}
