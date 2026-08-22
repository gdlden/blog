package service

import (
	"blog/internal/biz"
	"context"
	"errors"
	"time"

	pb "blog/api/ledger/v1"

	"github.com/shopspring/decimal"
)

type LedgerService struct {
	pb.UnimplementedLedgerServer
	uc *biz.LedgerUsecase
}

func NewLedgerService(uc *biz.LedgerUsecase) *LedgerService {
	return &LedgerService{uc: uc}
}

// ---------- 账户 ----------

func (s *LedgerService) SaveAccount(ctx context.Context, req *pb.LedgerAccountSaveRequest) (*pb.SaveLedgerReply, error) {
	account, openingBalance, err := ledgerAccountFromSaveRequest(req)
	if err != nil {
		return nil, err
	}
	id, err := s.uc.CreateAccount(ctx, account, openingBalance)
	if err != nil {
		return nil, err
	}
	return &pb.SaveLedgerReply{Id: int64(id), Message: "save success"}, nil
}

func (s *LedgerService) UpdateAccount(ctx context.Context, req *pb.LedgerAccount) (*pb.SaveLedgerReply, error) {
	if req == nil || req.Id <= 0 {
		return nil, errors.New("账户 id 不合法")
	}
	account, err := ledgerAccountFromMessage(req)
	if err != nil {
		return nil, err
	}
	id, err := s.uc.UpdateAccount(ctx, account)
	if err != nil {
		return nil, err
	}
	return &pb.SaveLedgerReply{Id: int64(id), Message: "update success"}, nil
}

func (s *LedgerService) DeleteAccount(ctx context.Context, req *pb.DeleteLedgerAccountRequest) (*pb.DeleteLedgerReply, error) {
	if req == nil || req.Id <= 0 {
		return &pb.DeleteLedgerReply{Flag: false}, errors.New("账户 id 不合法")
	}
	if err := s.uc.DeleteAccount(ctx, uint(req.Id)); err != nil {
		return &pb.DeleteLedgerReply{Flag: false}, err
	}
	return &pb.DeleteLedgerReply{Flag: true}, nil
}

func (s *LedgerService) GetAccount(ctx context.Context, req *pb.GetLedgerAccountRequest) (*pb.LedgerAccount, error) {
	if req == nil || req.Id <= 0 {
		return nil, errors.New("账户 id 不合法")
	}
	account, err := s.uc.GetAccount(ctx, uint(req.Id))
	if err != nil {
		return nil, err
	}
	return ledgerAccountToReply(account), nil
}

func (s *LedgerService) ListAccounts(ctx context.Context, req *pb.ListLedgerAccountsRequest) (*pb.ListLedgerAccountsReply, error) {
	items, err := s.uc.ListAccounts(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]*pb.LedgerAccount, 0, len(items))
	for _, account := range items {
		list = append(list, ledgerAccountToReply(account))
	}
	return &pb.ListLedgerAccountsReply{List: list}, nil
}

// ---------- 分类 ----------

func (s *LedgerService) SaveCategory(ctx context.Context, req *pb.LedgerCategory) (*pb.SaveLedgerReply, error) {
	category, err := ledgerCategoryFromMessage(req)
	if err != nil {
		return nil, err
	}
	id, err := s.uc.SaveCategory(ctx, category)
	if err != nil {
		return nil, err
	}
	return &pb.SaveLedgerReply{Id: int64(id), Message: "save success"}, nil
}

func (s *LedgerService) UpdateCategory(ctx context.Context, req *pb.LedgerCategory) (*pb.SaveLedgerReply, error) {
	if req == nil || req.Id <= 0 {
		return nil, errors.New("分类 id 不合法")
	}
	category, err := ledgerCategoryFromMessage(req)
	if err != nil {
		return nil, err
	}
	id, err := s.uc.UpdateCategory(ctx, category)
	if err != nil {
		return nil, err
	}
	return &pb.SaveLedgerReply{Id: int64(id), Message: "update success"}, nil
}

func (s *LedgerService) DeleteCategory(ctx context.Context, req *pb.DeleteLedgerCategoryRequest) (*pb.DeleteLedgerReply, error) {
	if req == nil || req.Id <= 0 {
		return &pb.DeleteLedgerReply{Flag: false}, errors.New("分类 id 不合法")
	}
	if err := s.uc.DeleteCategory(ctx, uint(req.Id)); err != nil {
		return &pb.DeleteLedgerReply{Flag: false}, err
	}
	return &pb.DeleteLedgerReply{Flag: true}, nil
}

func (s *LedgerService) ListCategories(ctx context.Context, req *pb.ListLedgerCategoriesRequest) (*pb.ListLedgerCategoriesReply, error) {
	items, err := s.uc.ListCategories(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]*pb.LedgerCategory, 0, len(items))
	for _, category := range items {
		list = append(list, ledgerCategoryToReply(category))
	}
	return &pb.ListLedgerCategoriesReply{List: list}, nil
}

// ---------- 交易 ----------

func (s *LedgerService) SaveTransaction(ctx context.Context, req *pb.LedgerTransactionSaveRequest) (*pb.SaveLedgerReply, error) {
	tx, err := ledgerTransactionFromSaveRequest(req)
	if err != nil {
		return nil, err
	}
	id, err := s.uc.CreateTransaction(ctx, tx)
	if err != nil {
		return nil, err
	}
	return &pb.SaveLedgerReply{Id: int64(id), Message: "save success"}, nil
}

func (s *LedgerService) UpdateTransaction(ctx context.Context, req *pb.LedgerTransaction) (*pb.SaveLedgerReply, error) {
	if req == nil || req.Id <= 0 {
		return nil, errors.New("交易 id 不合法")
	}
	tx, err := ledgerTransactionFromMessage(req)
	if err != nil {
		return nil, err
	}
	id, err := s.uc.UpdateTransaction(ctx, tx)
	if err != nil {
		return nil, err
	}
	return &pb.SaveLedgerReply{Id: int64(id), Message: "update success"}, nil
}

func (s *LedgerService) DeleteTransaction(ctx context.Context, req *pb.DeleteLedgerTransactionRequest) (*pb.DeleteLedgerReply, error) {
	if req == nil || req.Id <= 0 {
		return &pb.DeleteLedgerReply{Flag: false}, errors.New("交易 id 不合法")
	}
	if err := s.uc.DeleteTransaction(ctx, uint(req.Id)); err != nil {
		return &pb.DeleteLedgerReply{Flag: false}, err
	}
	return &pb.DeleteLedgerReply{Flag: true}, nil
}

func (s *LedgerService) GetTransaction(ctx context.Context, req *pb.GetLedgerTransactionRequest) (*pb.LedgerTransaction, error) {
	if req == nil || req.Id <= 0 {
		return nil, errors.New("交易 id 不合法")
	}
	tx, err := s.uc.GetTransaction(ctx, uint(req.Id))
	if err != nil {
		return nil, err
	}
	return ledgerTransactionToReply(tx), nil
}

func (s *LedgerService) PageTransactions(ctx context.Context, req *pb.PageLedgerTransactionsRequest) (*pb.PageLedgerTransactionsReply, error) {
	if req == nil {
		req = &pb.PageLedgerTransactionsRequest{}
	}
	page, pageSize := normalizeServicePage(req.Page, req.PageSize)
	items, total, err := s.uc.PageTransactions(ctx, &biz.LedgerTransactionQuery{
		Page:       page,
		PageSize:   pageSize,
		AccountId:  req.AccountId,
		CategoryId: req.CategoryId,
		Type:       req.Type,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
	})
	if err != nil {
		return nil, err
	}
	list := make([]*pb.LedgerTransaction, 0, len(items))
	for _, tx := range items {
		list = append(list, ledgerTransactionToReply(tx))
	}
	return &pb.PageLedgerTransactionsReply{Page: page, Total: total, List: list}, nil
}

// ---------- 月度统计 ----------

func (s *LedgerService) GetMonthlyStats(ctx context.Context, req *pb.GetLedgerMonthlyStatsRequest) (*pb.LedgerMonthlyStats, error) {
	if req == nil || req.Month == "" {
		return nil, errors.New("月份不能为空")
	}
	stats, err := s.uc.GetMonthlyStats(ctx, req.Month)
	if err != nil {
		return nil, err
	}
	return ledgerMonthlyStatsToReply(stats), nil
}

// ---------- 预算 ----------

func (s *LedgerService) SaveBudget(ctx context.Context, req *pb.LedgerBudgetSaveRequest) (*pb.SaveLedgerReply, error) {
	if req == nil {
		return nil, errors.New("请求不能为空")
	}
	if req.CategoryId <= 0 {
		return nil, errors.New("预算分类不能为空")
	}
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, errors.New("预算金额不合法")
	}
	id, err := s.uc.SaveBudget(ctx, &biz.LedgerBudget{
		CategoryId: req.CategoryId,
		Month:      req.Month,
		Amount:     amount,
	})
	if err != nil {
		return nil, err
	}
	return &pb.SaveLedgerReply{Id: int64(id), Message: "save success"}, nil
}

func (s *LedgerService) DeleteBudget(ctx context.Context, req *pb.DeleteLedgerBudgetRequest) (*pb.DeleteLedgerReply, error) {
	if req == nil || req.Id <= 0 {
		return &pb.DeleteLedgerReply{Flag: false}, errors.New("预算 id 不合法")
	}
	if err := s.uc.DeleteBudget(ctx, uint(req.Id)); err != nil {
		return &pb.DeleteLedgerReply{Flag: false}, err
	}
	return &pb.DeleteLedgerReply{Flag: true}, nil
}

func (s *LedgerService) ListBudgets(ctx context.Context, req *pb.ListLedgerBudgetsRequest) (*pb.ListLedgerBudgetsReply, error) {
	month := ""
	if req != nil {
		month = req.Month
	}
	items, err := s.uc.ListBudgets(ctx, month)
	if err != nil {
		return nil, err
	}
	list := make([]*pb.LedgerBudget, 0, len(items))
	for _, budget := range items {
		list = append(list, &pb.LedgerBudget{
			Id:           budget.Id,
			CategoryId:   budget.CategoryId,
			CategoryName: budget.CategoryName,
			Amount:       budget.Amount.String(),
			Used:         budget.Used.String(),
		})
	}
	return &pb.ListLedgerBudgetsReply{List: list}, nil
}

// ---------- 余额走势 ----------

func (s *LedgerService) GetBalanceTrend(ctx context.Context, req *pb.GetLedgerBalanceTrendRequest) (*pb.LedgerBalanceTrend, error) {
	if req == nil {
		req = &pb.GetLedgerBalanceTrendRequest{}
	}
	if req.AccountId < 0 {
		return nil, errors.New("账户 id 不合法")
	}
	points, err := s.uc.GetBalanceTrend(ctx, req.AccountId, req.StartTime, req.EndTime)
	if err != nil {
		return nil, err
	}
	list := make([]*pb.LedgerBalancePoint, 0, len(points))
	for _, point := range points {
		list = append(list, &pb.LedgerBalancePoint{
			Date:    point.Date,
			Balance: point.Balance.String(),
		})
	}
	return &pb.LedgerBalanceTrend{Points: list}, nil
}

// ---------- 周期账单 ----------

func (s *LedgerService) SaveRecurring(ctx context.Context, req *pb.LedgerRecurringSaveRequest) (*pb.SaveLedgerReply, error) {
	if req == nil {
		return nil, errors.New("请求不能为空")
	}
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, errors.New("周期账单金额不合法")
	}
	id, err := s.uc.SaveRecurring(ctx, &biz.LedgerRecurring{
		Id:         int64(req.Id),
		AccountId:  req.AccountId,
		CategoryId: req.CategoryId,
		Type:       req.Type,
		Amount:     amount,
		Remark:     req.Remark,
		DayOfMonth: req.DayOfMonth,
		StartMonth: req.StartMonth,
		Enabled:    req.Enabled,
	})
	if err != nil {
		return nil, err
	}
	return &pb.SaveLedgerReply{Id: int64(id), Message: "save success"}, nil
}

func (s *LedgerService) DeleteRecurring(ctx context.Context, req *pb.DeleteLedgerRecurringRequest) (*pb.DeleteLedgerReply, error) {
	if req == nil || req.Id <= 0 {
		return &pb.DeleteLedgerReply{Flag: false}, errors.New("周期账单 id 不合法")
	}
	if err := s.uc.DeleteRecurring(ctx, uint(req.Id)); err != nil {
		return &pb.DeleteLedgerReply{Flag: false}, err
	}
	return &pb.DeleteLedgerReply{Flag: true}, nil
}

func (s *LedgerService) ListRecurring(ctx context.Context, req *pb.ListLedgerRecurringRequest) (*pb.ListLedgerRecurringReply, error) {
	items, err := s.uc.ListRecurring(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]*pb.LedgerRecurring, 0, len(items))
	for _, rule := range items {
		list = append(list, ledgerRecurringToReply(rule))
	}
	return &pb.ListLedgerRecurringReply{List: list}, nil
}

func (s *LedgerService) ApplyRecurring(ctx context.Context, req *pb.ApplyLedgerRecurringRequest) (*pb.ApplyLedgerRecurringReply, error) {
	created, err := s.uc.ApplyRecurring(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.ApplyLedgerRecurringReply{Created: created}, nil
}

// ---------- 映射与解析 ----------
func ledgerAccountFromSaveRequest(req *pb.LedgerAccountSaveRequest) (*biz.LedgerAccount, *decimal.Decimal, error) {
	if req == nil {
		return nil, nil, errors.New("请求不能为空")
	}
	if req.Name == "" {
		return nil, nil, errors.New("账户名称不能为空")
	}
	creditLimit, err := parseNullableDecimal(req.CreditLimit, "信用额度不合法")
	if err != nil {
		return nil, nil, err
	}
	var openingBalance *decimal.Decimal
	if req.OpeningBalance != "" {
		value, err := decimal.NewFromString(req.OpeningBalance)
		if err != nil {
			return nil, nil, errors.New("期初余额不合法")
		}
		openingBalance = &value
	}
	account := &biz.LedgerAccount{
		Name:          req.Name,
		Type:          req.Type,
		Subtype:       req.Subtype,
		CreditLimit:   creditLimit,
		BillingDay:    req.BillingDay,
		PaymentDueDay: req.PaymentDueDay,
		Remark:        req.Remark,
	}
	return account, openingBalance, nil
}

func ledgerAccountFromMessage(req *pb.LedgerAccount) (*biz.LedgerAccount, error) {
	if req == nil {
		return nil, errors.New("请求不能为空")
	}
	if req.Name == "" {
		return nil, errors.New("账户名称不能为空")
	}
	creditLimit, err := parseNullableDecimal(req.CreditLimit, "信用额度不合法")
	if err != nil {
		return nil, err
	}
	return &biz.LedgerAccount{
		Id:            req.Id,
		Name:          req.Name,
		Type:          req.Type,
		Subtype:       req.Subtype,
		CreditLimit:   creditLimit,
		BillingDay:    req.BillingDay,
		PaymentDueDay: req.PaymentDueDay,
		Remark:        req.Remark,
		Sort:          req.Sort,
		Archived:      req.Archived,
	}, nil
}

func ledgerCategoryFromMessage(req *pb.LedgerCategory) (*biz.LedgerCategory, error) {
	if req == nil {
		return nil, errors.New("请求不能为空")
	}
	if req.Name == "" {
		return nil, errors.New("分类名称不能为空")
	}
	return &biz.LedgerCategory{
		Id:        req.Id,
		ParentId:  req.ParentId,
		Name:      req.Name,
		Direction: req.Direction,
		Sort:      req.Sort,
	}, nil
}

func ledgerTransactionFromSaveRequest(req *pb.LedgerTransactionSaveRequest) (*biz.LedgerTransaction, error) {
	if req == nil {
		return nil, errors.New("请求不能为空")
	}
	postings, err := ledgerPostingsFromRequest(req.Postings)
	if err != nil {
		return nil, err
	}
	if err := validateLedgerBookedAt(req.BookedAt); err != nil {
		return nil, err
	}
	return &biz.LedgerTransaction{
		Type:     req.Type,
		BookedAt: req.BookedAt,
		Remark:   req.Remark,
		Postings: postings,
	}, nil
}

func ledgerTransactionFromMessage(req *pb.LedgerTransaction) (*biz.LedgerTransaction, error) {
	if req == nil {
		return nil, errors.New("请求不能为空")
	}
	postings, err := ledgerPostingsFromRequest(req.Postings)
	if err != nil {
		return nil, err
	}
	if err := validateLedgerBookedAt(req.BookedAt); err != nil {
		return nil, err
	}
	return &biz.LedgerTransaction{
		Id:       req.Id,
		Type:     req.Type,
		BookedAt: req.BookedAt,
		Remark:   req.Remark,
		Postings: postings,
	}, nil
}

func ledgerPostingsFromRequest(items []*pb.LedgerPosting) ([]*biz.LedgerPosting, error) {
	if len(items) == 0 {
		return nil, errors.New("交易分录不能为空")
	}
	postings := make([]*biz.LedgerPosting, 0, len(items))
	for _, item := range items {
		if item == nil {
			return nil, errors.New("交易分录不能为空")
		}
		if item.AccountId < 0 {
			return nil, errors.New("分录账户 id 不合法")
		}
		amount, err := decimal.NewFromString(item.Amount)
		if err != nil {
			return nil, errors.New("分录金额不合法: " + item.Amount)
		}
		if item.CategoryId < 0 {
			return nil, errors.New("分录分类 id 不合法")
		}
		postings = append(postings, &biz.LedgerPosting{
			Id:         item.Id,
			AccountId:  item.AccountId,
			Amount:     amount,
			CategoryId: item.CategoryId,
			Sort:       item.Sort,
		})
	}
	return postings, nil
}

func validateLedgerBookedAt(bookedAt string) error {
	if bookedAt == "" {
		return errors.New("记账时间不能为空")
	}
	if _, err := time.ParseInLocation("2006-01-02 15:04:05", bookedAt, time.Local); err != nil {
		return errors.New("记账时间格式必须为 YYYY-MM-DD HH:mm:ss")
	}
	return nil
}

func parseNullableDecimal(value string, message string) (*decimal.Decimal, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := decimal.NewFromString(value)
	if err != nil {
		return nil, errors.New(message)
	}
	return &parsed, nil
}

func ledgerAccountToReply(account *biz.LedgerAccount) *pb.LedgerAccount {
	if account == nil {
		return &pb.LedgerAccount{}
	}
	reply := &pb.LedgerAccount{
		Id:            account.Id,
		Name:          account.Name,
		Type:          account.Type,
		Subtype:       account.Subtype,
		BillingDay:    account.BillingDay,
		PaymentDueDay: account.PaymentDueDay,
		Remark:        account.Remark,
		Sort:          account.Sort,
		Archived:      account.Archived,
		IsSystem:      account.IsSystem,
		Balance:       account.Balance.String(),
	}
	if account.CreditLimit != nil {
		reply.CreditLimit = account.CreditLimit.String()
	}
	return reply
}

func ledgerCategoryToReply(category *biz.LedgerCategory) *pb.LedgerCategory {
	if category == nil {
		return &pb.LedgerCategory{}
	}
	return &pb.LedgerCategory{
		Id:        category.Id,
		ParentId:  category.ParentId,
		Name:      category.Name,
		Direction: category.Direction,
		Sort:      category.Sort,
		IsSystem:  category.IsSystem,
	}
}

func ledgerTransactionToReply(tx *biz.LedgerTransaction) *pb.LedgerTransaction {
	if tx == nil {
		return &pb.LedgerTransaction{}
	}
	reply := &pb.LedgerTransaction{
		Id:        tx.Id,
		Type:      tx.Type,
		BookedAt:  tx.BookedAt,
		Remark:    tx.Remark,
		CreatedAt: tx.CreatedAt,
		Postings:  make([]*pb.LedgerPosting, 0, len(tx.Postings)),
	}
	for _, posting := range tx.Postings {
		reply.Postings = append(reply.Postings, &pb.LedgerPosting{
			Id:         posting.Id,
			AccountId:  posting.AccountId,
			Amount:     posting.Amount.String(),
			CategoryId: posting.CategoryId,
			Sort:       posting.Sort,
		})
	}
	return reply
}

func ledgerMonthlyStatsToReply(stats *biz.LedgerMonthlyStats) *pb.LedgerMonthlyStats {
	if stats == nil {
		return &pb.LedgerMonthlyStats{}
	}
	return &pb.LedgerMonthlyStats{
		Month:             stats.Month,
		TotalExpense:      stats.TotalExpense.String(),
		TotalIncome:       stats.TotalIncome.String(),
		ExpenseByCategory: ledgerCategoryAmountsToReply(stats.ExpenseByCategory),
		IncomeByCategory:  ledgerCategoryAmountsToReply(stats.IncomeByCategory),
	}
}

func ledgerCategoryAmountsToReply(items []*biz.LedgerCategoryAmount) []*pb.LedgerCategoryAmount {
	result := make([]*pb.LedgerCategoryAmount, 0, len(items))
	for _, item := range items {
		result = append(result, &pb.LedgerCategoryAmount{
			CategoryId:   item.CategoryId,
			CategoryName: item.CategoryName,
			Amount:       item.Amount.String(),
		})
	}
	return result
}

func ledgerRecurringToReply(rule *biz.LedgerRecurring) *pb.LedgerRecurring {
	if rule == nil {
		return &pb.LedgerRecurring{}
	}
	return &pb.LedgerRecurring{
		Id:                 rule.Id,
		AccountId:          rule.AccountId,
		AccountName:        rule.AccountName,
		CategoryId:         rule.CategoryId,
		CategoryName:       rule.CategoryName,
		Type:               rule.Type,
		Amount:             rule.Amount.String(),
		Remark:             rule.Remark,
		DayOfMonth:         rule.DayOfMonth,
		StartMonth:         rule.StartMonth,
		LastGeneratedMonth: rule.LastGeneratedMonth,
		Enabled:            rule.Enabled,
		NextDate:           rule.NextDate,
	}
}
