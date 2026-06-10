package data

import (
	"blog/internal/biz"
	"context"
	"errors"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Price struct {
	gorm.Model
	ProductName string          `gorm:"comment:'商品名称'"`
	Weight      decimal.Decimal `gorm:"type:decimal(10,2);comment:'重量'"`
	UnitPrice   decimal.Decimal `gorm:"type:decimal(10,2);comment:'单价'"`
	PriceDate   string          `gorm:"comment:'日期'"`
	UserID      string          `gorm:"comment:'用户ID';index"`
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

func (r *priceRepo) Save(ctx context.Context, p *biz.Price) (uint, error) {
	price := Price{
		ProductName: p.ProductName,
		Weight:      p.Weight,
		UnitPrice:   p.UnitPrice,
		PriceDate:   p.PriceDate,
		UserID:      p.UserId,
	}
	result := r.data.db.WithContext(ctx).Create(&price)
	if result.Error != nil {
		return 0, result.Error
	}
	return price.ID, nil
}

func (r *priceRepo) Update(ctx context.Context, p *biz.Price) error {
	updates := map[string]any{
		"product_name": p.ProductName,
		"weight":       p.Weight,
		"unit_price":   p.UnitPrice,
		"price_date":   p.PriceDate,
	}
	tx := r.data.db.WithContext(ctx).
		Model(&Price{}).
		Where("id = ? AND user_id = ?", p.ID, p.UserId).
		Updates(updates)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("no updatable price record found for current user")
	}
	return nil
}

func (r *priceRepo) FindByUserIdAndID(ctx context.Context, userId string, id uint) (*biz.Price, error) {
	var price Price
	tx := r.data.db.WithContext(ctx).Where("user_id = ? AND id = ?", userId, id).First(&price)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return mapPriceToBiz(&price), nil
}

func (r *priceRepo) FindByPage(ctx context.Context, userId string, req *biz.PricePageRequest) ([]*biz.Price, int64, error) {
	db := r.data.db.WithContext(ctx).Model(&Price{}).Where("user_id = ?", userId)
	var total int64
	err := db.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	page := req.Current
	size := req.Size
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	var prices []Price
	err = db.Order("created_at DESC").
		Offset(int((page - 1) * size)).
		Limit(int(size)).
		Find(&prices).Error
	if err != nil {
		return nil, 0, err
	}

	list := make([]*biz.Price, 0, len(prices))
	for _, p := range prices {
		list = append(list, mapPriceToBiz(&p))
	}
	return list, total, nil
}

func (r *priceRepo) DeleteByUserIdAndID(ctx context.Context, userId string, id uint) error {
	tx := r.data.db.WithContext(ctx).Where("user_id = ? AND id = ?", userId, id).Delete(&Price{})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("no deletable price record found or permission denied")
	}
	return nil
}

func mapPriceToBiz(price *Price) *biz.Price {
	if price == nil {
		return nil
	}
	return &biz.Price{
		ID:          price.ID,
		ProductName: price.ProductName,
		Weight:      price.Weight,
		UnitPrice:   price.UnitPrice,
		PriceDate:   price.PriceDate,
		UserId:      price.UserID,
	}
}
