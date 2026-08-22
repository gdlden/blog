package biz

import (
	"blog/internal/utils"
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/shopspring/decimal"
)

// 账户类型枚举。
const (
	LedgerAccountTypeAsset     = "asset"
	LedgerAccountTypeLiability = "liability"
	LedgerAccountTypeExpense   = "expense"
	LedgerAccountTypeIncome    = "income"
	LedgerAccountTypeEquity    = "equity"
)

// 交易类型枚举。
const (
	LedgerTxTypeExpense        = "expense"
	LedgerTxTypeIncome         = "income"
	LedgerTxTypeTransfer       = "transfer"
	LedgerTxTypeOpeningBalance = "opening_balance"
	LedgerTxTypeAdjustment     = "adjustment"
)

// 分类方向枚举。
const (
	LedgerDirectionExpense = "expense"
	LedgerDirectionIncome  = "income"
)

// ledgerSystemAccountNames 每用户的系统账户：费用/收入/期初调整，isSystem=true，惰性创建。
var ledgerSystemAccountNames = map[string]string{
	LedgerAccountTypeExpense: "费用",
	LedgerAccountTypeIncome:  "收入",
	LedgerAccountTypeEquity:  "期初调整",
}

// ledgerAccountSubtypes 各账户类型允许的子类型。
var ledgerAccountSubtypes = map[string]map[string]bool{
	LedgerAccountTypeAsset: {
		"cash": true, "debit_card": true, "e_wallet": true,
		"prepaid_card": true, "investment": true, "other": true,
	},
	LedgerAccountTypeLiability: {
		"credit_card": true, "huabei_like": true, "loan_payable": true,
		"loan_receivable": true, "other": true,
	},
}

var ledgerTxTypes = map[string]bool{
	LedgerTxTypeExpense:        true,
	LedgerTxTypeIncome:         true,
	LedgerTxTypeTransfer:       true,
	LedgerTxTypeOpeningBalance: true,
	LedgerTxTypeAdjustment:     true,
}

// LedgerAccount 记账账户。Balance 为实时聚合值（SUM(posting.amount)），不落库。
type LedgerAccount struct {
	Id            int64
	UserId        string
	Name          string
	Type          string
	Subtype       string
	CreditLimit   *decimal.Decimal
	BillingDay    *int32
	PaymentDueDay *int32
	Remark        string
	Sort          int32
	Archived      bool
	IsSystem      bool
	Balance       decimal.Decimal
}

// LedgerCategory 记账分类，最多两级（ParentId=0 为顶级）。
type LedgerCategory struct {
	Id        int64
	UserId    string
	ParentId  int64
	Name      string
	Direction string
	Sort      int32
	IsSystem  bool
}

// LedgerPosting 交易分录（腿）。CategoryId=0 表示未挂分类。
type LedgerPosting struct {
	Id            int64
	TransactionId int64
	AccountId     int64
	Amount        decimal.Decimal
	CategoryId    int64
	Sort          int32
}

// LedgerTransaction 复式交易。BookedAt 格式 "2006-01-02 15:04:05"。
type LedgerTransaction struct {
	Id        int64
	UserId    string
	Type      string
	BookedAt  string
	Remark    string
	CreatedAt string
	Postings  []*LedgerPosting
}

type LedgerTransactionQuery struct {
	Page       int64
	PageSize   int64
	AccountId  int64
	CategoryId int64
	Type       string
	StartTime  string
	EndTime    string
}

type LedgerCategoryAmount struct {
	CategoryId   int64
	CategoryName string
	Amount       decimal.Decimal
}

// LedgerBudget 分类月度预算。Used 为查询时实时聚合，不落库。
type LedgerBudget struct {
	Id           int64
	UserId       string
	CategoryId   int64
	Month        string
	Amount       decimal.Decimal
	CategoryName string
	Used         decimal.Decimal
}

// LedgerAccountLeg 带记账日期的账户分录腿，用于余额走势聚合。
type LedgerAccountLeg struct {
	AccountId int64
	BookedAt  string // "2006-01-02 15:04:05"
	Amount    decimal.Decimal
}

type LedgerBalancePoint struct {
	Date    string
	Balance decimal.Decimal
}

type LedgerMonthlyStats struct {
	Month             string
	TotalExpense      decimal.Decimal
	TotalIncome       decimal.Decimal
	ExpenseByCategory []*LedgerCategoryAmount
	IncomeByCategory  []*LedgerCategoryAmount
}

// LedgerRecurring 周期账单规则：每月 DayOfMonth 日惰性生成一笔 expense/income 交易。
// AccountName/CategoryName/NextDate 为查询时聚合的展示字段，不落库。
type LedgerRecurring struct {
	Id                 int64
	UserId             string
	AccountId          int64
	CategoryId         int64
	Type               string
	Amount             decimal.Decimal
	Remark             string
	DayOfMonth         int32
	StartMonth         string
	LastGeneratedMonth string
	Enabled            bool
	AccountName        string
	CategoryName       string
	NextDate           string
}

type LedgerAccountRepo interface {
	Save(context.Context, *LedgerAccount) (uint, error)
	Update(context.Context, *LedgerAccount) error
	DeleteByUserIdAndId(context.Context, string, uint) error
	FindByUserIdAndId(context.Context, string, uint) (*LedgerAccount, error)
	FindSystemByUserIdAndType(context.Context, string, string) (*LedgerAccount, error)
	// ListByUserId 不含系统账户
	ListByUserId(context.Context, string) ([]*LedgerAccount, error)
	// EnsureSystemAccounts 惰性创建费用/收入/期初调整三个系统账户
	EnsureSystemAccounts(context.Context, string) error
	SumBalanceByAccountIds(context.Context, []uint) (map[uint]decimal.Decimal, error)
	CountPostingsByAccountId(context.Context, uint) (int64, error)
}

type LedgerCategoryRepo interface {
	Save(context.Context, *LedgerCategory) (uint, error)
	Update(context.Context, *LedgerCategory) error
	FindByUserIdAndId(context.Context, string, uint) (*LedgerCategory, error)
	ListByUserId(context.Context, string) ([]*LedgerCategory, error)
	// EnsurePresetCategories 首次使用时惰性插入预置分类集
	EnsurePresetCategories(context.Context, string) error
	HasChildren(context.Context, uint) (bool, error)
	// FindFallbackByDirection 返回该方向的 isSystem 兜底分类（"其它"）
	FindFallbackByDirection(context.Context, string, string) (*LedgerCategory, error)
	// DeleteWithMigration 删除分类并把引用它的 posting 迁移到兜底分类（同一 DB 事务）
	DeleteWithMigration(context.Context, string, uint, uint) error
}

type LedgerTransactionRepo interface {
	Save(context.Context, *LedgerTransaction) (uint, error)
	Update(context.Context, *LedgerTransaction) error
	DeleteByUserIdAndId(context.Context, string, uint) error
	FindByUserIdAndId(context.Context, string, uint) (*LedgerTransaction, error)
	PageByUserId(context.Context, string, *LedgerTransactionQuery) ([]*LedgerTransaction, int64, error)
	// ListSystemLegsByMonth 返回某月内指定类型交易中落在系统账户上的分录腿
	ListSystemLegsByMonth(context.Context, string, string, uint, string) ([]*LedgerPosting, error)
	// ListAccountLegsUpTo 返回指定账户在 endExclusive 之前的全部分录腿（含记账日期，含所有交易类型）
	ListAccountLegsUpTo(context.Context, []uint, time.Time) ([]*LedgerAccountLeg, error)
}

type LedgerBudgetRepo interface {
	Upsert(context.Context, *LedgerBudget) (uint, error)
	DeleteByUserIdAndId(context.Context, string, uint) error
	ListByUserIdAndMonth(context.Context, string, string) ([]*LedgerBudget, error)
}

type LedgerRecurringRepo interface {
	Save(context.Context, *LedgerRecurring) (uint, error)
	Update(context.Context, *LedgerRecurring) error
	DeleteByUserIdAndId(context.Context, string, uint) error
	FindByUserIdAndId(context.Context, string, uint) (*LedgerRecurring, error)
	ListByUserId(context.Context, string) ([]*LedgerRecurring, error)
	ListEnabledByUserId(context.Context, string) ([]*LedgerRecurring, error)
	UpdateLastGeneratedMonth(context.Context, string, uint, string) error
}

type LedgerUsecase struct {
	accountRepo   LedgerAccountRepo
	categoryRepo  LedgerCategoryRepo
	txRepo        LedgerTransactionRepo
	budgetRepo    LedgerBudgetRepo
	recurringRepo LedgerRecurringRepo
	log           *log.Helper
}

func NewLedgerUsecase(accountRepo LedgerAccountRepo, categoryRepo LedgerCategoryRepo, txRepo LedgerTransactionRepo, budgetRepo LedgerBudgetRepo, recurringRepo LedgerRecurringRepo, logger log.Logger) *LedgerUsecase {
	return &LedgerUsecase{
		accountRepo:   accountRepo,
		categoryRepo:  categoryRepo,
		txRepo:        txRepo,
		budgetRepo:    budgetRepo,
		recurringRepo: recurringRepo,
		log:           log.NewHelper(logger),
	}
}

// ---------- 账户 ----------

func (uc *LedgerUsecase) CreateAccount(ctx context.Context, account *LedgerAccount, openingBalance *decimal.Decimal) (uint, error) {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return 0, err
	}
	if err := validateLedgerAccount(account); err != nil {
		return 0, err
	}
	account.UserId = userId
	account.IsSystem = false
	if openingBalance != nil && !openingBalance.IsZero() {
		if err := uc.accountRepo.EnsureSystemAccounts(ctx, userId); err != nil {
			return 0, err
		}
	}
	id, err := uc.accountRepo.Save(ctx, account)
	if err != nil {
		return 0, err
	}
	if openingBalance != nil && !openingBalance.IsZero() {
		// 开账：两腿 [新账户 +openingBalance, 期初调整系统账户 -openingBalance]，bookedAt=当天
		equity, err := uc.accountRepo.FindSystemByUserIdAndType(ctx, userId, LedgerAccountTypeEquity)
		if err != nil {
			return 0, err
		}
		openingTx := &LedgerTransaction{
			UserId:   userId,
			Type:     LedgerTxTypeOpeningBalance,
			BookedAt: time.Now().Format("2006-01-02 15:04:05"),
			Remark:   "期初余额",
			Postings: []*LedgerPosting{
				{AccountId: int64(id), Amount: *openingBalance, Sort: 0},
				{AccountId: equity.Id, Amount: openingBalance.Neg(), Sort: 1},
			},
		}
		if _, err := uc.txRepo.Save(ctx, openingTx); err != nil {
			return 0, err
		}
	}
	return id, nil
}

func (uc *LedgerUsecase) UpdateAccount(ctx context.Context, account *LedgerAccount) (uint, error) {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return 0, err
	}
	if err := validateLedgerAccount(account); err != nil {
		return 0, err
	}
	dbAccount, err := uc.accountRepo.FindByUserIdAndId(ctx, userId, uint(account.Id))
	if err != nil {
		return 0, err
	}
	if dbAccount.IsSystem {
		return 0, errors.New("系统账户不允许修改")
	}
	account.Id = dbAccount.Id
	account.UserId = userId
	account.IsSystem = false
	if err := uc.accountRepo.Update(ctx, account); err != nil {
		return 0, err
	}
	return uint(dbAccount.Id), nil
}

func (uc *LedgerUsecase) DeleteAccount(ctx context.Context, accountId uint) error {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return err
	}
	account, err := uc.accountRepo.FindByUserIdAndId(ctx, userId, accountId)
	if err != nil {
		return err
	}
	if account.IsSystem {
		return errors.New("系统账户不允许删除")
	}
	count, err := uc.accountRepo.CountPostingsByAccountId(ctx, accountId)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("该账户存在交易记录，禁止删除，可改为归档")
	}
	return uc.accountRepo.DeleteByUserIdAndId(ctx, userId, accountId)
}

func (uc *LedgerUsecase) GetAccount(ctx context.Context, accountId uint) (*LedgerAccount, error) {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return nil, err
	}
	account, err := uc.accountRepo.FindByUserIdAndId(ctx, userId, accountId)
	if err != nil {
		return nil, err
	}
	if err := uc.fillAccountBalances(ctx, []*LedgerAccount{account}); err != nil {
		return nil, err
	}
	return account, nil
}

func (uc *LedgerUsecase) ListAccounts(ctx context.Context) ([]*LedgerAccount, error) {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return nil, err
	}
	// 首次使用时惰性创建系统账户
	if err := uc.accountRepo.EnsureSystemAccounts(ctx, userId); err != nil {
		return nil, err
	}
	accounts, err := uc.accountRepo.ListByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}
	if err := uc.fillAccountBalances(ctx, accounts); err != nil {
		return nil, err
	}
	return accounts, nil
}

func (uc *LedgerUsecase) fillAccountBalances(ctx context.Context, accounts []*LedgerAccount) error {
	ids := make([]uint, 0, len(accounts))
	for _, account := range accounts {
		ids = append(ids, uint(account.Id))
	}
	if len(ids) == 0 {
		return nil
	}
	sums, err := uc.accountRepo.SumBalanceByAccountIds(ctx, ids)
	if err != nil {
		return err
	}
	for _, account := range accounts {
		account.Balance = sums[uint(account.Id)]
	}
	return nil
}

// validateLedgerAccount 校验账户：用户只能创建/更新 asset/liability，subtype 与 type 匹配，信用卡三字段仅 credit_card 使用。
func validateLedgerAccount(account *LedgerAccount) error {
	if account == nil {
		return errors.New("请求不能为空")
	}
	if account.Name == "" {
		return errors.New("账户名称不能为空")
	}
	if account.Type != LedgerAccountTypeAsset && account.Type != LedgerAccountTypeLiability {
		return errors.New("账户类型仅支持 asset 或 liability")
	}
	subtypes := ledgerAccountSubtypes[account.Type]
	if account.Subtype == "" || !subtypes[account.Subtype] {
		return fmt.Errorf("账户子类型不合法: %s", account.Subtype)
	}
	if account.Subtype != "credit_card" {
		if account.CreditLimit != nil || account.BillingDay != nil || account.PaymentDueDay != nil {
			return errors.New("信用额度、账单日、还款日仅信用卡账户可设置")
		}
	}
	if account.CreditLimit != nil && account.CreditLimit.IsNegative() {
		return errors.New("信用额度不能为负数")
	}
	if account.BillingDay != nil && (*account.BillingDay < 1 || *account.BillingDay > 31) {
		return errors.New("账单日必须在 1-31 之间")
	}
	if account.PaymentDueDay != nil && (*account.PaymentDueDay < 1 || *account.PaymentDueDay > 31) {
		return errors.New("还款日必须在 1-31 之间")
	}
	return nil
}

// ---------- 分类 ----------

func (uc *LedgerUsecase) ListCategories(ctx context.Context) ([]*LedgerCategory, error) {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return nil, err
	}
	// 首次调用自动初始化预置分类
	if err := uc.categoryRepo.EnsurePresetCategories(ctx, userId); err != nil {
		return nil, err
	}
	return uc.categoryRepo.ListByUserId(ctx, userId)
}

func (uc *LedgerUsecase) SaveCategory(ctx context.Context, category *LedgerCategory) (uint, error) {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return 0, err
	}
	if err := uc.validateCategory(ctx, userId, category); err != nil {
		return 0, err
	}
	category.UserId = userId
	category.IsSystem = false
	return uc.categoryRepo.Save(ctx, category)
}

func (uc *LedgerUsecase) UpdateCategory(ctx context.Context, category *LedgerCategory) (uint, error) {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return 0, err
	}
	dbCategory, err := uc.categoryRepo.FindByUserIdAndId(ctx, userId, uint(category.Id))
	if err != nil {
		return 0, err
	}
	if dbCategory.IsSystem {
		return 0, errors.New("系统分类不允许修改")
	}
	if err := uc.validateCategory(ctx, userId, category); err != nil {
		return 0, err
	}
	category.Id = dbCategory.Id
	category.UserId = userId
	category.IsSystem = false
	if err := uc.categoryRepo.Update(ctx, category); err != nil {
		return 0, err
	}
	return uint(dbCategory.Id), nil
}

// validateCategory 校验分类：名称/方向必填，parentId 非 0 时父分类必须存在、归属当前用户、为顶级分类且方向一致（最多两级）。
func (uc *LedgerUsecase) validateCategory(ctx context.Context, userId string, category *LedgerCategory) error {
	if category == nil {
		return errors.New("请求不能为空")
	}
	if category.Name == "" {
		return errors.New("分类名称不能为空")
	}
	if category.Direction != LedgerDirectionExpense && category.Direction != LedgerDirectionIncome {
		return errors.New("分类方向仅支持 expense 或 income")
	}
	if category.ParentId != 0 {
		parent, err := uc.categoryRepo.FindByUserIdAndId(ctx, userId, uint(category.ParentId))
		if err != nil {
			return errors.New("父分类不存在或不属于当前用户")
		}
		if parent.ParentId != 0 {
			return errors.New("分类最多支持两级")
		}
		if parent.Direction != category.Direction {
			return errors.New("子分类方向必须与父分类一致")
		}
	}
	return nil
}

func (uc *LedgerUsecase) DeleteCategory(ctx context.Context, categoryId uint) error {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return err
	}
	category, err := uc.categoryRepo.FindByUserIdAndId(ctx, userId, categoryId)
	if err != nil {
		return err
	}
	if category.IsSystem {
		return errors.New("系统分类不允许删除")
	}
	// 仅允许删除叶子分类
	hasChildren, err := uc.categoryRepo.HasChildren(ctx, categoryId)
	if err != nil {
		return err
	}
	if hasChildren {
		return errors.New("该分类存在子分类，禁止删除")
	}
	fallback, err := uc.categoryRepo.FindFallbackByDirection(ctx, userId, category.Direction)
	if err != nil {
		return err
	}
	if fallback.Id == category.Id {
		return errors.New("系统分类不允许删除")
	}
	return uc.categoryRepo.DeleteWithMigration(ctx, userId, categoryId, uint(fallback.Id))
}

// ---------- 交易 ----------

func (uc *LedgerUsecase) CreateTransaction(ctx context.Context, tx *LedgerTransaction) (uint, error) {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return 0, err
	}
	tx.UserId = userId
	if err := uc.validateTransaction(ctx, userId, tx); err != nil {
		return 0, err
	}
	return uc.txRepo.Save(ctx, tx)
}

func (uc *LedgerUsecase) UpdateTransaction(ctx context.Context, tx *LedgerTransaction) (uint, error) {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return 0, err
	}
	dbTx, err := uc.txRepo.FindByUserIdAndId(ctx, userId, uint(tx.Id))
	if err != nil {
		return 0, err
	}
	tx.Id = dbTx.Id
	tx.UserId = userId
	if err := uc.validateTransaction(ctx, userId, tx); err != nil {
		return 0, err
	}
	if err := uc.txRepo.Update(ctx, tx); err != nil {
		return 0, err
	}
	return uint(dbTx.Id), nil
}

func (uc *LedgerUsecase) DeleteTransaction(ctx context.Context, txId uint) error {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return err
	}
	if _, err := uc.txRepo.FindByUserIdAndId(ctx, userId, txId); err != nil {
		return err
	}
	return uc.txRepo.DeleteByUserIdAndId(ctx, userId, txId)
}

func (uc *LedgerUsecase) GetTransaction(ctx context.Context, txId uint) (*LedgerTransaction, error) {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return nil, err
	}
	return uc.txRepo.FindByUserIdAndId(ctx, userId, txId)
}

func (uc *LedgerUsecase) PageTransactions(ctx context.Context, query *LedgerTransactionQuery) ([]*LedgerTransaction, int64, error) {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return nil, 0, err
	}
	if query.Type != "" && !ledgerTxTypes[query.Type] {
		return nil, 0, fmt.Errorf("交易类型不合法: %s", query.Type)
	}
	return uc.txRepo.PageByUserId(ctx, userId, query)
}

// validateTransaction 校验复式交易不变量：accountId=0 系统腿解析、分录 ≥2、Σamount=0、
// 每条 amount≠0、账户归属当前用户、分类归属与方向匹配。解析替换后再做归属与平衡校验。
func (uc *LedgerUsecase) validateTransaction(ctx context.Context, userId string, tx *LedgerTransaction) error {
	if !ledgerTxTypes[tx.Type] {
		return fmt.Errorf("交易类型不合法: %s", tx.Type)
	}
	if tx.BookedAt == "" {
		return errors.New("记账时间不能为空")
	}
	if len(tx.Postings) < 2 {
		return errors.New("交易分录至少需要两条")
	}
	if err := uc.resolveSystemPostings(ctx, userId, tx); err != nil {
		return err
	}
	sum := decimal.Zero
	for _, posting := range tx.Postings {
		if posting.Amount.IsZero() {
			return errors.New("分录金额不能为 0")
		}
		sum = sum.Add(posting.Amount)
		if _, err := uc.accountRepo.FindByUserIdAndId(ctx, userId, uint(posting.AccountId)); err != nil {
			return fmt.Errorf("分录账户不存在或不属于当前用户: %d", posting.AccountId)
		}
		if err := uc.validatePostingCategory(ctx, userId, tx.Type, posting); err != nil {
			return err
		}
	}
	if !sum.IsZero() {
		return errors.New("交易分录金额合计必须为 0")
	}
	return nil
}

// resolveSystemPostings 把 accountId=0 的分录解析为系统对手账户：expense→系统费用账户，income→系统收入账户。
func (uc *LedgerUsecase) resolveSystemPostings(ctx context.Context, userId string, tx *LedgerTransaction) error {
	hasSystemLeg := false
	for _, posting := range tx.Postings {
		if posting.AccountId == 0 {
			hasSystemLeg = true
			break
		}
	}
	if !hasSystemLeg {
		return nil
	}
	var systemType string
	switch tx.Type {
	case LedgerTxTypeExpense:
		systemType = LedgerAccountTypeExpense
	case LedgerTxTypeIncome:
		systemType = LedgerAccountTypeIncome
	default:
		return errors.New("转账、开账与调整交易不支持系统对手账户（accountId 不能为 0）")
	}
	if err := uc.accountRepo.EnsureSystemAccounts(ctx, userId); err != nil {
		return err
	}
	systemAccount, err := uc.accountRepo.FindSystemByUserIdAndType(ctx, userId, systemType)
	if err != nil {
		return err
	}
	for _, posting := range tx.Postings {
		if posting.AccountId == 0 {
			posting.AccountId = systemAccount.Id
		}
	}
	return nil
}

// validatePostingCategory 校验分录分类：仅 expense/income 可挂分类，且分类归属当前用户、方向与交易类型匹配。
func (uc *LedgerUsecase) validatePostingCategory(ctx context.Context, userId, txType string, posting *LedgerPosting) error {
	if posting.CategoryId == 0 {
		return nil
	}
	if txType != LedgerTxTypeExpense && txType != LedgerTxTypeIncome {
		return errors.New("转账、开账与调整交易不能挂载分类")
	}
	category, err := uc.categoryRepo.FindByUserIdAndId(ctx, userId, uint(posting.CategoryId))
	if err != nil {
		return fmt.Errorf("分类不存在或不属于当前用户: %d", posting.CategoryId)
	}
	if category.Direction != txType {
		return errors.New("分类方向与交易类型不匹配")
	}
	return nil
}

// ---------- 月度统计 ----------

// GetMonthlyStats 月度统计：只统计 type=expense/income 的交易，
// 排除 transfer/opening_balance/adjustment —— 转账（如信用卡还款）不产生真实收支，
// 若计入会与还款前的消费支出重复计算。
// 口径：expense 总额 = 当月 expense 交易中落在系统费用账户上的腿的合计（多腿交易只计系统腿，不重复计）；
// 分类维度按这些腿的 categoryId 分组求和，未挂分类归入"其它"；income 对称（系统收入账户腿）。
func (uc *LedgerUsecase) GetMonthlyStats(ctx context.Context, month string) (*LedgerMonthlyStats, error) {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := time.Parse("2006-01", month); err != nil {
		return nil, errors.New("月份格式必须为 YYYY-MM")
	}
	if err := uc.accountRepo.EnsureSystemAccounts(ctx, userId); err != nil {
		return nil, err
	}
	expenseAccount, err := uc.accountRepo.FindSystemByUserIdAndType(ctx, userId, LedgerAccountTypeExpense)
	if err != nil {
		return nil, err
	}
	incomeAccount, err := uc.accountRepo.FindSystemByUserIdAndType(ctx, userId, LedgerAccountTypeIncome)
	if err != nil {
		return nil, err
	}
	categories, err := uc.categoryRepo.ListByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}
	expenseLegs, err := uc.txRepo.ListSystemLegsByMonth(ctx, userId, LedgerTxTypeExpense, uint(expenseAccount.Id), month)
	if err != nil {
		return nil, err
	}
	incomeLegs, err := uc.txRepo.ListSystemLegsByMonth(ctx, userId, LedgerTxTypeIncome, uint(incomeAccount.Id), month)
	if err != nil {
		return nil, err
	}
	totalExpense, expenseByCategory := aggregateLedgerLegs(expenseLegs, categories, LedgerDirectionExpense)
	totalIncome, incomeByCategory := aggregateLedgerLegs(incomeLegs, categories, LedgerDirectionIncome)
	return &LedgerMonthlyStats{
		Month:             month,
		TotalExpense:      totalExpense,
		TotalIncome:       totalIncome,
		ExpenseByCategory: expenseByCategory,
		IncomeByCategory:  incomeByCategory,
	}, nil
}

// aggregateLedgerLegs 按 categoryId 分组求和（取绝对值，收入腿符号为负）。
// categoryId=0 或分类已删除的腿归入该方向的"其它"兜底分类。
func aggregateLedgerLegs(legs []*LedgerPosting, categories []*LedgerCategory, direction string) (decimal.Decimal, []*LedgerCategoryAmount) {
	names := make(map[int64]string, len(categories))
	fallbackId := int64(0)
	for _, category := range categories {
		names[category.Id] = category.Name
		if category.Direction == direction && category.IsSystem {
			fallbackId = category.Id
		}
	}
	sums := make(map[int64]decimal.Decimal)
	order := make([]int64, 0)
	total := decimal.Zero
	for _, leg := range legs {
		categoryId := leg.CategoryId
		if categoryId == 0 {
			categoryId = fallbackId
		} else if _, ok := names[categoryId]; !ok {
			categoryId = fallbackId
		}
		amount := leg.Amount.Abs()
		if _, ok := sums[categoryId]; !ok {
			order = append(order, categoryId)
		}
		sums[categoryId] = sums[categoryId].Add(amount)
		total = total.Add(amount)
	}
	result := make([]*LedgerCategoryAmount, 0, len(order))
	for _, categoryId := range order {
		name, ok := names[categoryId]
		if !ok {
			name = "其它"
		}
		result = append(result, &LedgerCategoryAmount{
			CategoryId:   categoryId,
			CategoryName: name,
			Amount:       sums[categoryId],
		})
	}
	// 稳定输出：金额降序，同额按分类 id 升序
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].Amount.Equal(result[j].Amount) {
			return result[i].CategoryId < result[j].CategoryId
		}
		return result[i].Amount.GreaterThan(result[j].Amount)
	})
	return total, result
}

// LedgerSystemAccountSpec 供 data 层创建系统账户使用。
func LedgerSystemAccountSpec() map[string]string {
	return ledgerSystemAccountNames
}

// ---------- 预算 ----------

func (uc *LedgerUsecase) SaveBudget(ctx context.Context, budget *LedgerBudget) (uint, error) {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return 0, err
	}
	if budget.CategoryId <= 0 {
		return 0, errors.New("预算分类不能为空")
	}
	if budget.Amount.LessThanOrEqual(decimal.Zero) {
		return 0, errors.New("预算金额必须大于 0")
	}
	if budget.Month == "" {
		budget.Month = time.Now().Format("2006-01")
	}
	if _, err := time.Parse("2006-01", budget.Month); err != nil {
		return 0, errors.New("月份格式必须为 YYYY-MM")
	}
	category, err := uc.categoryRepo.FindByUserIdAndId(ctx, userId, uint(budget.CategoryId))
	if err != nil {
		return 0, errors.New("分类不存在或不属于当前用户")
	}
	if category.Direction != LedgerDirectionExpense {
		return 0, errors.New("预算仅支持支出分类")
	}
	budget.UserId = userId
	return uc.budgetRepo.Upsert(ctx, budget)
}

func (uc *LedgerUsecase) DeleteBudget(ctx context.Context, budgetId uint) error {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return err
	}
	return uc.budgetRepo.DeleteByUserIdAndId(ctx, userId, budgetId)
}

// ListBudgets 返回某月有预算的分类行；used = 当月 expense 交易中挂在该分类及其所有子孙分类的
// 系统费用腿合计（取绝对值）。分类最多两级，子孙即 parent_id = 该分类的直接子分类。
func (uc *LedgerUsecase) ListBudgets(ctx context.Context, month string) ([]*LedgerBudget, error) {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return nil, err
	}
	if month == "" {
		month = time.Now().Format("2006-01")
	}
	if _, err := time.Parse("2006-01", month); err != nil {
		return nil, errors.New("月份格式必须为 YYYY-MM")
	}
	budgets, err := uc.budgetRepo.ListByUserIdAndMonth(ctx, userId, month)
	if err != nil {
		return nil, err
	}
	if len(budgets) == 0 {
		return budgets, nil
	}
	if err := uc.accountRepo.EnsureSystemAccounts(ctx, userId); err != nil {
		return nil, err
	}
	expenseAccount, err := uc.accountRepo.FindSystemByUserIdAndType(ctx, userId, LedgerAccountTypeExpense)
	if err != nil {
		return nil, err
	}
	legs, err := uc.txRepo.ListSystemLegsByMonth(ctx, userId, LedgerTxTypeExpense, uint(expenseAccount.Id), month)
	if err != nil {
		return nil, err
	}
	categories, err := uc.categoryRepo.ListByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}
	categoryById := make(map[int64]*LedgerCategory, len(categories))
	for _, category := range categories {
		categoryById[category.Id] = category
	}
	// 每条腿归属的预算分类集合：腿分类自身 + 其父分类
	for _, budget := range budgets {
		used := decimal.Zero
		for _, leg := range legs {
			if leg.CategoryId == budget.CategoryId {
				used = used.Add(leg.Amount.Abs())
				continue
			}
			if category, ok := categoryById[leg.CategoryId]; ok && category.ParentId == budget.CategoryId {
				used = used.Add(leg.Amount.Abs())
			}
		}
		budget.Used = used
		if category, ok := categoryById[budget.CategoryId]; ok {
			budget.CategoryName = category.Name
		}
	}
	return budgets, nil
}

// ---------- 余额走势 ----------

// GetBalanceTrend 按日粒度返回余额走势，区间内无变动的日期也补齐点。
// 首日余额 = 区间开始前该账户（组）所有 posting 合计 + 当日净额，之后逐日累计。
// accountId=0 时为净资产走势：全部非系统账户（asset+liability）每日合计。
func (uc *LedgerUsecase) GetBalanceTrend(ctx context.Context, accountId int64, startTime, endTime string) ([]*LedgerBalancePoint, error) {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return nil, err
	}
	end := time.Now()
	if endTime != "" {
		end, err = time.ParseInLocation("2006-01-02", endTime, time.Local)
		if err != nil {
			return nil, errors.New("结束时间格式必须为 YYYY-MM-DD")
		}
	}
	end = time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, time.Local)
	start := end.AddDate(0, -6, 0)
	if startTime != "" {
		start, err = time.ParseInLocation("2006-01-02", startTime, time.Local)
		if err != nil {
			return nil, errors.New("开始时间格式必须为 YYYY-MM-DD")
		}
	}
	if start.After(end) {
		return nil, errors.New("开始时间不能晚于结束时间")
	}

	var accountIds []uint
	if accountId != 0 {
		account, err := uc.accountRepo.FindByUserIdAndId(ctx, userId, uint(accountId))
		if err != nil {
			return nil, err
		}
		accountIds = []uint{uint(account.Id)}
	} else {
		accounts, err := uc.accountRepo.ListByUserId(ctx, userId)
		if err != nil {
			return nil, err
		}
		for _, account := range accounts {
			accountIds = append(accountIds, uint(account.Id))
		}
	}

	startDate := start.Format("2006-01-02")
	endDate := end.Format("2006-01-02")
	opening := decimal.Zero
	dailyNet := make(map[string]decimal.Decimal)
	if len(accountIds) > 0 {
		legs, err := uc.txRepo.ListAccountLegsUpTo(ctx, accountIds, end.AddDate(0, 0, 1))
		if err != nil {
			return nil, err
		}
		for _, leg := range legs {
			date := leg.BookedAt
			if len(date) > 10 {
				date = date[:10]
			}
			if date < startDate {
				opening = opening.Add(leg.Amount)
			} else if date <= endDate {
				dailyNet[date] = dailyNet[date].Add(leg.Amount)
			}
		}
	}

	points := make([]*LedgerBalancePoint, 0)
	balance := opening
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		date := d.Format("2006-01-02")
		balance = balance.Add(dailyNet[date])
		points = append(points, &LedgerBalancePoint{Date: date, Balance: balance})
	}
	return points, nil
}

// ---------- 周期账单 ----------

func (uc *LedgerUsecase) SaveRecurring(ctx context.Context, rule *LedgerRecurring) (uint, error) {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return 0, err
	}
	if err := uc.validateRecurring(ctx, userId, rule); err != nil {
		return 0, err
	}
	rule.UserId = userId
	if rule.Id > 0 {
		dbRule, err := uc.recurringRepo.FindByUserIdAndId(ctx, userId, uint(rule.Id))
		if err != nil {
			return 0, err
		}
		rule.Id = dbRule.Id
		rule.LastGeneratedMonth = dbRule.LastGeneratedMonth
		if err := uc.recurringRepo.Update(ctx, rule); err != nil {
			return 0, err
		}
		return uint(rule.Id), nil
	}
	return uc.recurringRepo.Save(ctx, rule)
}

func (uc *LedgerUsecase) DeleteRecurring(ctx context.Context, recurringId uint) error {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return err
	}
	return uc.recurringRepo.DeleteByUserIdAndId(ctx, userId, recurringId)
}

// ListRecurring 返回全部规则，并聚合账户名/分类名/下一次应生成日期（disabled 时 NextDate 为空）。
func (uc *LedgerUsecase) ListRecurring(ctx context.Context) ([]*LedgerRecurring, error) {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return nil, err
	}
	rules, err := uc.recurringRepo.ListByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}
	if len(rules) == 0 {
		return rules, nil
	}
	accounts, err := uc.accountRepo.ListByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}
	categories, err := uc.categoryRepo.ListByUserId(ctx, userId)
	if err != nil {
		return nil, err
	}
	accountNames := make(map[int64]string, len(accounts))
	for _, account := range accounts {
		accountNames[account.Id] = account.Name
	}
	categoryNames := make(map[int64]string, len(categories))
	for _, category := range categories {
		categoryNames[category.Id] = category.Name
	}
	for _, rule := range rules {
		rule.AccountName = accountNames[rule.AccountId]
		rule.CategoryName = categoryNames[rule.CategoryId]
		if !rule.Enabled {
			rule.NextDate = ""
			continue
		}
		month, err := ledgerRecurringPendingMonth(rule)
		if err != nil {
			return nil, err
		}
		rule.NextDate = ledgerRecurringDueDate(month, rule.DayOfMonth).Format("2006-01-02")
	}
	return rules, nil
}

// ApplyRecurring 惰性生成：遍历启用规则，从 max(startMonth, lastGeneratedMonth 的下一月) 起逐月检查，
// 应生成日期 <= 今天则生成交易并推进 lastGeneratedMonth，> 今天则停止该规则。重复调用幂等。
func (uc *LedgerUsecase) ApplyRecurring(ctx context.Context) (int32, error) {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return 0, err
	}
	rules, err := uc.recurringRepo.ListEnabledByUserId(ctx, userId)
	if err != nil {
		return 0, err
	}
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	created := int32(0)
	for _, rule := range rules {
		month, err := ledgerRecurringPendingMonth(rule)
		if err != nil {
			return created, err
		}
		for {
			due := ledgerRecurringDueDate(month, rule.DayOfMonth)
			if due.After(today) {
				break
			}
			if _, err := uc.CreateTransaction(ctx, buildLedgerRecurringTx(userId, rule, due)); err != nil {
				return created, err
			}
			if err := uc.recurringRepo.UpdateLastGeneratedMonth(ctx, userId, uint(rule.Id), month.Format("2006-01")); err != nil {
				return created, err
			}
			created++
			month = month.AddDate(0, 1, 0)
		}
	}
	return created, nil
}

// validateRecurring 校验周期账单：仅 expense/income，dayOfMonth 1-31，amount > 0，
// startMonth 为 YYYY-MM，账户与分类存在且归属当前用户、分类方向与类型匹配。
func (uc *LedgerUsecase) validateRecurring(ctx context.Context, userId string, rule *LedgerRecurring) error {
	if rule == nil {
		return errors.New("请求不能为空")
	}
	if rule.Type != LedgerTxTypeExpense && rule.Type != LedgerTxTypeIncome {
		return errors.New("周期账单类型仅支持 expense 或 income")
	}
	if rule.DayOfMonth < 1 || rule.DayOfMonth > 31 {
		return errors.New("每月生成日必须在 1-31 之间")
	}
	if rule.Amount.LessThanOrEqual(decimal.Zero) {
		return errors.New("周期账单金额必须大于 0")
	}
	if _, err := time.ParseInLocation("2006-01", rule.StartMonth, time.Local); err != nil {
		return errors.New("起始月份格式必须为 YYYY-MM")
	}
	if rule.AccountId <= 0 {
		return errors.New("周期账单账户不能为空")
	}
	if _, err := uc.accountRepo.FindByUserIdAndId(ctx, userId, uint(rule.AccountId)); err != nil {
		return errors.New("账户不存在或不属于当前用户")
	}
	if rule.CategoryId <= 0 {
		return errors.New("周期账单分类不能为空")
	}
	category, err := uc.categoryRepo.FindByUserIdAndId(ctx, userId, uint(rule.CategoryId))
	if err != nil {
		return errors.New("分类不存在或不属于当前用户")
	}
	if category.Direction != rule.Type {
		return errors.New("分类方向与周期账单类型不匹配")
	}
	return nil
}

// ledgerRecurringPendingMonth 下一次待生成月份：max(startMonth, lastGeneratedMonth 的下一月)。
func ledgerRecurringPendingMonth(rule *LedgerRecurring) (time.Time, error) {
	start, err := time.ParseInLocation("2006-01", rule.StartMonth, time.Local)
	if err != nil {
		return time.Time{}, errors.New("起始月份格式必须为 YYYY-MM")
	}
	if rule.LastGeneratedMonth == "" {
		return start, nil
	}
	last, err := time.ParseInLocation("2006-01", rule.LastGeneratedMonth, time.Local)
	if err != nil {
		return time.Time{}, errors.New("最近生成月份格式必须为 YYYY-MM")
	}
	next := last.AddDate(0, 1, 0)
	if next.Before(start) {
		return start, nil
	}
	return next, nil
}

// ledgerRecurringDueDate 短月钳制：应生成日 = min(dayOfMonth, 当月最后一天)，如 2 月 30 → 2 月 28/29。
func ledgerRecurringDueDate(month time.Time, dayOfMonth int32) time.Time {
	lastDay := time.Date(month.Year(), month.Month()+1, 0, 0, 0, 0, 0, time.Local).Day()
	day := int(dayOfMonth)
	if day > lastDay {
		day = lastDay
	}
	return time.Date(month.Year(), month.Month(), day, 0, 0, 0, 0, time.Local)
}

// buildLedgerRecurringTx 构造周期账单交易，走现有复式写入路径：系统腿 accountId=0 由
// validateTransaction 解析为费用/收入系统账户；分类挂在系统腿上以进入月度统计。
func buildLedgerRecurringTx(userId string, rule *LedgerRecurring, due time.Time) *LedgerTransaction {
	tx := &LedgerTransaction{
		UserId:   userId,
		Type:     rule.Type,
		BookedAt: due.Format("2006-01-02 15:04:05"),
		Remark:   rule.Remark,
	}
	if rule.Type == LedgerTxTypeIncome {
		tx.Postings = []*LedgerPosting{
			{AccountId: rule.AccountId, Amount: rule.Amount, Sort: 0},
			{AccountId: 0, Amount: rule.Amount.Neg(), CategoryId: rule.CategoryId, Sort: 1},
		}
		return tx
	}
	tx.Postings = []*LedgerPosting{
		{AccountId: rule.AccountId, Amount: rule.Amount.Neg(), Sort: 0},
		{AccountId: 0, Amount: rule.Amount, CategoryId: rule.CategoryId, Sort: 1},
	}
	return tx
}
