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

type mockLedgerAccountRepo struct {
	saveFunc       func(context.Context, *LedgerAccount) (uint, error)
	findByUserFunc func(context.Context, string, uint) (*LedgerAccount, error)
	findSystemFunc func(context.Context, string, string) (*LedgerAccount, error)
	listFunc       func(context.Context, string) ([]*LedgerAccount, error)
	ensureFunc     func(context.Context, string) error
	sumFunc        func(context.Context, []uint) (map[uint]decimal.Decimal, error)
	countFunc      func(context.Context, uint) (int64, error)
}

func (m *mockLedgerAccountRepo) Save(ctx context.Context, account *LedgerAccount) (uint, error) {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, account)
	}
	return 0, nil
}

func (m *mockLedgerAccountRepo) Update(ctx context.Context, account *LedgerAccount) error {
	return nil
}

func (m *mockLedgerAccountRepo) DeleteByUserIdAndId(ctx context.Context, userId string, id uint) error {
	return nil
}

func (m *mockLedgerAccountRepo) FindByUserIdAndId(ctx context.Context, userId string, id uint) (*LedgerAccount, error) {
	if m.findByUserFunc != nil {
		return m.findByUserFunc(ctx, userId, id)
	}
	return nil, errors.New("not found")
}

func (m *mockLedgerAccountRepo) FindSystemByUserIdAndType(ctx context.Context, userId string, accountType string) (*LedgerAccount, error) {
	if m.findSystemFunc != nil {
		return m.findSystemFunc(ctx, userId, accountType)
	}
	return nil, errors.New("not found")
}

func (m *mockLedgerAccountRepo) ListByUserId(ctx context.Context, userId string) ([]*LedgerAccount, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, userId)
	}
	return nil, nil
}

func (m *mockLedgerAccountRepo) EnsureSystemAccounts(ctx context.Context, userId string) error {
	if m.ensureFunc != nil {
		return m.ensureFunc(ctx, userId)
	}
	return nil
}

func (m *mockLedgerAccountRepo) SumBalanceByAccountIds(ctx context.Context, ids []uint) (map[uint]decimal.Decimal, error) {
	if m.sumFunc != nil {
		return m.sumFunc(ctx, ids)
	}
	return map[uint]decimal.Decimal{}, nil
}

func (m *mockLedgerAccountRepo) CountPostingsByAccountId(ctx context.Context, accountId uint) (int64, error) {
	if m.countFunc != nil {
		return m.countFunc(ctx, accountId)
	}
	return 0, nil
}

type mockLedgerCategoryRepo struct {
	saveFunc           func(context.Context, *LedgerCategory) (uint, error)
	findByUserFunc     func(context.Context, string, uint) (*LedgerCategory, error)
	listFunc           func(context.Context, string) ([]*LedgerCategory, error)
	hasChildrenFunc    func(context.Context, uint) (bool, error)
	findFallbackFunc   func(context.Context, string, string) (*LedgerCategory, error)
	deleteWithMigrFunc func(context.Context, string, uint, uint) error
}

func (m *mockLedgerCategoryRepo) Save(ctx context.Context, category *LedgerCategory) (uint, error) {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, category)
	}
	return 0, nil
}

func (m *mockLedgerCategoryRepo) Update(ctx context.Context, category *LedgerCategory) error {
	return nil
}

func (m *mockLedgerCategoryRepo) FindByUserIdAndId(ctx context.Context, userId string, id uint) (*LedgerCategory, error) {
	if m.findByUserFunc != nil {
		return m.findByUserFunc(ctx, userId, id)
	}
	return nil, errors.New("not found")
}

func (m *mockLedgerCategoryRepo) ListByUserId(ctx context.Context, userId string) ([]*LedgerCategory, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, userId)
	}
	return nil, nil
}

func (m *mockLedgerCategoryRepo) EnsurePresetCategories(ctx context.Context, userId string) error {
	return nil
}

func (m *mockLedgerCategoryRepo) HasChildren(ctx context.Context, id uint) (bool, error) {
	if m.hasChildrenFunc != nil {
		return m.hasChildrenFunc(ctx, id)
	}
	return false, nil
}

func (m *mockLedgerCategoryRepo) FindFallbackByDirection(ctx context.Context, userId string, direction string) (*LedgerCategory, error) {
	if m.findFallbackFunc != nil {
		return m.findFallbackFunc(ctx, userId, direction)
	}
	return nil, errors.New("not found")
}

func (m *mockLedgerCategoryRepo) DeleteWithMigration(ctx context.Context, userId string, id uint, fallbackId uint) error {
	if m.deleteWithMigrFunc != nil {
		return m.deleteWithMigrFunc(ctx, userId, id, fallbackId)
	}
	return nil
}

type mockLedgerBudgetRepo struct {
	upsertFunc func(context.Context, *LedgerBudget) (uint, error)
	listFunc   func(context.Context, string, string) ([]*LedgerBudget, error)
	deleteFunc func(context.Context, string, uint) error
}

func (m *mockLedgerBudgetRepo) Upsert(ctx context.Context, budget *LedgerBudget) (uint, error) {
	if m.upsertFunc != nil {
		return m.upsertFunc(ctx, budget)
	}
	return 0, nil
}

func (m *mockLedgerBudgetRepo) DeleteByUserIdAndId(ctx context.Context, userId string, id uint) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, userId, id)
	}
	return nil
}

func (m *mockLedgerBudgetRepo) ListByUserIdAndMonth(ctx context.Context, userId string, month string) ([]*LedgerBudget, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, userId, month)
	}
	return nil, nil
}

type mockLedgerTransactionRepo struct {
	saveFunc     func(context.Context, *LedgerTransaction) (uint, error)
	findFunc     func(context.Context, string, uint) (*LedgerTransaction, error)
	legsFunc     func(context.Context, string, string, uint, string) ([]*LedgerPosting, error)
	legsCalls    []string
	legsUpToFunc func(context.Context, []uint, time.Time) ([]*LedgerAccountLeg, error)
	deleteFunc   func(context.Context, string, uint) error
	deletedId    uint
	updateFunc   func(context.Context, *LedgerTransaction) error
	updatedTx    *LedgerTransaction
}

func (m *mockLedgerTransactionRepo) Save(ctx context.Context, tx *LedgerTransaction) (uint, error) {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, tx)
	}
	return 0, nil
}

func (m *mockLedgerTransactionRepo) Update(ctx context.Context, tx *LedgerTransaction) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, tx)
	}
	return nil
}

func (m *mockLedgerTransactionRepo) DeleteByUserIdAndId(ctx context.Context, userId string, id uint) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, userId, id)
	}
	return nil
}

func (m *mockLedgerTransactionRepo) FindByUserIdAndId(ctx context.Context, userId string, id uint) (*LedgerTransaction, error) {
	if m.findFunc != nil {
		return m.findFunc(ctx, userId, id)
	}
	return nil, errors.New("not found")
}

func (m *mockLedgerTransactionRepo) PageByUserId(ctx context.Context, userId string, query *LedgerTransactionQuery) ([]*LedgerTransaction, int64, error) {
	return nil, 0, nil
}

func (m *mockLedgerTransactionRepo) ListSystemLegsByMonth(ctx context.Context, userId string, txType string, systemAccountId uint, month string) ([]*LedgerPosting, error) {
	m.legsCalls = append(m.legsCalls, txType)
	if m.legsFunc != nil {
		return m.legsFunc(ctx, userId, txType, systemAccountId, month)
	}
	return nil, nil
}

func (m *mockLedgerTransactionRepo) ListAccountLegsUpTo(ctx context.Context, accountIds []uint, endExclusive time.Time) ([]*LedgerAccountLeg, error) {
	if m.legsUpToFunc != nil {
		return m.legsUpToFunc(ctx, accountIds, endExclusive)
	}
	return nil, nil
}

// 常用 mock：账户 1/2 为普通资产账户，100/101/102 为系统账户（费用/收入/期初调整）。
func ledgerAccountRepoFixture() *mockLedgerAccountRepo {
	return &mockLedgerAccountRepo{
		findByUserFunc: func(ctx context.Context, userId string, id uint) (*LedgerAccount, error) {
			switch id {
			case 1, 2:
				return &LedgerAccount{Id: int64(id), UserId: userId, Type: LedgerAccountTypeAsset, Subtype: "cash"}, nil
			case 100:
				return &LedgerAccount{Id: 100, UserId: userId, Type: LedgerAccountTypeExpense, IsSystem: true}, nil
			case 101:
				return &LedgerAccount{Id: 101, UserId: userId, Type: LedgerAccountTypeIncome, IsSystem: true}, nil
			case 102:
				return &LedgerAccount{Id: 102, UserId: userId, Type: LedgerAccountTypeEquity, IsSystem: true}, nil
			}
			return nil, errors.New("not found")
		},
		findSystemFunc: func(ctx context.Context, userId string, accountType string) (*LedgerAccount, error) {
			switch accountType {
			case LedgerAccountTypeExpense:
				return &LedgerAccount{Id: 100, UserId: userId, Type: accountType, IsSystem: true}, nil
			case LedgerAccountTypeIncome:
				return &LedgerAccount{Id: 101, UserId: userId, Type: accountType, IsSystem: true}, nil
			case LedgerAccountTypeEquity:
				return &LedgerAccount{Id: 102, UserId: userId, Type: accountType, IsSystem: true}, nil
			}
			return nil, errors.New("not found")
		},
	}
}

func ledgerCategoryRepoFixture() *mockLedgerCategoryRepo {
	return &mockLedgerCategoryRepo{
		findByUserFunc: func(ctx context.Context, userId string, id uint) (*LedgerCategory, error) {
			switch id {
			case 5:
				return &LedgerCategory{Id: 5, UserId: userId, Name: "餐饮", Direction: LedgerDirectionExpense}, nil
			case 7:
				return &LedgerCategory{Id: 7, UserId: userId, Name: "工资", Direction: LedgerDirectionIncome}, nil
			}
			return nil, errors.New("not found")
		},
	}
}

func TestLedgerUsecase_CreateTransactionRejectsUnbalanced(t *testing.T) {
	uc := NewLedgerUsecase(ledgerAccountRepoFixture(), ledgerCategoryRepoFixture(), &mockLedgerTransactionRepo{}, &mockLedgerBudgetRepo{}, &mockLedgerRecurringRepo{}, log.DefaultLogger)
	tx := &LedgerTransaction{
		Type:     LedgerTxTypeTransfer,
		BookedAt: "2026-08-01 10:00:00",
		Postings: []*LedgerPosting{
			{AccountId: 1, Amount: decimal.NewFromInt(-100)},
			{AccountId: 2, Amount: decimal.NewFromInt(90)},
		},
	}

	_, err := uc.CreateTransaction(withUser(context.Background(), "user-123"), tx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "合计必须为 0")
}

func TestLedgerUsecase_CreateTransactionRejectsSingleLegAndZeroAmount(t *testing.T) {
	uc := NewLedgerUsecase(ledgerAccountRepoFixture(), ledgerCategoryRepoFixture(), &mockLedgerTransactionRepo{}, &mockLedgerBudgetRepo{}, &mockLedgerRecurringRepo{}, log.DefaultLogger)

	// 单腿拒绝
	tx := &LedgerTransaction{
		Type:     LedgerTxTypeTransfer,
		BookedAt: "2026-08-01 10:00:00",
		Postings: []*LedgerPosting{{AccountId: 1, Amount: decimal.NewFromInt(-100)}},
	}
	_, err := uc.CreateTransaction(withUser(context.Background(), "user-123"), tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "至少需要两条")

	// amount=0 拒绝
	tx = &LedgerTransaction{
		Type:     LedgerTxTypeTransfer,
		BookedAt: "2026-08-01 10:00:00",
		Postings: []*LedgerPosting{
			{AccountId: 1, Amount: decimal.NewFromInt(-100)},
			{AccountId: 2, Amount: decimal.Zero},
		},
	}
	_, err = uc.CreateTransaction(withUser(context.Background(), "user-123"), tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "不能为 0")
}

func TestLedgerUsecase_CreateTransactionResolvesSystemLeg(t *testing.T) {
	var saved *LedgerTransaction
	txRepo := &mockLedgerTransactionRepo{
		saveFunc: func(ctx context.Context, tx *LedgerTransaction) (uint, error) {
			saved = tx
			return 1, nil
		},
	}
	uc := NewLedgerUsecase(ledgerAccountRepoFixture(), ledgerCategoryRepoFixture(), txRepo, &mockLedgerBudgetRepo{}, &mockLedgerRecurringRepo{}, log.DefaultLogger)

	// expense：accountId=0 解析为系统费用账户 100
	tx := &LedgerTransaction{
		Type:     LedgerTxTypeExpense,
		BookedAt: "2026-08-01 10:00:00",
		Postings: []*LedgerPosting{
			{AccountId: 0, Amount: decimal.NewFromInt(100), CategoryId: 5},
			{AccountId: 1, Amount: decimal.NewFromInt(-100)},
		},
	}
	_, err := uc.CreateTransaction(withUser(context.Background(), "user-123"), tx)
	assert.NoError(t, err)
	assert.NotNil(t, saved)
	assert.Equal(t, int64(100), saved.Postings[0].AccountId)

	// income：accountId=0 解析为系统收入账户 101
	saved = nil
	tx = &LedgerTransaction{
		Type:     LedgerTxTypeIncome,
		BookedAt: "2026-08-01 10:00:00",
		Postings: []*LedgerPosting{
			{AccountId: 1, Amount: decimal.NewFromInt(200)},
			{AccountId: 0, Amount: decimal.NewFromInt(-200), CategoryId: 7},
		},
	}
	_, err = uc.CreateTransaction(withUser(context.Background(), "user-123"), tx)
	assert.NoError(t, err)
	assert.NotNil(t, saved)
	assert.Equal(t, int64(101), saved.Postings[1].AccountId)

	// transfer：accountId=0 报中文错误
	tx = &LedgerTransaction{
		Type:     LedgerTxTypeTransfer,
		BookedAt: "2026-08-01 10:00:00",
		Postings: []*LedgerPosting{
			{AccountId: 1, Amount: decimal.NewFromInt(-100)},
			{AccountId: 0, Amount: decimal.NewFromInt(100)},
		},
	}
	_, err = uc.CreateTransaction(withUser(context.Background(), "user-123"), tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "不支持系统对手账户")
}

func TestLedgerUsecase_CreateTransactionMultiLegSplit(t *testing.T) {
	var saved *LedgerTransaction
	txRepo := &mockLedgerTransactionRepo{
		saveFunc: func(ctx context.Context, tx *LedgerTransaction) (uint, error) {
			saved = tx
			return 1, nil
		},
	}
	uc := NewLedgerUsecase(ledgerAccountRepoFixture(), ledgerCategoryRepoFixture(), txRepo, &mockLedgerBudgetRepo{}, &mockLedgerRecurringRepo{}, log.DefaultLogger)

	// 多腿拆分：消费 100，现金付 60、卡付 40，两条腿挂在系统费用账户 +100 上拆分两条
	tx := &LedgerTransaction{
		Type:     LedgerTxTypeExpense,
		BookedAt: "2026-08-01 10:00:00",
		Postings: []*LedgerPosting{
			{AccountId: 1, Amount: decimal.NewFromInt(-60)},
			{AccountId: 2, Amount: decimal.NewFromInt(-40)},
			{AccountId: 0, Amount: decimal.NewFromInt(60), CategoryId: 5},
			{AccountId: 0, Amount: decimal.NewFromInt(40), CategoryId: 5},
		},
	}
	_, err := uc.CreateTransaction(withUser(context.Background(), "user-123"), tx)

	assert.NoError(t, err)
	assert.NotNil(t, saved)
	assert.Len(t, saved.Postings, 4)
	assert.Equal(t, int64(100), saved.Postings[2].AccountId)
	assert.Equal(t, int64(100), saved.Postings[3].AccountId)
}

func TestLedgerUsecase_CreateTransactionValidatesCategoryDirection(t *testing.T) {
	uc := NewLedgerUsecase(ledgerAccountRepoFixture(), ledgerCategoryRepoFixture(), &mockLedgerTransactionRepo{}, &mockLedgerBudgetRepo{}, &mockLedgerRecurringRepo{}, log.DefaultLogger)

	// expense 交易挂 income 分类 → 拒绝
	tx := &LedgerTransaction{
		Type:     LedgerTxTypeExpense,
		BookedAt: "2026-08-01 10:00:00",
		Postings: []*LedgerPosting{
			{AccountId: 0, Amount: decimal.NewFromInt(100), CategoryId: 7},
			{AccountId: 1, Amount: decimal.NewFromInt(-100)},
		},
	}
	_, err := uc.CreateTransaction(withUser(context.Background(), "user-123"), tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "方向与交易类型不匹配")

	// transfer 挂分类 → 拒绝
	tx = &LedgerTransaction{
		Type:     LedgerTxTypeTransfer,
		BookedAt: "2026-08-01 10:00:00",
		Postings: []*LedgerPosting{
			{AccountId: 1, Amount: decimal.NewFromInt(-100), CategoryId: 5},
			{AccountId: 2, Amount: decimal.NewFromInt(100)},
		},
	}
	_, err = uc.CreateTransaction(withUser(context.Background(), "user-123"), tx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "不能挂载分类")
}

func TestLedgerUsecase_DeleteCategoryMigratesToFallback(t *testing.T) {
	var migratedFrom, migratedTo uint
	categoryRepo := &mockLedgerCategoryRepo{
		findByUserFunc: func(ctx context.Context, userId string, id uint) (*LedgerCategory, error) {
			return &LedgerCategory{Id: int64(id), UserId: userId, Name: "数码", Direction: LedgerDirectionExpense}, nil
		},
		findFallbackFunc: func(ctx context.Context, userId string, direction string) (*LedgerCategory, error) {
			assert.Equal(t, LedgerDirectionExpense, direction)
			return &LedgerCategory{Id: 99, UserId: userId, Name: "其它", Direction: direction, IsSystem: true}, nil
		},
		deleteWithMigrFunc: func(ctx context.Context, userId string, id uint, fallbackId uint) error {
			migratedFrom, migratedTo = id, fallbackId
			return nil
		},
	}
	uc := NewLedgerUsecase(ledgerAccountRepoFixture(), categoryRepo, &mockLedgerTransactionRepo{}, &mockLedgerBudgetRepo{}, &mockLedgerRecurringRepo{}, log.DefaultLogger)

	err := uc.DeleteCategory(withUser(context.Background(), "user-123"), 5)

	assert.NoError(t, err)
	assert.Equal(t, uint(5), migratedFrom)
	assert.Equal(t, uint(99), migratedTo)
}

func TestLedgerUsecase_DeleteCategoryRejectsSystemAndNonLeaf(t *testing.T) {
	// 系统分类不可删
	categoryRepo := &mockLedgerCategoryRepo{
		findByUserFunc: func(ctx context.Context, userId string, id uint) (*LedgerCategory, error) {
			return &LedgerCategory{Id: int64(id), UserId: userId, Name: "其它", Direction: LedgerDirectionExpense, IsSystem: true}, nil
		},
	}
	uc := NewLedgerUsecase(ledgerAccountRepoFixture(), categoryRepo, &mockLedgerTransactionRepo{}, &mockLedgerBudgetRepo{}, &mockLedgerRecurringRepo{}, log.DefaultLogger)
	err := uc.DeleteCategory(withUser(context.Background(), "user-123"), 99)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "系统分类不允许删除")

	// 非叶子分类不可删
	categoryRepo = &mockLedgerCategoryRepo{
		findByUserFunc: func(ctx context.Context, userId string, id uint) (*LedgerCategory, error) {
			return &LedgerCategory{Id: int64(id), UserId: userId, Name: "餐饮", Direction: LedgerDirectionExpense}, nil
		},
		hasChildrenFunc: func(ctx context.Context, id uint) (bool, error) {
			return true, nil
		},
	}
	uc = NewLedgerUsecase(ledgerAccountRepoFixture(), categoryRepo, &mockLedgerTransactionRepo{}, &mockLedgerBudgetRepo{}, &mockLedgerRecurringRepo{}, log.DefaultLogger)
	err = uc.DeleteCategory(withUser(context.Background(), "user-123"), 5)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "存在子分类")
}

func TestLedgerUsecase_DeleteAccountBlockedWhenPostingsExist(t *testing.T) {
	accountRepo := ledgerAccountRepoFixture()
	accountRepo.countFunc = func(ctx context.Context, accountId uint) (int64, error) {
		return 3, nil
	}
	uc := NewLedgerUsecase(accountRepo, ledgerCategoryRepoFixture(), &mockLedgerTransactionRepo{}, &mockLedgerBudgetRepo{}, &mockLedgerRecurringRepo{}, log.DefaultLogger)

	err := uc.DeleteAccount(withUser(context.Background(), "user-123"), 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "存在交易记录，禁止删除")
}

func TestLedgerUsecase_CreateAccountWithOpeningBalance(t *testing.T) {
	var savedTx *LedgerTransaction
	accountRepo := ledgerAccountRepoFixture()
	accountRepo.saveFunc = func(ctx context.Context, account *LedgerAccount) (uint, error) {
		return 42, nil
	}
	txRepo := &mockLedgerTransactionRepo{
		saveFunc: func(ctx context.Context, tx *LedgerTransaction) (uint, error) {
			savedTx = tx
			return 1, nil
		},
	}
	uc := NewLedgerUsecase(accountRepo, ledgerCategoryRepoFixture(), txRepo, &mockLedgerBudgetRepo{}, &mockLedgerRecurringRepo{}, log.DefaultLogger)
	opening := decimal.NewFromInt(1000)

	id, err := uc.CreateAccount(withUser(context.Background(), "user-123"), &LedgerAccount{
		Name: "招行储蓄卡", Type: LedgerAccountTypeAsset, Subtype: "debit_card",
	}, &opening)

	assert.NoError(t, err)
	assert.Equal(t, uint(42), id)
	// 开账交易：两腿 [新账户 +1000, 期初调整系统账户 -1000]
	assert.NotNil(t, savedTx)
	assert.Equal(t, LedgerTxTypeOpeningBalance, savedTx.Type)
	assert.Len(t, savedTx.Postings, 2)
	assert.Equal(t, int64(42), savedTx.Postings[0].AccountId)
	assert.True(t, savedTx.Postings[0].Amount.Equal(decimal.NewFromInt(1000)))
	assert.Equal(t, int64(102), savedTx.Postings[1].AccountId)
	assert.True(t, savedTx.Postings[1].Amount.Equal(decimal.NewFromInt(-1000)))

	// 不传期初余额：不产生开账交易
	savedTx = nil
	_, err = uc.CreateAccount(withUser(context.Background(), "user-123"), &LedgerAccount{
		Name: "现金", Type: LedgerAccountTypeAsset, Subtype: "cash",
	}, nil)
	assert.NoError(t, err)
	assert.Nil(t, savedTx)
}

func TestLedgerUsecase_CreateAccountRejectsInvalidTypeAndCreditFields(t *testing.T) {
	uc := NewLedgerUsecase(ledgerAccountRepoFixture(), ledgerCategoryRepoFixture(), &mockLedgerTransactionRepo{}, &mockLedgerBudgetRepo{}, &mockLedgerRecurringRepo{}, log.DefaultLogger)
	ctx := withUser(context.Background(), "user-123")

	// 用户不能创建 expense 类型账户
	_, err := uc.CreateAccount(ctx, &LedgerAccount{Name: "费用", Type: LedgerAccountTypeExpense}, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "仅支持 asset 或 liability")

	// subtype 与 type 不匹配
	_, err = uc.CreateAccount(ctx, &LedgerAccount{Name: "卡", Type: LedgerAccountTypeAsset, Subtype: "credit_card"}, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "子类型不合法")

	// 非信用卡设置账单日 → 拒绝
	billingDay := int32(5)
	_, err = uc.CreateAccount(ctx, &LedgerAccount{
		Name: "现金", Type: LedgerAccountTypeAsset, Subtype: "cash", BillingDay: &billingDay,
	}, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "仅信用卡账户可设置")
}

func TestLedgerUsecase_GetMonthlyStatsScope(t *testing.T) {
	txRepo := &mockLedgerTransactionRepo{
		legsFunc: func(ctx context.Context, userId string, txType string, systemAccountId uint, month string) ([]*LedgerPosting, error) {
			assert.Equal(t, "2026-08", month)
			if txType == LedgerTxTypeExpense {
				assert.Equal(t, uint(100), systemAccountId)
				// 同一笔多腿交易的两条系统腿 + 一条未挂分类的腿
				return []*LedgerPosting{
					{AccountId: 100, Amount: decimal.NewFromInt(60), CategoryId: 5},
					{AccountId: 100, Amount: decimal.NewFromInt(40), CategoryId: 5},
					{AccountId: 100, Amount: decimal.NewFromInt(20), CategoryId: 0},
				}, nil
			}
			assert.Equal(t, uint(101), systemAccountId)
			// 收入系统腿符号为负，统计取绝对值
			return []*LedgerPosting{
				{AccountId: 101, Amount: decimal.NewFromInt(-200), CategoryId: 7},
			}, nil
		},
	}
	categoryRepo := &mockLedgerCategoryRepo{
		listFunc: func(ctx context.Context, userId string) ([]*LedgerCategory, error) {
			return []*LedgerCategory{
				{Id: 5, Name: "餐饮", Direction: LedgerDirectionExpense},
				{Id: 9, Name: "其它", Direction: LedgerDirectionExpense, IsSystem: true},
				{Id: 7, Name: "工资", Direction: LedgerDirectionIncome},
				{Id: 12, Name: "其它", Direction: LedgerDirectionIncome, IsSystem: true},
			}, nil
		},
	}
	uc := NewLedgerUsecase(ledgerAccountRepoFixture(), categoryRepo, txRepo, &mockLedgerBudgetRepo{}, &mockLedgerRecurringRepo{}, log.DefaultLogger)

	stats, err := uc.GetMonthlyStats(withUser(context.Background(), "user-123"), "2026-08")

	assert.NoError(t, err)
	// 只查询 expense/income 两类腿：transfer/opening_balance/adjustment 不计入
	assert.ElementsMatch(t, []string{LedgerTxTypeExpense, LedgerTxTypeIncome}, txRepo.legsCalls)
	// 多腿不重复计：同一交易两条系统腿各计一次（60+40），总额 120+20=120... 即 60+40+20=120
	assert.True(t, stats.TotalExpense.Equal(decimal.NewFromInt(120)))
	assert.True(t, stats.TotalIncome.Equal(decimal.NewFromInt(200)))
	// 分类维度：餐饮 100，未挂分类归入"其它" 20
	assert.Len(t, stats.ExpenseByCategory, 2)
	assert.Equal(t, int64(5), stats.ExpenseByCategory[0].CategoryId)
	assert.Equal(t, "餐饮", stats.ExpenseByCategory[0].CategoryName)
	assert.True(t, stats.ExpenseByCategory[0].Amount.Equal(decimal.NewFromInt(100)))
	assert.Equal(t, int64(9), stats.ExpenseByCategory[1].CategoryId)
	assert.Equal(t, "其它", stats.ExpenseByCategory[1].CategoryName)
	assert.True(t, stats.ExpenseByCategory[1].Amount.Equal(decimal.NewFromInt(20)))
	assert.Len(t, stats.IncomeByCategory, 1)
	assert.Equal(t, "工资", stats.IncomeByCategory[0].CategoryName)

	// 月份格式错误
	_, err = uc.GetMonthlyStats(withUser(context.Background(), "user-123"), "2026/08")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "YYYY-MM")
}

func TestLedgerUsecase_ListAccountsExcludesSystemAndFillsBalance(t *testing.T) {
	accountRepo := ledgerAccountRepoFixture()
	ensured := false
	accountRepo.ensureFunc = func(ctx context.Context, userId string) error {
		ensured = true
		return nil
	}
	accountRepo.listFunc = func(ctx context.Context, userId string) ([]*LedgerAccount, error) {
		// data 层已过滤系统账户
		return []*LedgerAccount{{Id: 1, UserId: userId, Name: "现金"}}, nil
	}
	accountRepo.sumFunc = func(ctx context.Context, ids []uint) (map[uint]decimal.Decimal, error) {
		return map[uint]decimal.Decimal{1: decimal.NewFromInt(250)}, nil
	}
	uc := NewLedgerUsecase(accountRepo, ledgerCategoryRepoFixture(), &mockLedgerTransactionRepo{}, &mockLedgerBudgetRepo{}, &mockLedgerRecurringRepo{}, log.DefaultLogger)

	accounts, err := uc.ListAccounts(withUser(context.Background(), "user-123"))

	assert.NoError(t, err)
	assert.True(t, ensured)
	assert.Len(t, accounts, 1)
	assert.True(t, accounts[0].Balance.Equal(decimal.NewFromInt(250)))
}

func TestLedgerUsecase_SaveBudgetValidates(t *testing.T) {
	ctx := withUser(context.Background(), "user-123")
	categoryRepo := &mockLedgerCategoryRepo{
		findByUserFunc: func(ctx context.Context, userId string, id uint) (*LedgerCategory, error) {
			if id == 7 {
				return &LedgerCategory{Id: 7, UserId: userId, Name: "工资", Direction: LedgerDirectionIncome}, nil
			}
			return &LedgerCategory{Id: int64(id), UserId: userId, Name: "餐饮", Direction: LedgerDirectionExpense}, nil
		},
	}

	// income 分类 → 拒绝
	uc := NewLedgerUsecase(ledgerAccountRepoFixture(), categoryRepo, &mockLedgerTransactionRepo{}, &mockLedgerBudgetRepo{}, &mockLedgerRecurringRepo{}, log.DefaultLogger)
	_, err := uc.SaveBudget(ctx, &LedgerBudget{CategoryId: 7, Month: "2026-08", Amount: decimal.NewFromInt(100)})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "预算仅支持支出分类")

	// amount <= 0 → 拒绝
	_, err = uc.SaveBudget(ctx, &LedgerBudget{CategoryId: 5, Month: "2026-08", Amount: decimal.Zero})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "必须大于 0")

	// 月份格式错误
	_, err = uc.SaveBudget(ctx, &LedgerBudget{CategoryId: 5, Month: "2026/08", Amount: decimal.NewFromInt(100)})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "YYYY-MM")

	// 合法：走 upsert
	var upserted *LedgerBudget
	budgetRepo := &mockLedgerBudgetRepo{
		upsertFunc: func(ctx context.Context, b *LedgerBudget) (uint, error) {
			upserted = b
			return 1, nil
		},
	}
	uc = NewLedgerUsecase(ledgerAccountRepoFixture(), categoryRepo, &mockLedgerTransactionRepo{}, budgetRepo, &mockLedgerRecurringRepo{}, log.DefaultLogger)
	id, err := uc.SaveBudget(ctx, &LedgerBudget{CategoryId: 5, Month: "2026-08", Amount: decimal.NewFromInt(300)})
	assert.NoError(t, err)
	assert.Equal(t, uint(1), id)
	assert.NotNil(t, upserted)
	assert.Equal(t, "user-123", upserted.UserId)
}

func TestLedgerUsecase_ListBudgetsUsedIncludesSubtree(t *testing.T) {
	ctx := withUser(context.Background(), "user-123")
	// 分类树：餐饮(5) → 外卖(6)；收入：工资(7)
	categoryRepo := &mockLedgerCategoryRepo{
		listFunc: func(ctx context.Context, userId string) ([]*LedgerCategory, error) {
			return []*LedgerCategory{
				{Id: 5, Name: "餐饮", Direction: LedgerDirectionExpense},
				{Id: 6, Name: "外卖", Direction: LedgerDirectionExpense, ParentId: 5},
				{Id: 7, Name: "工资", Direction: LedgerDirectionIncome},
			}, nil
		},
	}
	txRepo := &mockLedgerTransactionRepo{
		legsFunc: func(ctx context.Context, userId string, txType string, systemAccountId uint, month string) ([]*LedgerPosting, error) {
			assert.Equal(t, LedgerTxTypeExpense, txType)
			assert.Equal(t, "2026-08", month)
			return []*LedgerPosting{
				{Amount: decimal.NewFromInt(100), CategoryId: 5}, // 挂在父分类
				{Amount: decimal.NewFromInt(50), CategoryId: 6},  // 挂在子分类
				{Amount: decimal.NewFromInt(30), CategoryId: 0},  // 未挂分类，不计入任何预算
			}, nil
		},
	}
	budgetRepo := &mockLedgerBudgetRepo{
		listFunc: func(ctx context.Context, userId string, month string) ([]*LedgerBudget, error) {
			return []*LedgerBudget{
				{Id: 1, UserId: userId, CategoryId: 5, Month: month, Amount: decimal.NewFromInt(300)},
				{Id: 2, UserId: userId, CategoryId: 6, Month: month, Amount: decimal.NewFromInt(100)},
			}, nil
		},
	}
	uc := NewLedgerUsecase(ledgerAccountRepoFixture(), categoryRepo, txRepo, budgetRepo, &mockLedgerRecurringRepo{}, log.DefaultLogger)

	budgets, err := uc.ListBudgets(ctx, "2026-08")

	assert.NoError(t, err)
	assert.Len(t, budgets, 2)
	// 父分类预算 used = 自身腿 + 子分类腿 = 100 + 50
	assert.Equal(t, "餐饮", budgets[0].CategoryName)
	assert.True(t, budgets[0].Used.Equal(decimal.NewFromInt(150)))
	// 子分类预算 used 只含自身 = 50
	assert.Equal(t, "外卖", budgets[1].CategoryName)
	assert.True(t, budgets[1].Used.Equal(decimal.NewFromInt(50)))
}

func TestLedgerUsecase_GetBalanceTrendFillsMissingDays(t *testing.T) {
	ctx := withUser(context.Background(), "user-123")
	txRepo := &mockLedgerTransactionRepo{
		legsUpToFunc: func(ctx context.Context, accountIds []uint, endExclusive time.Time) ([]*LedgerAccountLeg, error) {
			assert.Equal(t, []uint{1}, accountIds)
			return []*LedgerAccountLeg{
				// 区间前累计：+1000（开账）
				{AccountId: 1, BookedAt: "2026-07-20 09:00:00", Amount: decimal.NewFromInt(1000)},
				// 区间内：8-01 消费 -100，8-03 转账 -50；8-02 无交易
				{AccountId: 1, BookedAt: "2026-08-01 10:00:00", Amount: decimal.NewFromInt(-100)},
				{AccountId: 1, BookedAt: "2026-08-03 10:00:00", Amount: decimal.NewFromInt(-50)},
			}, nil
		},
	}
	uc := NewLedgerUsecase(ledgerAccountRepoFixture(), ledgerCategoryRepoFixture(), txRepo, &mockLedgerBudgetRepo{}, &mockLedgerRecurringRepo{}, log.DefaultLogger)

	points, err := uc.GetBalanceTrend(ctx, 1, "2026-08-01", "2026-08-04")

	assert.NoError(t, err)
	assert.Len(t, points, 4)
	// 首日余额 = 区间前累计 1000 + 当日 -100 = 900
	assert.Equal(t, "2026-08-01", points[0].Date)
	assert.True(t, points[0].Balance.Equal(decimal.NewFromInt(900)))
	// 8-02 无交易：补点，余额延续
	assert.Equal(t, "2026-08-02", points[1].Date)
	assert.True(t, points[1].Balance.Equal(decimal.NewFromInt(900)))
	assert.True(t, points[2].Balance.Equal(decimal.NewFromInt(850)))
	assert.True(t, points[3].Balance.Equal(decimal.NewFromInt(850)))
}

func TestLedgerUsecase_GetBalanceTrendNetAssetAcrossAccounts(t *testing.T) {
	ctx := withUser(context.Background(), "user-123")
	accountRepo := ledgerAccountRepoFixture()
	accountRepo.listFunc = func(ctx context.Context, userId string) ([]*LedgerAccount, error) {
		return []*LedgerAccount{
			{Id: 1, UserId: userId, Name: "现金", Type: LedgerAccountTypeAsset},
			{Id: 2, UserId: userId, Name: "信用卡", Type: LedgerAccountTypeLiability},
		}, nil
	}
	txRepo := &mockLedgerTransactionRepo{
		legsUpToFunc: func(ctx context.Context, accountIds []uint, endExclusive time.Time) ([]*LedgerAccountLeg, error) {
			assert.ElementsMatch(t, []uint{1, 2}, accountIds)
			return []*LedgerAccountLeg{
				// 区间前：现金 +500，信用卡 -200 → 净资产 300
				{AccountId: 1, BookedAt: "2026-07-01 09:00:00", Amount: decimal.NewFromInt(500)},
				{AccountId: 2, BookedAt: "2026-07-01 09:00:00", Amount: decimal.NewFromInt(-200)},
				// 区间内同日两账户变动合并：+100 与 -40 → 净 +60
				{AccountId: 1, BookedAt: "2026-08-01 10:00:00", Amount: decimal.NewFromInt(100)},
				{AccountId: 2, BookedAt: "2026-08-01 11:00:00", Amount: decimal.NewFromInt(-40)},
			}, nil
		},
	}
	uc := NewLedgerUsecase(accountRepo, ledgerCategoryRepoFixture(), txRepo, &mockLedgerBudgetRepo{}, &mockLedgerRecurringRepo{}, log.DefaultLogger)

	// accountId=0：净资产走势
	points, err := uc.GetBalanceTrend(ctx, 0, "2026-08-01", "2026-08-02")

	assert.NoError(t, err)
	assert.Len(t, points, 2)
	assert.True(t, points[0].Balance.Equal(decimal.NewFromInt(360)))
	assert.True(t, points[1].Balance.Equal(decimal.NewFromInt(360)))

	// 开始晚于结束 → 拒绝
	_, err = uc.GetBalanceTrend(ctx, 0, "2026-08-02", "2026-08-01")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "不能晚于结束时间")
}
