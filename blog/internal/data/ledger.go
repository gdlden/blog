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

type LedgerAccount struct {
	gorm.Model
	UserId        string           `gorm:"index;comment:user id"`
	Name          string           `gorm:"comment:account name"`
	Type          string           `gorm:"type:varchar(16);index;comment:asset/liability/expense/income/equity"`
	Subtype       string           `gorm:"type:varchar(32);comment:account subtype"`
	CreditLimit   *decimal.Decimal `gorm:"type:decimal(14,2);comment:credit limit"`
	BillingDay    *int32           `gorm:"comment:billing day"`
	PaymentDueDay *int32           `gorm:"comment:payment due day"`
	Remark        string           `gorm:"comment:remark"`
	Sort          int32            `gorm:"default:0;comment:sort"`
	Archived      bool             `gorm:"default:false;comment:archived"`
	IsSystem      bool             `gorm:"default:false;index;comment:system account"`
}

type LedgerCategory struct {
	gorm.Model
	UserId    string `gorm:"index;comment:user id"`
	ParentId  int64  `gorm:"default:0;comment:parent category id"`
	Name      string `gorm:"comment:category name"`
	Direction string `gorm:"type:varchar(16);index;comment:expense/income"`
	Sort      int32  `gorm:"default:0;comment:sort"`
	IsSystem  bool   `gorm:"default:false;comment:system fallback category"`
}

type LedgerTransaction struct {
	gorm.Model
	UserId   string    `gorm:"index;comment:user id"`
	Type     string    `gorm:"type:varchar(32);index;comment:expense/income/transfer/opening_balance/adjustment"`
	BookedAt time.Time `gorm:"index;comment:booked at"`
	Remark   string    `gorm:"comment:remark"`
}

// LedgerPosting 交易分录腿。随交易整组替换/删除，不使用软删。
type LedgerPosting struct {
	Id            uint            `gorm:"primaryKey;autoIncrement"`
	TransactionId uint            `gorm:"index;comment:transaction id"`
	AccountId     uint            `gorm:"index;comment:account id"`
	Amount        decimal.Decimal `gorm:"type:decimal(14,2);comment:signed amount"`
	CategoryId    uint            `gorm:"index;comment:category id, 0 = none"`
	Sort          int32           `gorm:"default:0;comment:sort"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// LedgerBudget 分类月度预算，(user_id, category_id, month) 唯一。
type LedgerBudget struct {
	gorm.Model
	UserId     string          `gorm:"uniqueIndex:idx_ledger_budget_ucm;comment:user id"`
	CategoryId uint            `gorm:"uniqueIndex:idx_ledger_budget_ucm;comment:category id"`
	Month      string          `gorm:"type:varchar(7);uniqueIndex:idx_ledger_budget_ucm;comment:month YYYY-MM"`
	Amount     decimal.Decimal `gorm:"type:decimal(14,2);comment:budget amount"`
}

// LedgerRecurring 周期账单规则：每月 dayOfMonth 日惰性生成一笔 expense/income 交易。
type LedgerRecurring struct {
	gorm.Model
	UserId             string          `gorm:"index;comment:user id"`
	AccountId          uint            `gorm:"index;comment:account id"`
	CategoryId         uint            `gorm:"index;comment:category id"`
	Type               string          `gorm:"type:varchar(16);comment:expense/income"`
	Amount             decimal.Decimal `gorm:"type:decimal(14,2);comment:amount"`
	Remark             string          `gorm:"comment:remark"`
	DayOfMonth         int32           `gorm:"comment:day of month 1-31"`
	StartMonth         string          `gorm:"type:varchar(7);comment:start month YYYY-MM"`
	Freq               string          `gorm:"type:varchar(16);default:monthly;comment:frequency, reserved"`
	LastGeneratedMonth string          `gorm:"type:varchar(7);default:'';comment:last generated month YYYY-MM"`
	Enabled            bool            `gorm:"comment:enabled"`
}

// ledgerPresetCategories 预置分类集（每用户首次 list 时插入），"其它"为 isSystem 兜底分类。
var ledgerPresetCategories = []struct {
	Name      string
	Direction string
	IsSystem  bool
}{
	{"餐饮", "expense", false},
	{"交通", "expense", false},
	{"购物", "expense", false},
	{"居住", "expense", false},
	{"娱乐", "expense", false},
	{"医疗", "expense", false},
	{"教育", "expense", false},
	{"通讯", "expense", false},
	{"人情", "expense", false},
	{"旅行", "expense", false},
	{"数码", "expense", false},
	{"其它", "expense", true},
	{"工资", "income", false},
	{"奖金", "income", false},
	{"理财", "income", false},
	{"报销", "income", false},
	{"其它", "income", true},
}

type LedgerAccountRepo struct {
	data *Data
	log  *log.Helper
}

type LedgerCategoryRepo struct {
	data *Data
	log  *log.Helper
}

type LedgerTransactionRepo struct {
	data *Data
	log  *log.Helper
}

type LedgerBudgetRepo struct {
	data *Data
	log  *log.Helper
}

type LedgerRecurringRepo struct {
	data *Data
	log  *log.Helper
}

func NewLedgerAccountRepo(data *Data, logger log.Logger) biz.LedgerAccountRepo {
	return &LedgerAccountRepo{data: data, log: log.NewHelper(logger)}
}

func NewLedgerCategoryRepo(data *Data, logger log.Logger) biz.LedgerCategoryRepo {
	return &LedgerCategoryRepo{data: data, log: log.NewHelper(logger)}
}

func NewLedgerTransactionRepo(data *Data, logger log.Logger) biz.LedgerTransactionRepo {
	return &LedgerTransactionRepo{data: data, log: log.NewHelper(logger)}
}

func NewLedgerBudgetRepo(data *Data, logger log.Logger) biz.LedgerBudgetRepo {
	return &LedgerBudgetRepo{data: data, log: log.NewHelper(logger)}
}

func NewLedgerRecurringRepo(data *Data, logger log.Logger) biz.LedgerRecurringRepo {
	return &LedgerRecurringRepo{data: data, log: log.NewHelper(logger)}
}

// ---------- 账户 ----------

func (r *LedgerAccountRepo) Save(ctx context.Context, account *biz.LedgerAccount) (uint, error) {
	dbAccount := LedgerAccount{
		UserId:        account.UserId,
		Name:          account.Name,
		Type:          account.Type,
		Subtype:       account.Subtype,
		CreditLimit:   account.CreditLimit,
		BillingDay:    account.BillingDay,
		PaymentDueDay: account.PaymentDueDay,
		Remark:        account.Remark,
		Sort:          account.Sort,
		Archived:      account.Archived,
		IsSystem:      account.IsSystem,
	}
	tx := r.data.db.WithContext(ctx).Create(&dbAccount)
	if tx.Error != nil {
		return 0, tx.Error
	}
	return dbAccount.ID, nil
}

func (r *LedgerAccountRepo) Update(ctx context.Context, account *biz.LedgerAccount) error {
	updates := map[string]any{
		"name":            account.Name,
		"type":            account.Type,
		"subtype":         account.Subtype,
		"credit_limit":    account.CreditLimit,
		"billing_day":     account.BillingDay,
		"payment_due_day": account.PaymentDueDay,
		"remark":          account.Remark,
		"sort":            account.Sort,
		"archived":        account.Archived,
	}
	tx := r.data.db.WithContext(ctx).
		Model(&LedgerAccount{}).
		Where("id = ? AND user_id = ? AND is_system = ?", account.Id, account.UserId, false).
		Updates(updates)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("no updatable ledger account found for current user")
	}
	return nil
}

func (r *LedgerAccountRepo) DeleteByUserIdAndId(ctx context.Context, userId string, id uint) error {
	tx := r.data.db.WithContext(ctx).Where("user_id = ? AND id = ? AND is_system = ?", userId, id, false).Delete(&LedgerAccount{})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("no deletable ledger account found or permission denied")
	}
	return nil
}

func (r *LedgerAccountRepo) FindByUserIdAndId(ctx context.Context, userId string, id uint) (*biz.LedgerAccount, error) {
	var account LedgerAccount
	tx := r.data.db.WithContext(ctx).Where("user_id = ? AND id = ?", userId, id).First(&account)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return mapLedgerAccountToBiz(&account), nil
}

func (r *LedgerAccountRepo) FindSystemByUserIdAndType(ctx context.Context, userId string, accountType string) (*biz.LedgerAccount, error) {
	var account LedgerAccount
	tx := r.data.db.WithContext(ctx).
		Where("user_id = ? AND type = ? AND is_system = ?", userId, accountType, true).
		First(&account)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return mapLedgerAccountToBiz(&account), nil
}

func (r *LedgerAccountRepo) ListByUserId(ctx context.Context, userId string) ([]*biz.LedgerAccount, error) {
	var accounts []LedgerAccount
	err := r.data.db.WithContext(ctx).
		Where("user_id = ? AND is_system = ?", userId, false).
		Order("sort ASC, id ASC").
		Find(&accounts).Error
	if err != nil {
		return nil, err
	}
	items := make([]*biz.LedgerAccount, 0, len(accounts))
	for i := range accounts {
		items = append(items, mapLedgerAccountToBiz(&accounts[i]))
	}
	return items, nil
}

func (r *LedgerAccountRepo) EnsureSystemAccounts(ctx context.Context, userId string) error {
	for accountType, name := range biz.LedgerSystemAccountSpec() {
		var count int64
		if err := r.data.db.WithContext(ctx).Model(&LedgerAccount{}).
			Where("user_id = ? AND type = ? AND is_system = ?", userId, accountType, true).
			Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		if err := r.data.db.WithContext(ctx).Create(&LedgerAccount{
			UserId:   userId,
			Name:     name,
			Type:     accountType,
			IsSystem: true,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *LedgerAccountRepo) SumBalanceByAccountIds(ctx context.Context, accountIds []uint) (map[uint]decimal.Decimal, error) {
	result := make(map[uint]decimal.Decimal, len(accountIds))
	if len(accountIds) == 0 {
		return result, nil
	}
	var postings []LedgerPosting
	if err := r.data.db.WithContext(ctx).
		Where("account_id IN ?", accountIds).
		Find(&postings).Error; err != nil {
		return nil, err
	}
	for i := range postings {
		result[postings[i].AccountId] = result[postings[i].AccountId].Add(postings[i].Amount)
	}
	return result, nil
}

func (r *LedgerAccountRepo) CountPostingsByAccountId(ctx context.Context, accountId uint) (int64, error) {
	var count int64
	tx := r.data.db.WithContext(ctx).Model(&LedgerPosting{}).Where("account_id = ?", accountId).Count(&count)
	if tx.Error != nil {
		return 0, tx.Error
	}
	return count, nil
}

// ---------- 分类 ----------

func (r *LedgerCategoryRepo) Save(ctx context.Context, category *biz.LedgerCategory) (uint, error) {
	dbCategory := LedgerCategory{
		UserId:    category.UserId,
		ParentId:  category.ParentId,
		Name:      category.Name,
		Direction: category.Direction,
		Sort:      category.Sort,
		IsSystem:  category.IsSystem,
	}
	tx := r.data.db.WithContext(ctx).Create(&dbCategory)
	if tx.Error != nil {
		return 0, tx.Error
	}
	return dbCategory.ID, nil
}

func (r *LedgerCategoryRepo) Update(ctx context.Context, category *biz.LedgerCategory) error {
	updates := map[string]any{
		"parent_id": category.ParentId,
		"name":      category.Name,
		"direction": category.Direction,
		"sort":      category.Sort,
	}
	tx := r.data.db.WithContext(ctx).
		Model(&LedgerCategory{}).
		Where("id = ? AND user_id = ? AND is_system = ?", category.Id, category.UserId, false).
		Updates(updates)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("no updatable ledger category found for current user")
	}
	return nil
}

func (r *LedgerCategoryRepo) FindByUserIdAndId(ctx context.Context, userId string, id uint) (*biz.LedgerCategory, error) {
	var category LedgerCategory
	tx := r.data.db.WithContext(ctx).Where("user_id = ? AND id = ?", userId, id).First(&category)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return mapLedgerCategoryToBiz(&category), nil
}

func (r *LedgerCategoryRepo) ListByUserId(ctx context.Context, userId string) ([]*biz.LedgerCategory, error) {
	var categories []LedgerCategory
	err := r.data.db.WithContext(ctx).
		Where("user_id = ?", userId).
		Order("sort ASC, id ASC").
		Find(&categories).Error
	if err != nil {
		return nil, err
	}
	items := make([]*biz.LedgerCategory, 0, len(categories))
	for i := range categories {
		items = append(items, mapLedgerCategoryToBiz(&categories[i]))
	}
	return items, nil
}

func (r *LedgerCategoryRepo) EnsurePresetCategories(ctx context.Context, userId string) error {
	var count int64
	if err := r.data.db.WithContext(ctx).Model(&LedgerCategory{}).
		Where("user_id = ?", userId).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	entities := make([]LedgerCategory, 0, len(ledgerPresetCategories))
	for i, preset := range ledgerPresetCategories {
		entities = append(entities, LedgerCategory{
			UserId:    userId,
			Name:      preset.Name,
			Direction: preset.Direction,
			Sort:      int32(i),
			IsSystem:  preset.IsSystem,
		})
	}
	return r.data.db.WithContext(ctx).Create(&entities).Error
}

func (r *LedgerCategoryRepo) HasChildren(ctx context.Context, categoryId uint) (bool, error) {
	var count int64
	if err := r.data.db.WithContext(ctx).Model(&LedgerCategory{}).
		Where("parent_id = ?", categoryId).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *LedgerCategoryRepo) FindFallbackByDirection(ctx context.Context, userId string, direction string) (*biz.LedgerCategory, error) {
	var category LedgerCategory
	tx := r.data.db.WithContext(ctx).
		Where("user_id = ? AND direction = ? AND is_system = ?", userId, direction, true).
		First(&category)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return mapLedgerCategoryToBiz(&category), nil
}

func (r *LedgerCategoryRepo) DeleteWithMigration(ctx context.Context, userId string, id uint, fallbackId uint) error {
	return r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 引用该分类的 posting 迁移到同 direction 的兜底分类
		if err := tx.Model(&LedgerPosting{}).
			Where("category_id = ?", id).
			Update("category_id", fallbackId).Error; err != nil {
			return err
		}
		res := tx.Where("user_id = ? AND id = ? AND is_system = ?", userId, id, false).Delete(&LedgerCategory{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("no deletable ledger category found or permission denied")
		}
		return nil
	})
}

// ---------- 交易 ----------

func (r *LedgerTransactionRepo) Save(ctx context.Context, txBiz *biz.LedgerTransaction) (uint, error) {
	bookedAt, err := parseLedgerTime(txBiz.BookedAt)
	if err != nil {
		return 0, err
	}
	var txId uint
	err = r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		dbTx := LedgerTransaction{
			UserId:   txBiz.UserId,
			Type:     txBiz.Type,
			BookedAt: bookedAt,
			Remark:   txBiz.Remark,
		}
		if err := tx.Create(&dbTx).Error; err != nil {
			return err
		}
		txId = dbTx.ID
		return saveLedgerPostings(tx, dbTx.ID, txBiz.Postings)
	})
	if err != nil {
		return 0, err
	}
	return txId, nil
}

func (r *LedgerTransactionRepo) Update(ctx context.Context, txBiz *biz.LedgerTransaction) error {
	bookedAt, err := parseLedgerTime(txBiz.BookedAt)
	if err != nil {
		return err
	}
	updates := map[string]any{
		"type":      txBiz.Type,
		"booked_at": bookedAt,
		"remark":    txBiz.Remark,
	}
	return r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&LedgerTransaction{}).
			Where("id = ? AND user_id = ?", txBiz.Id, txBiz.UserId).
			Updates(updates)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("no updatable ledger transaction found for current user")
		}
		// postings 整组替换：先删后插（同一事务）
		if err := tx.Where("transaction_id = ?", txBiz.Id).Delete(&LedgerPosting{}).Error; err != nil {
			return err
		}
		return saveLedgerPostings(tx, uint(txBiz.Id), txBiz.Postings)
	})
}

func (r *LedgerTransactionRepo) DeleteByUserIdAndId(ctx context.Context, userId string, id uint) error {
	return r.data.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Where("user_id = ? AND id = ?", userId, id).Delete(&LedgerTransaction{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("no deletable ledger transaction found or permission denied")
		}
		return tx.Where("transaction_id = ?", id).Delete(&LedgerPosting{}).Error
	})
}

func (r *LedgerTransactionRepo) FindByUserIdAndId(ctx context.Context, userId string, id uint) (*biz.LedgerTransaction, error) {
	var record LedgerTransaction
	tx := r.data.db.WithContext(ctx).Where("user_id = ? AND id = ?", userId, id).First(&record)
	if tx.Error != nil {
		return nil, tx.Error
	}
	item := mapLedgerTransactionToBiz(&record)
	postings, err := r.loadPostings(ctx, []uint{record.ID})
	if err != nil {
		return nil, err
	}
	item.Postings = postings[record.ID]
	return item, nil
}

func (r *LedgerTransactionRepo) PageByUserId(ctx context.Context, userId string, query *biz.LedgerTransactionQuery) ([]*biz.LedgerTransaction, int64, error) {
	db := r.data.db.WithContext(ctx).Model(&LedgerTransaction{}).Where("user_id = ?", userId)
	if query != nil {
		if query.Type != "" {
			db = db.Where("type = ?", query.Type)
		}
		if query.AccountId != 0 {
			db = db.Where("EXISTS (SELECT 1 FROM ledger_postings WHERE ledger_postings.transaction_id = ledger_transactions.id AND ledger_postings.account_id = ?)", query.AccountId)
		}
		if query.CategoryId != 0 {
			db = db.Where("EXISTS (SELECT 1 FROM ledger_postings WHERE ledger_postings.transaction_id = ledger_transactions.id AND ledger_postings.category_id = ?)", query.CategoryId)
		}
		if query.StartTime != "" {
			start, err := parseLedgerTime(query.StartTime)
			if err != nil {
				return nil, 0, err
			}
			db = db.Where("booked_at >= ?", start)
		}
		if query.EndTime != "" {
			end, err := parseLedgerTime(query.EndTime)
			if err != nil {
				return nil, 0, err
			}
			db = db.Where("booked_at <= ?", end)
		}
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page, pageSize := normalizeLedgerPage(query)
	var records []LedgerTransaction
	err := db.Order("booked_at DESC, id DESC").
		Offset(int((page - 1) * pageSize)).
		Limit(int(pageSize)).
		Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	items := make([]*biz.LedgerTransaction, 0, len(records))
	for i := range records {
		items = append(items, mapLedgerTransactionToBiz(&records[i]))
	}
	if len(records) > 0 {
		txIds := make([]uint, 0, len(records))
		for i := range records {
			txIds = append(txIds, records[i].ID)
		}
		postings, err := r.loadPostings(ctx, txIds)
		if err != nil {
			return nil, 0, err
		}
		for i := range items {
			items[i].Postings = postings[uint(items[i].Id)]
		}
	}
	return items, total, nil
}

// ListSystemLegsByMonth 返回指定月份内、指定交易类型、落在系统账户上的分录腿（按交易软删过滤）。
func (r *LedgerTransactionRepo) ListSystemLegsByMonth(ctx context.Context, userId string, txType string, systemAccountId uint, month string) ([]*biz.LedgerPosting, error) {
	start, err := time.Parse("2006-01", month)
	if err != nil {
		return nil, err
	}
	end := start.AddDate(0, 1, 0)
	var postings []LedgerPosting
	err = r.data.db.WithContext(ctx).
		Model(&LedgerPosting{}).
		Joins("JOIN ledger_transactions ON ledger_transactions.id = ledger_postings.transaction_id").
		Where("ledger_transactions.deleted_at IS NULL").
		Where("ledger_transactions.user_id = ? AND ledger_transactions.type = ?", userId, txType).
		Where("ledger_postings.account_id = ?", systemAccountId).
		Where("ledger_transactions.booked_at >= ? AND ledger_transactions.booked_at < ?", start, end).
		Find(&postings).Error
	if err != nil {
		return nil, err
	}
	items := make([]*biz.LedgerPosting, 0, len(postings))
	for i := range postings {
		items = append(items, mapLedgerPostingToBiz(&postings[i]))
	}
	return items, nil
}

// ListAccountLegsUpTo 返回指定账户在 endExclusive 之前的所有分录腿（排除软删交易），按记账时间升序。
func (r *LedgerTransactionRepo) ListAccountLegsUpTo(ctx context.Context, accountIds []uint, endExclusive time.Time) ([]*biz.LedgerAccountLeg, error) {
	result := make([]*biz.LedgerAccountLeg, 0)
	if len(accountIds) == 0 {
		return result, nil
	}
	type legRow struct {
		AccountId uint
		BookedAt  time.Time
		Amount    decimal.Decimal
	}
	var rows []legRow
	err := r.data.db.WithContext(ctx).
		Model(&LedgerPosting{}).
		Select("ledger_postings.account_id, ledger_transactions.booked_at, ledger_postings.amount").
		Joins("JOIN ledger_transactions ON ledger_transactions.id = ledger_postings.transaction_id").
		Where("ledger_transactions.deleted_at IS NULL").
		Where("ledger_postings.account_id IN ?", accountIds).
		Where("ledger_transactions.booked_at < ?", endExclusive).
		Order("ledger_transactions.booked_at ASC, ledger_postings.id ASC").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	for i := range rows {
		result = append(result, &biz.LedgerAccountLeg{
			AccountId: int64(rows[i].AccountId),
			BookedAt:  rows[i].BookedAt.Format("2006-01-02 15:04:05"),
			Amount:    rows[i].Amount,
		})
	}
	return result, nil
}

func (r *LedgerTransactionRepo) loadPostings(ctx context.Context, txIds []uint) (map[uint][]*biz.LedgerPosting, error) {
	result := make(map[uint][]*biz.LedgerPosting)
	if len(txIds) == 0 {
		return result, nil
	}
	var postings []LedgerPosting
	if err := r.data.db.WithContext(ctx).
		Where("transaction_id IN ?", txIds).
		Order("sort ASC, id ASC").
		Find(&postings).Error; err != nil {
		return nil, err
	}
	for i := range postings {
		result[postings[i].TransactionId] = append(result[postings[i].TransactionId], mapLedgerPostingToBiz(&postings[i]))
	}
	return result, nil
}

func saveLedgerPostings(tx *gorm.DB, txId uint, postings []*biz.LedgerPosting) error {
	if len(postings) == 0 {
		return nil
	}
	entities := make([]LedgerPosting, 0, len(postings))
	for _, posting := range postings {
		entities = append(entities, LedgerPosting{
			TransactionId: txId,
			AccountId:     uint(posting.AccountId),
			Amount:        posting.Amount,
			CategoryId:    uint(posting.CategoryId),
			Sort:          posting.Sort,
		})
	}
	return tx.Create(&entities).Error
}

func parseLedgerTime(value string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02 15:04:05", value, time.Local)
}

// ---------- 预算 ----------

// Upsert 同用户同分类同月已存在则更新 amount，否则新建。
func (r *LedgerBudgetRepo) Upsert(ctx context.Context, budget *biz.LedgerBudget) (uint, error) {
	var existing LedgerBudget
	tx := r.data.db.WithContext(ctx).
		Where("user_id = ? AND category_id = ? AND month = ?", budget.UserId, budget.CategoryId, budget.Month).
		First(&existing)
	if tx.Error == nil {
		res := r.data.db.WithContext(ctx).Model(&LedgerBudget{}).
			Where("id = ?", existing.ID).
			Update("amount", budget.Amount)
		if res.Error != nil {
			return 0, res.Error
		}
		return existing.ID, nil
	}
	if !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return 0, tx.Error
	}
	dbBudget := LedgerBudget{
		UserId:     budget.UserId,
		CategoryId: uint(budget.CategoryId),
		Month:      budget.Month,
		Amount:     budget.Amount,
	}
	if err := r.data.db.WithContext(ctx).Create(&dbBudget).Error; err != nil {
		return 0, err
	}
	return dbBudget.ID, nil
}

func (r *LedgerBudgetRepo) DeleteByUserIdAndId(ctx context.Context, userId string, id uint) error {
	tx := r.data.db.WithContext(ctx).Where("user_id = ? AND id = ?", userId, id).Delete(&LedgerBudget{})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("no deletable ledger budget found or permission denied")
	}
	return nil
}

func (r *LedgerBudgetRepo) ListByUserIdAndMonth(ctx context.Context, userId string, month string) ([]*biz.LedgerBudget, error) {
	var budgets []LedgerBudget
	err := r.data.db.WithContext(ctx).
		Where("user_id = ? AND month = ?", userId, month).
		Order("id ASC").
		Find(&budgets).Error
	if err != nil {
		return nil, err
	}
	items := make([]*biz.LedgerBudget, 0, len(budgets))
	for i := range budgets {
		items = append(items, &biz.LedgerBudget{
			Id:         int64(budgets[i].ID),
			UserId:     budgets[i].UserId,
			CategoryId: int64(budgets[i].CategoryId),
			Month:      budgets[i].Month,
			Amount:     budgets[i].Amount,
		})
	}
	return items, nil
}

// ---------- 周期账单 ----------

func (r *LedgerRecurringRepo) Save(ctx context.Context, rule *biz.LedgerRecurring) (uint, error) {
	dbRule := LedgerRecurring{
		UserId:     rule.UserId,
		AccountId:  uint(rule.AccountId),
		CategoryId: uint(rule.CategoryId),
		Type:       rule.Type,
		Amount:     rule.Amount,
		Remark:     rule.Remark,
		DayOfMonth: rule.DayOfMonth,
		StartMonth: rule.StartMonth,
		Freq:       "monthly",
		Enabled:    rule.Enabled,
	}
	tx := r.data.db.WithContext(ctx).Create(&dbRule)
	if tx.Error != nil {
		return 0, tx.Error
	}
	return dbRule.ID, nil
}

// Update 整字段更新（不含 last_generated_month，生成进度只走 UpdateLastGeneratedMonth）。
func (r *LedgerRecurringRepo) Update(ctx context.Context, rule *biz.LedgerRecurring) error {
	updates := map[string]any{
		"account_id":   rule.AccountId,
		"category_id":  rule.CategoryId,
		"type":         rule.Type,
		"amount":       rule.Amount,
		"remark":       rule.Remark,
		"day_of_month": rule.DayOfMonth,
		"start_month":  rule.StartMonth,
		"enabled":      rule.Enabled,
	}
	tx := r.data.db.WithContext(ctx).
		Model(&LedgerRecurring{}).
		Where("id = ? AND user_id = ?", rule.Id, rule.UserId).
		Updates(updates)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("no updatable ledger recurring found for current user")
	}
	return nil
}

func (r *LedgerRecurringRepo) DeleteByUserIdAndId(ctx context.Context, userId string, id uint) error {
	tx := r.data.db.WithContext(ctx).Where("user_id = ? AND id = ?", userId, id).Delete(&LedgerRecurring{})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("no deletable ledger recurring found or permission denied")
	}
	return nil
}

func (r *LedgerRecurringRepo) FindByUserIdAndId(ctx context.Context, userId string, id uint) (*biz.LedgerRecurring, error) {
	var rule LedgerRecurring
	tx := r.data.db.WithContext(ctx).Where("user_id = ? AND id = ?", userId, id).First(&rule)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return mapLedgerRecurringToBiz(&rule), nil
}

func (r *LedgerRecurringRepo) ListByUserId(ctx context.Context, userId string) ([]*biz.LedgerRecurring, error) {
	var rules []LedgerRecurring
	err := r.data.db.WithContext(ctx).
		Where("user_id = ?", userId).
		Order("id ASC").
		Find(&rules).Error
	if err != nil {
		return nil, err
	}
	items := make([]*biz.LedgerRecurring, 0, len(rules))
	for i := range rules {
		items = append(items, mapLedgerRecurringToBiz(&rules[i]))
	}
	return items, nil
}

func (r *LedgerRecurringRepo) ListEnabledByUserId(ctx context.Context, userId string) ([]*biz.LedgerRecurring, error) {
	var rules []LedgerRecurring
	err := r.data.db.WithContext(ctx).
		Where("user_id = ? AND enabled = ?", userId, true).
		Order("id ASC").
		Find(&rules).Error
	if err != nil {
		return nil, err
	}
	items := make([]*biz.LedgerRecurring, 0, len(rules))
	for i := range rules {
		items = append(items, mapLedgerRecurringToBiz(&rules[i]))
	}
	return items, nil
}

func (r *LedgerRecurringRepo) UpdateLastGeneratedMonth(ctx context.Context, userId string, id uint, month string) error {
	tx := r.data.db.WithContext(ctx).
		Model(&LedgerRecurring{}).
		Where("id = ? AND user_id = ?", id, userId).
		Update("last_generated_month", month)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("no updatable ledger recurring found for current user")
	}
	return nil
}

func normalizeLedgerPage(query *biz.LedgerTransactionQuery) (int64, int64) {
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
	return page, pageSize
}

func mapLedgerAccountToBiz(account *LedgerAccount) *biz.LedgerAccount {
	if account == nil {
		return nil
	}
	return &biz.LedgerAccount{
		Id:            int64(account.ID),
		UserId:        account.UserId,
		Name:          account.Name,
		Type:          account.Type,
		Subtype:       account.Subtype,
		CreditLimit:   account.CreditLimit,
		BillingDay:    account.BillingDay,
		PaymentDueDay: account.PaymentDueDay,
		Remark:        account.Remark,
		Sort:          account.Sort,
		Archived:      account.Archived,
		IsSystem:      account.IsSystem,
	}
}

func mapLedgerCategoryToBiz(category *LedgerCategory) *biz.LedgerCategory {
	if category == nil {
		return nil
	}
	return &biz.LedgerCategory{
		Id:        int64(category.ID),
		UserId:    category.UserId,
		ParentId:  category.ParentId,
		Name:      category.Name,
		Direction: category.Direction,
		Sort:      category.Sort,
		IsSystem:  category.IsSystem,
	}
}

func mapLedgerTransactionToBiz(record *LedgerTransaction) *biz.LedgerTransaction {
	if record == nil {
		return nil
	}
	return &biz.LedgerTransaction{
		Id:        int64(record.ID),
		UserId:    record.UserId,
		Type:      record.Type,
		BookedAt:  record.BookedAt.Format("2006-01-02 15:04:05"),
		Remark:    record.Remark,
		CreatedAt: record.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func mapLedgerPostingToBiz(posting *LedgerPosting) *biz.LedgerPosting {
	if posting == nil {
		return nil
	}
	return &biz.LedgerPosting{
		Id:            int64(posting.Id),
		TransactionId: int64(posting.TransactionId),
		AccountId:     int64(posting.AccountId),
		Amount:        posting.Amount,
		CategoryId:    int64(posting.CategoryId),
		Sort:          posting.Sort,
	}
}

func mapLedgerRecurringToBiz(rule *LedgerRecurring) *biz.LedgerRecurring {
	if rule == nil {
		return nil
	}
	return &biz.LedgerRecurring{
		Id:                 int64(rule.ID),
		UserId:             rule.UserId,
		AccountId:          int64(rule.AccountId),
		CategoryId:         int64(rule.CategoryId),
		Type:               rule.Type,
		Amount:             rule.Amount,
		Remark:             rule.Remark,
		DayOfMonth:         rule.DayOfMonth,
		StartMonth:         rule.StartMonth,
		LastGeneratedMonth: rule.LastGeneratedMonth,
		Enabled:            rule.Enabled,
	}
}
