package data

import (
	"context"
	"testing"
	"time"

	"blog/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupLedgerTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	if err := db.AutoMigrate(&LedgerAccount{}, &LedgerCategory{}, &LedgerTransaction{}, &LedgerPosting{}); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}
	return db
}

func newLedgerRepos(db *gorm.DB) (*LedgerAccountRepo, *LedgerCategoryRepo, *LedgerTransactionRepo) {
	helper := log.NewHelper(log.DefaultLogger)
	data := &Data{db: db}
	return &LedgerAccountRepo{data: data, log: helper},
		&LedgerCategoryRepo{data: data, log: helper},
		&LedgerTransactionRepo{data: data, log: helper}
}

func TestLedgerTransactionRepo_SaveRollbackOnPostingFailure(t *testing.T) {
	db := setupLedgerTestDB(t)
	accountRepo, _, txRepo := newLedgerRepos(db)
	ctx := context.Background()

	accountId, err := accountRepo.Save(ctx, &biz.LedgerAccount{UserId: "user-123", Name: "现金", Type: "asset", Subtype: "cash"})
	assert.NoError(t, err)

	newTx := func() *biz.LedgerTransaction {
		return &biz.LedgerTransaction{
			UserId:   "user-123",
			Type:     "transfer",
			BookedAt: "2026-08-01 10:00:00",
			Postings: []*biz.LedgerPosting{
				{AccountId: int64(accountId), Amount: decimal.NewFromInt(-100)},
				{AccountId: int64(accountId), Amount: decimal.NewFromInt(100)},
			},
		}
	}

	// 正常保存：transaction + postings 都落库
	txId, err := txRepo.Save(ctx, newTx())
	assert.NoError(t, err)
	var postingCount int64
	db.Model(&LedgerPosting{}).Where("transaction_id = ?", txId).Count(&postingCount)
	assert.Equal(t, int64(2), postingCount)

	// 制造 postings 写入失败（删表），整个事务必须回滚，transaction 不留痕
	assert.NoError(t, db.Migrator().DropTable(&LedgerPosting{}))
	var before int64
	db.Model(&LedgerTransaction{}).Count(&before)
	_, err = txRepo.Save(ctx, newTx())
	assert.Error(t, err)
	var after int64
	db.Model(&LedgerTransaction{}).Count(&after)
	assert.Equal(t, before, after)
}

func TestLedgerAccountRepo_SumBalanceByAccountIds(t *testing.T) {
	db := setupLedgerTestDB(t)
	accountRepo, _, txRepo := newLedgerRepos(db)
	ctx := context.Background()

	cashId, err := accountRepo.Save(ctx, &biz.LedgerAccount{UserId: "user-123", Name: "现金", Type: "asset", Subtype: "cash"})
	assert.NoError(t, err)
	cardId, err := accountRepo.Save(ctx, &biz.LedgerAccount{UserId: "user-123", Name: "储蓄卡", Type: "asset", Subtype: "debit_card"})
	assert.NoError(t, err)

	// 消费 100（现金 -100，系统费用 +100），转账 50（现金 -50，储蓄卡 +50）
	_, err = txRepo.Save(ctx, &biz.LedgerTransaction{
		UserId: "user-123", Type: "expense", BookedAt: "2026-08-01 10:00:00",
		Postings: []*biz.LedgerPosting{
			{AccountId: int64(cashId), Amount: decimal.NewFromInt(-100)},
			{AccountId: 999, Amount: decimal.NewFromInt(100)},
		},
	})
	assert.NoError(t, err)
	_, err = txRepo.Save(ctx, &biz.LedgerTransaction{
		UserId: "user-123", Type: "transfer", BookedAt: "2026-08-02 10:00:00",
		Postings: []*biz.LedgerPosting{
			{AccountId: int64(cashId), Amount: decimal.NewFromInt(-50)},
			{AccountId: int64(cardId), Amount: decimal.NewFromInt(50)},
		},
	})
	assert.NoError(t, err)

	sums, err := accountRepo.SumBalanceByAccountIds(ctx, []uint{cashId, cardId})

	assert.NoError(t, err)
	assert.True(t, sums[cashId].Equal(decimal.NewFromInt(-150)))
	assert.True(t, sums[cardId].Equal(decimal.NewFromInt(50)))
}

func TestLedgerCategoryRepo_DeleteWithMigrationMovesPostings(t *testing.T) {
	db := setupLedgerTestDB(t)
	accountRepo, categoryRepo, txRepo := newLedgerRepos(db)
	ctx := context.Background()

	accountId, err := accountRepo.Save(ctx, &biz.LedgerAccount{UserId: "user-123", Name: "现金", Type: "asset", Subtype: "cash"})
	assert.NoError(t, err)
	categoryId, err := categoryRepo.Save(ctx, &biz.LedgerCategory{UserId: "user-123", Name: "数码", Direction: "expense"})
	assert.NoError(t, err)
	fallbackId, err := categoryRepo.Save(ctx, &biz.LedgerCategory{UserId: "user-123", Name: "其它", Direction: "expense", IsSystem: true})
	assert.NoError(t, err)

	txId, err := txRepo.Save(ctx, &biz.LedgerTransaction{
		UserId: "user-123", Type: "expense", BookedAt: "2026-08-01 10:00:00",
		Postings: []*biz.LedgerPosting{
			{AccountId: int64(accountId), Amount: decimal.NewFromInt(-100)},
			{AccountId: 999, Amount: decimal.NewFromInt(100), CategoryId: int64(categoryId)},
		},
	})
	assert.NoError(t, err)

	err = categoryRepo.DeleteWithMigration(ctx, "user-123", categoryId, fallbackId)

	assert.NoError(t, err)
	// 分类已删除，posting 的 categoryId 已迁移到兜底分类
	var count int64
	db.Model(&LedgerCategory{}).Where("id = ?", categoryId).Count(&count)
	assert.Equal(t, int64(0), count)
	var posting LedgerPosting
	assert.NoError(t, db.Where("transaction_id = ? AND account_id = ?", txId, 999).First(&posting).Error)
	assert.Equal(t, fallbackId, posting.CategoryId)
}

func TestLedgerTransactionRepo_ListSystemLegsByMonthFilters(t *testing.T) {
	db := setupLedgerTestDB(t)
	accountRepo, _, txRepo := newLedgerRepos(db)
	ctx := context.Background()

	expenseSysId, err := accountRepo.Save(ctx, &biz.LedgerAccount{UserId: "user-123", Name: "费用", Type: "expense", IsSystem: true})
	assert.NoError(t, err)
	cashId, err := accountRepo.Save(ctx, &biz.LedgerAccount{UserId: "user-123", Name: "现金", Type: "asset", Subtype: "cash"})
	assert.NoError(t, err)

	// 8 月 expense 一笔：现金腿 + 系统费用腿（挂分类）
	_, err = txRepo.Save(ctx, &biz.LedgerTransaction{
		UserId: "user-123", Type: "expense", BookedAt: "2026-08-01 10:00:00",
		Postings: []*biz.LedgerPosting{
			{AccountId: int64(cashId), Amount: decimal.NewFromInt(-100)},
			{AccountId: int64(expenseSysId), Amount: decimal.NewFromInt(100), CategoryId: 5},
		},
	})
	assert.NoError(t, err)
	// 8 月 transfer 一笔（信用卡还款）：必须被排除
	_, err = txRepo.Save(ctx, &biz.LedgerTransaction{
		UserId: "user-123", Type: "transfer", BookedAt: "2026-08-10 10:00:00",
		Postings: []*biz.LedgerPosting{
			{AccountId: int64(cashId), Amount: decimal.NewFromInt(-500)},
			{AccountId: 888, Amount: decimal.NewFromInt(500)},
		},
	})
	assert.NoError(t, err)
	// 7 月 expense 一笔：月份外，必须被排除
	_, err = txRepo.Save(ctx, &biz.LedgerTransaction{
		UserId: "user-123", Type: "expense", BookedAt: "2026-07-31 23:59:59",
		Postings: []*biz.LedgerPosting{
			{AccountId: int64(cashId), Amount: decimal.NewFromInt(-30)},
			{AccountId: int64(expenseSysId), Amount: decimal.NewFromInt(30)},
		},
	})
	assert.NoError(t, err)

	legs, err := txRepo.ListSystemLegsByMonth(ctx, "user-123", "expense", expenseSysId, "2026-08")

	assert.NoError(t, err)
	// 只返回 8 月 expense 交易落在系统费用账户上的那条腿
	assert.Len(t, legs, 1)
	assert.True(t, legs[0].Amount.Equal(decimal.NewFromInt(100)))
	assert.Equal(t, int64(5), legs[0].CategoryId)
}

func TestLedgerAccountRepo_EnsureSystemAccountsIdempotent(t *testing.T) {
	db := setupLedgerTestDB(t)
	accountRepo, _, _ := newLedgerRepos(db)
	ctx := context.Background()

	assert.NoError(t, accountRepo.EnsureSystemAccounts(ctx, "user-123"))
	assert.NoError(t, accountRepo.EnsureSystemAccounts(ctx, "user-123"))

	var count int64
	db.Model(&LedgerAccount{}).Where("user_id = ? AND is_system = ?", "user-123", true).Count(&count)
	assert.Equal(t, int64(3), count)

	// list 不返回系统账户
	_, err := accountRepo.Save(ctx, &biz.LedgerAccount{UserId: "user-123", Name: "现金", Type: "asset", Subtype: "cash"})
	assert.NoError(t, err)
	accounts, err := accountRepo.ListByUserId(ctx, "user-123")
	assert.NoError(t, err)
	assert.Len(t, accounts, 1)
	assert.Equal(t, "现金", accounts[0].Name)
}

func TestLedgerCategoryRepo_EnsurePresetCategoriesIdempotent(t *testing.T) {
	db := setupLedgerTestDB(t)
	_, categoryRepo, _ := newLedgerRepos(db)
	ctx := context.Background()

	assert.NoError(t, categoryRepo.EnsurePresetCategories(ctx, "user-123"))
	assert.NoError(t, categoryRepo.EnsurePresetCategories(ctx, "user-123"))

	categories, err := categoryRepo.ListByUserId(ctx, "user-123")
	assert.NoError(t, err)
	// 支出 12 个 + 收入 5 个
	assert.Len(t, categories, 17)
	// 两个方向的"其它"均为 isSystem 兜底分类
	fallback, err := categoryRepo.FindFallbackByDirection(ctx, "user-123", "expense")
	assert.NoError(t, err)
	assert.Equal(t, "其它", fallback.Name)
	assert.True(t, fallback.IsSystem)
	fallback, err = categoryRepo.FindFallbackByDirection(ctx, "user-123", "income")
	assert.NoError(t, err)
	assert.Equal(t, "其它", fallback.Name)
}

func TestLedgerTransactionRepo_UpdateReplacesPostings(t *testing.T) {
	db := setupLedgerTestDB(t)
	accountRepo, _, txRepo := newLedgerRepos(db)
	ctx := context.Background()

	cashId, err := accountRepo.Save(ctx, &biz.LedgerAccount{UserId: "user-123", Name: "现金", Type: "asset", Subtype: "cash"})
	assert.NoError(t, err)
	cardId, err := accountRepo.Save(ctx, &biz.LedgerAccount{UserId: "user-123", Name: "储蓄卡", Type: "asset", Subtype: "debit_card"})
	assert.NoError(t, err)

	txId, err := txRepo.Save(ctx, &biz.LedgerTransaction{
		UserId: "user-123", Type: "transfer", BookedAt: "2026-08-01 10:00:00",
		Postings: []*biz.LedgerPosting{
			{AccountId: int64(cashId), Amount: decimal.NewFromInt(-100)},
			{AccountId: int64(cardId), Amount: decimal.NewFromInt(100)},
		},
	})
	assert.NoError(t, err)

	// 整组替换 postings
	err = txRepo.Update(ctx, &biz.LedgerTransaction{
		Id: int64(txId), UserId: "user-123", Type: "transfer", BookedAt: "2026-08-02 10:00:00",
		Postings: []*biz.LedgerPosting{
			{AccountId: int64(cashId), Amount: decimal.NewFromInt(-200)},
			{AccountId: int64(cardId), Amount: decimal.NewFromInt(200)},
		},
	})
	assert.NoError(t, err)

	item, err := txRepo.FindByUserIdAndId(ctx, "user-123", txId)
	assert.NoError(t, err)
	assert.Equal(t, "2026-08-02 10:00:00", item.BookedAt)
	assert.Len(t, item.Postings, 2)
	assert.True(t, item.Postings[0].Amount.Equal(decimal.NewFromInt(-200)))
	var total int64
	db.Model(&LedgerPosting{}).Where("transaction_id = ?", txId).Count(&total)
	assert.Equal(t, int64(2), total)
}

func setupLedgerBudgetTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	if err := db.AutoMigrate(&LedgerAccount{}, &LedgerCategory{}, &LedgerTransaction{}, &LedgerPosting{}, &LedgerBudget{}, &LedgerRecurring{}); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}
	return db
}

func TestLedgerRecurringRepo_SaveUpdateDeleteList(t *testing.T) {
	db := setupLedgerBudgetTestDB(t)
	helper := log.NewHelper(log.DefaultLogger)
	recurringRepo := &LedgerRecurringRepo{data: &Data{db: db}, log: helper}
	ctx := context.Background()

	// 新建：freq 默认 monthly，last_generated_month 默认空
	id, err := recurringRepo.Save(ctx, &biz.LedgerRecurring{
		UserId: "user-123", AccountId: 1, CategoryId: 5, Type: "expense",
		Amount: decimal.NewFromInt(100), Remark: "房租", DayOfMonth: 1, StartMonth: "2026-08", Enabled: true,
	})
	assert.NoError(t, err)
	rule, err := recurringRepo.FindByUserIdAndId(ctx, "user-123", id)
	assert.NoError(t, err)
	assert.Equal(t, "2026-08", rule.StartMonth)
	assert.Equal(t, "", rule.LastGeneratedMonth)
	assert.True(t, rule.Enabled)
	assert.True(t, rule.Amount.Equal(decimal.NewFromInt(100)))
	var dbRule LedgerRecurring
	assert.NoError(t, db.First(&dbRule, id).Error)
	assert.Equal(t, "monthly", dbRule.Freq)

	// 更新：字段整组替换，但不动 last_generated_month
	assert.NoError(t, recurringRepo.UpdateLastGeneratedMonth(ctx, "user-123", id, "2026-08"))
	assert.NoError(t, recurringRepo.Update(ctx, &biz.LedgerRecurring{
		Id: int64(id), UserId: "user-123", AccountId: 2, CategoryId: 6, Type: "income",
		Amount: decimal.NewFromInt(500), Remark: "工资", DayOfMonth: 15, StartMonth: "2026-09", Enabled: false,
	}))
	rule, err = recurringRepo.FindByUserIdAndId(ctx, "user-123", id)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), rule.AccountId)
	assert.Equal(t, "income", rule.Type)
	assert.False(t, rule.Enabled)
	assert.Equal(t, "2026-08", rule.LastGeneratedMonth)

	// 更新/删除仅限本人
	assert.Error(t, recurringRepo.Update(ctx, &biz.LedgerRecurring{
		Id: int64(id), UserId: "user-456", AccountId: 2, CategoryId: 6, Type: "income",
		Amount: decimal.NewFromInt(500), DayOfMonth: 15, StartMonth: "2026-09",
	}))
	assert.Error(t, recurringRepo.DeleteByUserIdAndId(ctx, "user-456", id))
	assert.NoError(t, recurringRepo.DeleteByUserIdAndId(ctx, "user-123", id))
	_, err = recurringRepo.FindByUserIdAndId(ctx, "user-123", id)
	assert.Error(t, err)
}

func TestLedgerRecurringRepo_ListEnabledFiltersDisabled(t *testing.T) {
	db := setupLedgerBudgetTestDB(t)
	helper := log.NewHelper(log.DefaultLogger)
	recurringRepo := &LedgerRecurringRepo{data: &Data{db: db}, log: helper}
	ctx := context.Background()

	_, err := recurringRepo.Save(ctx, &biz.LedgerRecurring{
		UserId: "user-123", AccountId: 1, CategoryId: 5, Type: "expense",
		Amount: decimal.NewFromInt(100), DayOfMonth: 1, StartMonth: "2026-08", Enabled: true,
	})
	assert.NoError(t, err)
	_, err = recurringRepo.Save(ctx, &biz.LedgerRecurring{
		UserId: "user-123", AccountId: 1, CategoryId: 7, Type: "income",
		Amount: decimal.NewFromInt(500), DayOfMonth: 15, StartMonth: "2026-08", Enabled: false,
	})
	assert.NoError(t, err)

	all, err := recurringRepo.ListByUserId(ctx, "user-123")
	assert.NoError(t, err)
	assert.Len(t, all, 2)

	enabled, err := recurringRepo.ListEnabledByUserId(ctx, "user-123")
	assert.NoError(t, err)
	assert.Len(t, enabled, 1)
	assert.Equal(t, "expense", enabled[0].Type)
}

func TestLedgerBudgetRepo_UpsertAndList(t *testing.T) {
	db := setupLedgerBudgetTestDB(t)
	helper := log.NewHelper(log.DefaultLogger)
	budgetRepo := &LedgerBudgetRepo{data: &Data{db: db}, log: helper}
	ctx := context.Background()

	// 新建
	id, err := budgetRepo.Upsert(ctx, &biz.LedgerBudget{UserId: "user-123", CategoryId: 5, Month: "2026-08", Amount: decimal.NewFromInt(300)})
	assert.NoError(t, err)

	// 同用户同分类同月：更新 amount，不产生新行
	id2, err := budgetRepo.Upsert(ctx, &biz.LedgerBudget{UserId: "user-123", CategoryId: 5, Month: "2026-08", Amount: decimal.NewFromInt(500)})
	assert.NoError(t, err)
	assert.Equal(t, id, id2)
	var count int64
	db.Model(&LedgerBudget{}).Where("user_id = ? AND category_id = ? AND month = ?", "user-123", 5, "2026-08").Count(&count)
	assert.Equal(t, int64(1), count)

	// 不同月/不同用户互不影响
	_, err = budgetRepo.Upsert(ctx, &biz.LedgerBudget{UserId: "user-123", CategoryId: 5, Month: "2026-09", Amount: decimal.NewFromInt(400)})
	assert.NoError(t, err)
	_, err = budgetRepo.Upsert(ctx, &biz.LedgerBudget{UserId: "user-456", CategoryId: 5, Month: "2026-08", Amount: decimal.NewFromInt(100)})
	assert.NoError(t, err)

	items, err := budgetRepo.ListByUserIdAndMonth(ctx, "user-123", "2026-08")
	assert.NoError(t, err)
	assert.Len(t, items, 1)
	assert.True(t, items[0].Amount.Equal(decimal.NewFromInt(500)))

	// 删除仅限本人
	assert.Error(t, budgetRepo.DeleteByUserIdAndId(ctx, "user-456", uint(id)))
	assert.NoError(t, budgetRepo.DeleteByUserIdAndId(ctx, "user-123", uint(id)))
}

func TestLedgerTransactionRepo_ListAccountLegsUpTo(t *testing.T) {
	db := setupLedgerBudgetTestDB(t)
	accountRepo, _, txRepo := newLedgerRepos(db)
	ctx := context.Background()

	cashId, err := accountRepo.Save(ctx, &biz.LedgerAccount{UserId: "user-123", Name: "现金", Type: "asset", Subtype: "cash"})
	assert.NoError(t, err)
	cardId, err := accountRepo.Save(ctx, &biz.LedgerAccount{UserId: "user-123", Name: "储蓄卡", Type: "asset", Subtype: "debit_card"})
	assert.NoError(t, err)

	// 8-01 现金 -100；8-02 现金 -50 / 卡 +50；8-05 现金 -10（endExclusive 之外应排除）
	_, err = txRepo.Save(ctx, &biz.LedgerTransaction{
		UserId: "user-123", Type: "expense", BookedAt: "2026-08-01 10:00:00",
		Postings: []*biz.LedgerPosting{
			{AccountId: int64(cashId), Amount: decimal.NewFromInt(-100)},
			{AccountId: 999, Amount: decimal.NewFromInt(100)},
		},
	})
	assert.NoError(t, err)
	tx2, err := txRepo.Save(ctx, &biz.LedgerTransaction{
		UserId: "user-123", Type: "transfer", BookedAt: "2026-08-02 10:00:00",
		Postings: []*biz.LedgerPosting{
			{AccountId: int64(cashId), Amount: decimal.NewFromInt(-50)},
			{AccountId: int64(cardId), Amount: decimal.NewFromInt(50)},
		},
	})
	assert.NoError(t, err)
	_, err = txRepo.Save(ctx, &biz.LedgerTransaction{
		UserId: "user-123", Type: "expense", BookedAt: "2026-08-05 10:00:00",
		Postings: []*biz.LedgerPosting{
			{AccountId: int64(cashId), Amount: decimal.NewFromInt(-10)},
			{AccountId: 999, Amount: decimal.NewFromInt(10)},
		},
	})
	assert.NoError(t, err)

	end, _ := time.ParseInLocation("2006-01-02", "2026-08-03", time.Local)
	legs, err := txRepo.ListAccountLegsUpTo(ctx, []uint{cashId}, end)

	assert.NoError(t, err)
	assert.Len(t, legs, 2)
	assert.Equal(t, "2026-08-01 10:00:00", legs[0].BookedAt)
	assert.True(t, legs[0].Amount.Equal(decimal.NewFromInt(-100)))
	assert.Equal(t, "2026-08-02 10:00:00", legs[1].BookedAt)

	// 软删交易的分录腿不计入走势
	assert.NoError(t, txRepo.DeleteByUserIdAndId(ctx, "user-123", tx2))
	legs, err = txRepo.ListAccountLegsUpTo(ctx, []uint{cashId}, end)
	assert.NoError(t, err)
	assert.Len(t, legs, 1)
}
