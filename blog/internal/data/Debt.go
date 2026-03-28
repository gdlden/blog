package data

import (
	"blog/internal/biz"
	"context"
	"errors"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Debt struct {
	gorm.Model
	Name        string          `gorm:"comment:product name"`
	BankName    string          `gorm:"comment:bank name"`
	BankAccount string          `gorm:"comment:bank account"`
	ApplyTime   time.Time       `gorm:"comment:apply time"`
	EndTime     time.Time       `gorm:"comment:end time"`
	Amount      decimal.Decimal `gorm:"type:decimal(10,2);comment:amount"`
	Tenor       decimal.Decimal `gorm:"comment:tenor"`
	Status      int             `gorm:"comment:status"`
	Remark      string          `gorm:"comment:remark"`
	Apr         decimal.Decimal `gorm:"comment:apr"`
	Fee         decimal.Decimal `gorm:"type:decimal(10,2);comment:fee"`
	UserId      string          `gorm:"comment:user id"`
}

type DebtRepo struct {
	data *Data
	log  *log.Helper
}

func (d *DebtRepo) FindByUserIdAndDebtId(ctx context.Context, userId string, id uint) (*biz.Debt, error) {
	var debt Debt
	tx := d.data.db.WithContext(ctx).Where("user_id = ? and id = ?", userId, id).First(&debt)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return mapDebtToBiz(&debt), nil
}

func (d *DebtRepo) ListByUserId(ctx context.Context, userId string, query *biz.DebtListQuery) ([]*biz.Debt, int64, error) {
	db := d.data.db.WithContext(ctx).Model(&Debt{}).Where("user_id = ?", userId)

	if query != nil {
		if query.Name != "" {
			db = db.Where("name LIKE ?", "%"+query.Name+"%")
		}
		if query.BankName != "" {
			db = db.Where("bank_name LIKE ?", "%"+query.BankName+"%")
		}
		if query.Status != nil {
			db = db.Where("status = ?", *query.Status)
		}
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := int64(1)
	pageSize := int64(10)
	if query != nil {
		if query.Page > 0 {
			page = query.Page
		}
		if query.PageSize > 0 {
			pageSize = query.PageSize
		}
	}

	var debts []Debt
	err := db.Order("created_at DESC").
		Offset(int((page - 1) * pageSize)).
		Limit(int(pageSize)).
		Find(&debts).Error
	if err != nil {
		return nil, 0, err
	}

	items := make([]*biz.Debt, 0, len(debts))
	for i := range debts {
		items = append(items, mapDebtToBiz(&debts[i]))
	}
	return items, total, nil
}

func (d *DebtRepo) Save(ctx context.Context, debt *biz.Debt) (uint, error) {
	applyTime, err := time.ParseInLocation("2006-01-02 15:04:05", debt.ApplyTime, time.Local)
	if err != nil {
		return 0, err
	}
	endTime, err := time.ParseInLocation("2006-01-02 15:04:05", debt.EndTime, time.Local)
	if err != nil {
		return 0, err
	}

	debtDb := Debt{
		Name:        debt.Name,
		BankName:    debt.BankName,
		BankAccount: debt.BankAccount,
		ApplyTime:   applyTime,
		EndTime:     endTime,
		Amount:      debt.Amount,
		Tenor:       debt.Tenor,
		Status:      debt.Status,
		Remark:      debt.Remark,
		Apr:         debt.Apr,
		Fee:         debt.Fee,
		UserId:      debt.UserId,
	}
	tx := d.data.db.WithContext(ctx).Create(&debtDb)
	if tx.Error != nil {
		return 0, tx.Error
	}
	return debtDb.ID, nil
}

func (d *DebtRepo) Update(ctx context.Context, debt *biz.Debt) error {
	debtDb := Debt{
		Name:        debt.Name,
		BankName:    debt.BankName,
		BankAccount: debt.BankAccount,
		Amount:      debt.Amount,
		Tenor:       debt.Tenor,
		Status:      debt.Status,
		Remark:      debt.Remark,
		Apr:         debt.Apr,
		Fee:         debt.Fee,
		UserId:      debt.UserId,
	}
	debtDb.ID = uint(debt.Id)
	if debt.ApplyTime != "" {
		applyTime, err := time.ParseInLocation("2006-01-02 15:04:05", debt.ApplyTime, time.Local)
		if err == nil {
			debtDb.ApplyTime = applyTime
		}
	}

	if debt.EndTime != "" {
		endTime, err := time.ParseInLocation("2006-01-02 15:04:05", debt.EndTime, time.Local)
		if err == nil {
			debtDb.EndTime = endTime
		}
	}
	tx := d.data.db.WithContext(ctx).Model(&Debt{}).Where("id = ? AND user_id = ?", debt.Id, debt.UserId).Updates(debtDb)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("no updatable debt found for current user")
	}
	return nil
}

func (d *DebtRepo) DeleteByUserIdAndDebtId(ctx context.Context, userId string, id uint) error {
	tx := d.data.db.WithContext(ctx).Where("user_id = ? and id = ?", userId, id).Delete(&Debt{})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("no deletable debt found or permission denied")
	}
	return nil
}

func (d *DebtRepo) CountDebtDetailByDebtId(ctx context.Context, debtId uint) (int64, error) {
	var count int64
	tx := d.data.db.WithContext(ctx).Model(&DebtDetail{}).Where("debt_id = ?", debtId).Count(&count)
	if tx.Error != nil {
		return 0, tx.Error
	}
	return count, nil
}

func (d *DebtRepo) FindByID(ctx context.Context, id int64) (*biz.Debt, error) {
	var debt Debt
	tx := d.data.db.WithContext(ctx).First(&debt, id)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return mapDebtToBiz(&debt), nil
}

func (d *DebtRepo) ListByHello(ctx context.Context, keyword string) ([]*biz.Debt, error) {
	var debts []Debt
	tx := d.data.db.WithContext(ctx).Where("name LIKE ?", "%"+keyword+"%").Order("created_at DESC").Find(&debts)
	if tx.Error != nil {
		return nil, tx.Error
	}
	items := make([]*biz.Debt, 0, len(debts))
	for i := range debts {
		items = append(items, mapDebtToBiz(&debts[i]))
	}
	return items, nil
}

func (d *DebtRepo) ListAll(ctx context.Context) ([]*biz.Debt, error) {
	var debts []Debt
	tx := d.data.db.WithContext(ctx).Order("created_at DESC").Find(&debts)
	if tx.Error != nil {
		return nil, tx.Error
	}
	items := make([]*biz.Debt, 0, len(debts))
	for i := range debts {
		items = append(items, mapDebtToBiz(&debts[i]))
	}
	return items, nil
}

func NewDebtRepo(data *Data, logger log.Logger) biz.DebtRepo {
	return &DebtRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func mapDebtToBiz(debt *Debt) *biz.Debt {
	if debt == nil {
		return nil
	}
	return &biz.Debt{
		Id:          int64(debt.ID),
		Name:        debt.Name,
		BankName:    debt.BankName,
		BankAccount: debt.BankAccount,
		ApplyTime:   debt.ApplyTime.Format("2006-01-02 15:04:05"),
		EndTime:     debt.EndTime.Format("2006-01-02 15:04:05"),
		Amount:      debt.Amount,
		Tenor:       debt.Tenor,
		Status:      debt.Status,
		Remark:      debt.Remark,
		Apr:         debt.Apr,
		Fee:         debt.Fee,
		UserId:      debt.UserId,
	}
}
