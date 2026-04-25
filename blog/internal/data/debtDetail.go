package data

import (
	"blog/internal/biz"
	"context"
	"errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"strconv"
	"time"
)

type DebtDetail struct {
	gorm.Model
	DebtId      uint            `gorm:"comment:debt id"`
	PostingDate time.Time       `gorm:"comment:posting date"`
	Principal   decimal.Decimal `gorm:"type:decimal(10,2);comment:principal"`
	Interest    decimal.Decimal `gorm:"type:decimal(10,2);comment:interest"`
	Period      uint            `gorm:"size:8;comment:period"`
}

type DebtDetailRepo struct {
	data *Data
	log  *log.Helper
}

func (d *DebtDetailRepo) SaveDb(ctx context.Context, b *biz.DebtDetail) (string, error) {
	var debtDetail DebtDetail
	debtDetail.DebtId = b.DebtId
	postingDate, err := time.ParseInLocation("2006-01-02 15:04:05", b.PostingDate, time.Local)
	if err != nil {
		return "", err
	}
	debtDetail.PostingDate = postingDate
	debtDetail.Principal = b.Principal
	debtDetail.Interest = b.Interest
	debtDetail.Period = b.Period
	tx := d.data.db.WithContext(ctx).Create(&debtDetail)
	if tx.Error != nil {
		return "", tx.Error
	}
	id := strconv.FormatUint(uint64(debtDetail.ID), 10)
	return id, nil
}

func (d *DebtDetailRepo) GetByUserIdAndID(ctx context.Context, userId string, id uint) (*biz.DebtDetail, error) {
	var detail DebtDetail
	err := d.data.db.WithContext(ctx).
		Table("debt_details as dd").
		Select("dd.*").
		Joins("JOIN debts d ON d.id = dd.debt_id").
		Where("dd.id = ? AND d.user_id = ? AND dd.deleted_at IS NULL AND d.deleted_at IS NULL", id, userId).
		Take(&detail).Error
	if err != nil {
		return nil, err
	}
	return mapDebtDetailToBiz(&detail), nil
}

func (d *DebtDetailRepo) ListByUserIdAndDebtId(ctx context.Context, userId string, debtId uint) ([]*biz.DebtDetail, error) {
	var details []DebtDetail
	err := d.data.db.WithContext(ctx).
		Table("debt_details as dd").
		Select("dd.*").
		Joins("JOIN debts d ON d.id = dd.debt_id").
		Where("d.user_id = ? AND dd.debt_id = ? AND dd.deleted_at IS NULL AND d.deleted_at IS NULL", userId, debtId).
		Order("dd.created_at DESC").
		Find(&details).Error
	if err != nil {
		return nil, err
	}

	items := make([]*biz.DebtDetail, 0, len(details))
	for i := range details {
		items = append(items, mapDebtDetailToBiz(&details[i]))
	}
	return items, nil
}

func (d *DebtDetailRepo) EditDb(ctx context.Context, detail *biz.DebtDetail) error {
	postingDate, err := time.ParseInLocation("2006-01-02 15:04:05", detail.PostingDate, time.Local)
	if err != nil {
		return err
	}
	var detailDb DebtDetail
	detailDb.ID = detail.Id
	detailDb.DebtId = detail.DebtId
	detailDb.PostingDate = postingDate
	detailDb.Principal = detail.Principal
	detailDb.Interest = detail.Interest
	detailDb.Period = detail.Period

	tx := d.data.db.WithContext(ctx).Model(&DebtDetail{}).Where("id = ?", detail.Id).Updates(&detailDb)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("no updatable debt detail found")
	}
	return nil
}

func (d *DebtDetailRepo) DeleteDb(ctx context.Context, userId string, id uint) error {
	// verify ownership before delete
	_, err := d.GetByUserIdAndID(ctx, userId, id)
	if err != nil {
		return errors.New("no deletable debt detail found or permission denied")
	}
	tx := d.data.db.WithContext(ctx).Delete(&DebtDetail{}, id)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("no deletable debt detail found or permission denied")
	}
	return nil
}

func NewDebtDetailRepo(data *Data, logger log.Logger) biz.DebtDetailRepo {
	return &DebtDetailRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func mapDebtDetailToBiz(detail *DebtDetail) *biz.DebtDetail {
	if detail == nil {
		return nil
	}
	return &biz.DebtDetail{
		Id:          detail.ID,
		DebtId:      detail.DebtId,
		PostingDate: detail.PostingDate.Format("2006-01-02 15:04:05"),
		Principal:   detail.Principal,
		Interest:    detail.Interest,
		Period:      detail.Period,
	}
}
