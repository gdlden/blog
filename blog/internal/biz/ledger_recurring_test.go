package biz

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

type mockLedgerRecurringRepo struct {
	saveFunc        func(context.Context, *LedgerRecurring) (uint, error)
	updateFunc      func(context.Context, *LedgerRecurring) error
	deleteFunc      func(context.Context, string, uint) error
	findByUserFunc  func(context.Context, string, uint) (*LedgerRecurring, error)
	listFunc        func(context.Context, string) ([]*LedgerRecurring, error)
	listEnabledFunc func(context.Context, string) ([]*LedgerRecurring, error)
	updateMonthFunc func(context.Context, string, uint, string) error
}

func (m *mockLedgerRecurringRepo) Save(ctx context.Context, rule *LedgerRecurring) (uint, error) {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, rule)
	}
	return 0, nil
}

func (m *mockLedgerRecurringRepo) Update(ctx context.Context, rule *LedgerRecurring) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, rule)
	}
	return nil
}

func (m *mockLedgerRecurringRepo) DeleteByUserIdAndId(ctx context.Context, userId string, id uint) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, userId, id)
	}
	return nil
}

func (m *mockLedgerRecurringRepo) FindByUserIdAndId(ctx context.Context, userId string, id uint) (*LedgerRecurring, error) {
	if m.findByUserFunc != nil {
		return m.findByUserFunc(ctx, userId, id)
	}
	return nil, errors.New("not found")
}

func (m *mockLedgerRecurringRepo) ListByUserId(ctx context.Context, userId string) ([]*LedgerRecurring, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, userId)
	}
	return nil, nil
}

func (m *mockLedgerRecurringRepo) ListEnabledByUserId(ctx context.Context, userId string) ([]*LedgerRecurring, error) {
	if m.listEnabledFunc != nil {
		return m.listEnabledFunc(ctx, userId)
	}
	return nil, nil
}

func (m *mockLedgerRecurringRepo) UpdateLastGeneratedMonth(ctx context.Context, userId string, id uint, month string) error {
	if m.updateMonthFunc != nil {
		return m.updateMonthFunc(ctx, userId, id, month)
	}
	return nil
}

// recurringRuleFixture 返回一条合法支出规则：账户 1、分类 5（expense）、金额 100、每月 1 日。
func recurringRuleFixture(startMonth string) *LedgerRecurring {
	return &LedgerRecurring{
		AccountId: 1, CategoryId: 5, Type: LedgerTxTypeExpense,
		Amount: decimal.NewFromInt(100), Remark: "房租",
		DayOfMonth: 1, StartMonth: startMonth, Enabled: true,
	}
}

func TestLedgerUsecase_SaveRecurringValidates(t *testing.T) {
	ctx := withUser(context.Background(), "user-123")
	uc := NewLedgerUsecase(ledgerAccountRepoFixture(), ledgerCategoryRepoFixture(), &mockLedgerTransactionRepo{}, &mockLedgerBudgetRepo{}, &mockLedgerRecurringRepo{}, log.DefaultLogger)

	// dayOfMonth 边界：0 / 32 拒绝，1 / 31 放行
	for _, day := range []int32{0, 32, -1} {
		rule := recurringRuleFixture("2026-08")
		rule.DayOfMonth = day
		_, err := uc.SaveRecurring(ctx, rule)
		assert.Error(t, err, "dayOfMonth=%d", day)
		assert.Contains(t, err.Error(), "1-31")
	}
	for _, day := range []int32{1, 31} {
		rule := recurringRuleFixture("2026-08")
		rule.DayOfMonth = day
		_, err := uc.SaveRecurring(ctx, rule)
		assert.NoError(t, err, "dayOfMonth=%d", day)
	}

	// type=transfer 拒绝
	rule := recurringRuleFixture("2026-08")
	rule.Type = LedgerTxTypeTransfer
	_, err := uc.SaveRecurring(ctx, rule)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "仅支持 expense 或 income")

	// amount <= 0 拒绝
	rule = recurringRuleFixture("2026-08")
	rule.Amount = decimal.Zero
	_, err = uc.SaveRecurring(ctx, rule)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "必须大于 0")

	// startMonth 格式错误
	rule = recurringRuleFixture("2026/08")
	_, err = uc.SaveRecurring(ctx, rule)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "YYYY-MM")

	// 账户不存在
	rule = recurringRuleFixture("2026-08")
	rule.AccountId = 999
	_, err = uc.SaveRecurring(ctx, rule)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "账户不存在")

	// 分类不存在 / 方向不匹配（expense 规则挂 income 分类 7）
	rule = recurringRuleFixture("2026-08")
	rule.CategoryId = 999
	_, err = uc.SaveRecurring(ctx, rule)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "分类不存在")
	rule = recurringRuleFixture("2026-08")
	rule.CategoryId = 7
	_, err = uc.SaveRecurring(ctx, rule)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "方向与周期账单类型不匹配")
}

func TestLedgerUsecase_SaveRecurringUpdateKeepsProgress(t *testing.T) {
	ctx := withUser(context.Background(), "user-123")
	var updated *LedgerRecurring
	recurringRepo := &mockLedgerRecurringRepo{
		findByUserFunc: func(ctx context.Context, userId string, id uint) (*LedgerRecurring, error) {
			if id != 3 {
				return nil, errors.New("not found")
			}
			return &LedgerRecurring{Id: int64(id), UserId: userId, LastGeneratedMonth: "2026-07"}, nil
		},
		updateFunc: func(ctx context.Context, rule *LedgerRecurring) error {
			updated = rule
			return nil
		},
	}
	uc := NewLedgerUsecase(ledgerAccountRepoFixture(), ledgerCategoryRepoFixture(), &mockLedgerTransactionRepo{}, &mockLedgerBudgetRepo{}, recurringRepo, log.DefaultLogger)

	rule := recurringRuleFixture("2026-05")
	rule.Id = 3
	id, err := uc.SaveRecurring(ctx, rule)

	assert.NoError(t, err)
	assert.Equal(t, uint(3), id)
	assert.NotNil(t, updated)
	// 更新不回退生成进度
	assert.Equal(t, "2026-07", updated.LastGeneratedMonth)

	// 更新他人规则 → 拒绝
	_, err = uc.SaveRecurring(ctx, &LedgerRecurring{Id: 9, AccountId: 1, CategoryId: 5, Type: LedgerTxTypeExpense, Amount: decimal.NewFromInt(1), DayOfMonth: 1, StartMonth: "2026-08"})
	assert.Error(t, err)
}

// applyRecurringFixture 构造带状态的 mock：apply 逐月推进 lastGeneratedMonth，txRepo 记录生成的交易。
func applyRecurringFixture(rules []*LedgerRecurring) (*mockLedgerRecurringRepo, *mockLedgerTransactionRepo) {
	recurringRepo := &mockLedgerRecurringRepo{
		listEnabledFunc: func(ctx context.Context, userId string) ([]*LedgerRecurring, error) {
			enabled := make([]*LedgerRecurring, 0, len(rules))
			for _, rule := range rules {
				if rule.Enabled {
					enabled = append(enabled, rule)
				}
			}
			return enabled, nil
		},
		updateMonthFunc: func(ctx context.Context, userId string, id uint, month string) error {
			for _, rule := range rules {
				if uint(rule.Id) == id {
					rule.LastGeneratedMonth = month
				}
			}
			return nil
		},
	}
	txRepo := &mockLedgerTransactionRepo{
		saveFunc: func(ctx context.Context, tx *LedgerTransaction) (uint, error) {
			return 1, nil
		},
	}
	return recurringRepo, txRepo
}

func TestLedgerUsecase_ApplyRecurringDoubleEntryLegs(t *testing.T) {
	ctx := withUser(context.Background(), "user-123")
	thisMonth := time.Now().Format("2006-01")

	var savedTxs []*LedgerTransaction
	expenseRule := recurringRuleFixture(thisMonth)
	expenseRule.Id = 1
	incomeRule := &LedgerRecurring{
		Id: 2, AccountId: 1, CategoryId: 7, Type: LedgerTxTypeIncome,
		Amount: decimal.NewFromInt(500), Remark: "工资",
		DayOfMonth: 1, StartMonth: thisMonth, Enabled: true,
	}
	recurringRepo, txRepo := applyRecurringFixture([]*LedgerRecurring{expenseRule, incomeRule})
	txRepo.saveFunc = func(ctx context.Context, tx *LedgerTransaction) (uint, error) {
		savedTxs = append(savedTxs, tx)
		return 1, nil
	}
	uc := NewLedgerUsecase(ledgerAccountRepoFixture(), ledgerCategoryRepoFixture(), txRepo, &mockLedgerBudgetRepo{}, recurringRepo, log.DefaultLogger)

	created, err := uc.ApplyRecurring(ctx)

	assert.NoError(t, err)
	assert.Equal(t, int32(2), created)
	assert.Len(t, savedTxs, 2)
	due := time.Now().Format("2006-01") + "-01 00:00:00"
	// expense：账户腿（负）+ 费用系统腿（正，挂分类），accountId=0 已解析为系统账户 100
	expenseTx := savedTxs[0]
	assert.Equal(t, LedgerTxTypeExpense, expenseTx.Type)
	assert.Equal(t, due, expenseTx.BookedAt)
	assert.Equal(t, "房租", expenseTx.Remark)
	assert.Len(t, expenseTx.Postings, 2)
	assert.Equal(t, int64(1), expenseTx.Postings[0].AccountId)
	assert.True(t, expenseTx.Postings[0].Amount.Equal(decimal.NewFromInt(-100)))
	assert.Equal(t, int64(100), expenseTx.Postings[1].AccountId)
	assert.True(t, expenseTx.Postings[1].Amount.Equal(decimal.NewFromInt(100)))
	assert.Equal(t, int64(5), expenseTx.Postings[1].CategoryId)
	// income：账户腿（正）+ 收入系统腿（负，挂分类），系统账户 101
	incomeTx := savedTxs[1]
	assert.Equal(t, LedgerTxTypeIncome, incomeTx.Type)
	assert.True(t, incomeTx.Postings[0].Amount.Equal(decimal.NewFromInt(500)))
	assert.Equal(t, int64(101), incomeTx.Postings[1].AccountId)
	assert.True(t, incomeTx.Postings[1].Amount.Equal(decimal.NewFromInt(-500)))
	assert.Equal(t, int64(7), incomeTx.Postings[1].CategoryId)
}

func TestLedgerUsecase_ApplyRecurringIdempotent(t *testing.T) {
	ctx := withUser(context.Background(), "user-123")
	// 两个月前起、每月 1 日：首次 apply 生成 3 笔（含当月），第二次 apply created=0
	rule := recurringRuleFixture(time.Now().AddDate(0, -2, 0).Format("2006-01"))
	rule.Id = 1
	recurringRepo, txRepo := applyRecurringFixture([]*LedgerRecurring{rule})
	uc := NewLedgerUsecase(ledgerAccountRepoFixture(), ledgerCategoryRepoFixture(), txRepo, &mockLedgerBudgetRepo{}, recurringRepo, log.DefaultLogger)

	created, err := uc.ApplyRecurring(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int32(3), created)
	assert.Equal(t, time.Now().Format("2006-01"), rule.LastGeneratedMonth)

	created, err = uc.ApplyRecurring(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int32(0), created)
}

func TestLedgerUsecase_ApplyRecurringBackfillsMonths(t *testing.T) {
	ctx := withUser(context.Background(), "user-123")
	var bookedAts []string
	rule := recurringRuleFixture(time.Now().AddDate(0, -3, 0).Format("2006-01"))
	rule.Id = 1
	recurringRepo, txRepo := applyRecurringFixture([]*LedgerRecurring{rule})
	txRepo.saveFunc = func(ctx context.Context, tx *LedgerTransaction) (uint, error) {
		bookedAts = append(bookedAts, tx.BookedAt)
		return 1, nil
	}
	uc := NewLedgerUsecase(ledgerAccountRepoFixture(), ledgerCategoryRepoFixture(), txRepo, &mockLedgerBudgetRepo{}, recurringRepo, log.DefaultLogger)

	created, err := uc.ApplyRecurring(ctx)

	assert.NoError(t, err)
	// 跨多月补齐：3 个月前至当月共 4 笔
	assert.Equal(t, int32(4), created)
	assert.Len(t, bookedAts, 4)
	for i, bookedAt := range bookedAts {
		expected := time.Now().AddDate(0, -3+i, 0).Format("2006-01") + "-01 00:00:00"
		assert.Equal(t, expected, bookedAt)
	}
}

func TestLedgerUsecase_ApplyRecurringClampsShortMonth(t *testing.T) {
	ctx := withUser(context.Background(), "user-123")
	var bookedAts []string
	// 2024-02 为闰月：dayOfMonth=30 钳制到 29 日；3 月恢复 30 日
	rule := recurringRuleFixture("2024-02")
	rule.Id = 1
	rule.DayOfMonth = 30
	recurringRepo, txRepo := applyRecurringFixture([]*LedgerRecurring{rule})
	txRepo.saveFunc = func(ctx context.Context, tx *LedgerTransaction) (uint, error) {
		bookedAts = append(bookedAts, tx.BookedAt)
		return 1, nil
	}
	uc := NewLedgerUsecase(ledgerAccountRepoFixture(), ledgerCategoryRepoFixture(), txRepo, &mockLedgerBudgetRepo{}, recurringRepo, log.DefaultLogger)

	created, err := uc.ApplyRecurring(ctx)

	assert.NoError(t, err)
	assert.GreaterOrEqual(t, created, int32(2))
	assert.Equal(t, "2024-02-29 00:00:00", bookedAts[0])
	assert.Equal(t, "2024-03-30 00:00:00", bookedAts[1])
}

func TestLedgerUsecase_ApplyRecurringSkipsDisabledAndFuture(t *testing.T) {
	ctx := withUser(context.Background(), "user-123")
	enabled := recurringRuleFixture(time.Now().AddDate(0, -1, 0).Format("2006-01"))
	enabled.Id = 1
	disabled := recurringRuleFixture(time.Now().AddDate(0, -1, 0).Format("2006-01"))
	disabled.Id = 2
	disabled.Enabled = false
	future := recurringRuleFixture(time.Now().AddDate(0, 1, 0).Format("2006-01"))
	future.Id = 3
	recurringRepo, txRepo := applyRecurringFixture([]*LedgerRecurring{enabled, disabled, future})
	savedCount := 0
	txRepo.saveFunc = func(ctx context.Context, tx *LedgerTransaction) (uint, error) {
		savedCount++
		return 1, nil
	}
	uc := NewLedgerUsecase(ledgerAccountRepoFixture(), ledgerCategoryRepoFixture(), txRepo, &mockLedgerBudgetRepo{}, recurringRepo, log.DefaultLogger)

	created, err := uc.ApplyRecurring(ctx)

	assert.NoError(t, err)
	// disabled 规则不在启用列表中不生成；future 规则应生成日 > 今天不生成；仅 enabled 补 2 个月
	assert.Equal(t, int32(2), created)
	assert.Equal(t, 2, savedCount)
	assert.Equal(t, "", disabled.LastGeneratedMonth)
	assert.Equal(t, "", future.LastGeneratedMonth)
}

func TestLedgerUsecase_ListRecurringFillsNamesAndNextDate(t *testing.T) {
	ctx := withUser(context.Background(), "user-123")
	recurringRepo := &mockLedgerRecurringRepo{
		listFunc: func(ctx context.Context, userId string) ([]*LedgerRecurring, error) {
			return []*LedgerRecurring{
				{Id: 1, AccountId: 1, CategoryId: 5, Type: LedgerTxTypeExpense, Amount: decimal.NewFromInt(100), DayOfMonth: 31, StartMonth: "2026-01", Enabled: true},
				{Id: 2, AccountId: 2, CategoryId: 7, Type: LedgerTxTypeIncome, Amount: decimal.NewFromInt(500), DayOfMonth: 30, StartMonth: "2026-01", LastGeneratedMonth: "2026-03", Enabled: true},
				{Id: 3, AccountId: 1, CategoryId: 5, Type: LedgerTxTypeExpense, Amount: decimal.NewFromInt(100), DayOfMonth: 15, StartMonth: "2026-01", Enabled: false},
			}, nil
		},
	}
	accountRepo := ledgerAccountRepoFixture()
	accountRepo.listFunc = func(ctx context.Context, userId string) ([]*LedgerAccount, error) {
		return []*LedgerAccount{
			{Id: 1, UserId: userId, Name: "现金"},
			{Id: 2, UserId: userId, Name: "储蓄卡"},
		}, nil
	}
	categoryRepo := ledgerCategoryRepoFixture()
	categoryRepo.listFunc = func(ctx context.Context, userId string) ([]*LedgerCategory, error) {
		return []*LedgerCategory{
			{Id: 5, Name: "餐饮", Direction: LedgerDirectionExpense},
			{Id: 7, Name: "工资", Direction: LedgerDirectionIncome},
		}, nil
	}
	uc := NewLedgerUsecase(accountRepo, categoryRepo, &mockLedgerTransactionRepo{}, &mockLedgerBudgetRepo{}, recurringRepo, log.DefaultLogger)

	rules, err := uc.ListRecurring(ctx)

	assert.NoError(t, err)
	assert.Len(t, rules, 3)
	// 未生成过：nextDate = startMonth 应生成日
	assert.Equal(t, "现金", rules[0].AccountName)
	assert.Equal(t, "餐饮", rules[0].CategoryName)
	assert.Equal(t, "2026-01-31", rules[0].NextDate)
	// 已生成到 2026-03：nextDate = 下一月应生成日
	assert.Equal(t, "储蓄卡", rules[1].AccountName)
	assert.Equal(t, "工资", rules[1].CategoryName)
	assert.Equal(t, "2026-04-30", rules[1].NextDate)
	// disabled：nextDate 为空
	assert.Equal(t, "", rules[2].NextDate)
}
