package service

import (
	"context"
	"testing"

	pb "blog/api/ledger/v1"
	"blog/internal/biz"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func mustLedgerDecimal(t *testing.T, value string) *decimal.Decimal {
	t.Helper()
	parsed, err := decimal.NewFromString(value)
	assert.NoError(t, err)
	return &parsed
}

func ledgerAccountFixture(creditLimit *decimal.Decimal, billingDay int32) *biz.LedgerAccount {
	account := &biz.LedgerAccount{
		Id: 1, Name: "卡", Type: "liability", Subtype: "credit_card",
		CreditLimit: creditLimit,
	}
	if billingDay > 0 {
		account.BillingDay = &billingDay
	}
	return account
}

func TestLedgerTransactionFromSaveRequest_AcceptsValid(t *testing.T) {
	req := &pb.LedgerTransactionSaveRequest{
		Type:     "expense",
		BookedAt: "2026-08-01 10:00:00",
		Remark:   "午餐",
		Postings: []*pb.LedgerPosting{
			{AccountId: 0, Amount: "35.50", CategoryId: 5},
			{AccountId: 1, Amount: "-35.50"},
		},
	}

	tx, err := ledgerTransactionFromSaveRequest(req)

	assert.NoError(t, err)
	assert.Equal(t, "expense", tx.Type)
	assert.Len(t, tx.Postings, 2)
	assert.True(t, tx.Postings[0].Amount.String() == "35.5")
	assert.True(t, tx.Postings[1].Amount.IsNegative())
}

func TestLedgerTransactionFromSaveRequest_RejectsInvalidAmount(t *testing.T) {
	req := &pb.LedgerTransactionSaveRequest{
		Type:     "expense",
		BookedAt: "2026-08-01 10:00:00",
		Postings: []*pb.LedgerPosting{
			{AccountId: 0, Amount: "abc"},
			{AccountId: 1, Amount: "-100"},
		},
	}

	_, err := ledgerTransactionFromSaveRequest(req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "分录金额不合法")
}

func TestLedgerTransactionFromSaveRequest_RejectsEmptyPostings(t *testing.T) {
	req := &pb.LedgerTransactionSaveRequest{
		Type:     "transfer",
		BookedAt: "2026-08-01 10:00:00",
	}

	_, err := ledgerTransactionFromSaveRequest(req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "交易分录不能为空")
}

func TestLedgerTransactionFromSaveRequest_RejectsBadBookedAt(t *testing.T) {
	for _, bookedAt := range []string{"", "2026-08-01", "2026/08/01 10:00:00"} {
		req := &pb.LedgerTransactionSaveRequest{
			Type:     "transfer",
			BookedAt: bookedAt,
			Postings: []*pb.LedgerPosting{
				{AccountId: 1, Amount: "-100"},
				{AccountId: 2, Amount: "100"},
			},
		}

		_, err := ledgerTransactionFromSaveRequest(req)

		assert.Error(t, err, "bookedAt=%q", bookedAt)
	}
}

func TestLedgerAccountFromSaveRequest_ParsesOptionalFields(t *testing.T) {
	billingDay := int32(5)
	paymentDueDay := int32(20)
	req := &pb.LedgerAccountSaveRequest{
		Name:           "招行信用卡",
		Type:           "liability",
		Subtype:        "credit_card",
		CreditLimit:    "10000.00",
		BillingDay:     &billingDay,
		PaymentDueDay:  &paymentDueDay,
		OpeningBalance: "-2500.50",
	}

	account, opening, err := ledgerAccountFromSaveRequest(req)

	assert.NoError(t, err)
	assert.NotNil(t, account.CreditLimit)
	assert.True(t, account.CreditLimit.Equal(decimal.RequireFromString("10000")))
	assert.NotNil(t, opening)
	assert.True(t, opening.Equal(decimal.RequireFromString("-2500.5")))
}

func TestLedgerAccountFromSaveRequest_RejectsInvalidDecimals(t *testing.T) {
	req := &pb.LedgerAccountSaveRequest{Name: "卡", Type: "liability", Subtype: "credit_card", CreditLimit: "abc"}
	_, _, err := ledgerAccountFromSaveRequest(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "信用额度不合法")

	req = &pb.LedgerAccountSaveRequest{Name: "卡", Type: "asset", Subtype: "cash", OpeningBalance: "xyz"}
	_, _, err = ledgerAccountFromSaveRequest(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "期初余额不合法")
}

func TestLedgerAccountFromSaveRequest_OmitsNilOptionals(t *testing.T) {
	account, opening, err := ledgerAccountFromSaveRequest(&pb.LedgerAccountSaveRequest{
		Name: "现金", Type: "asset", Subtype: "cash",
	})

	assert.NoError(t, err)
	assert.Nil(t, account.CreditLimit)
	assert.Nil(t, account.BillingDay)
	assert.Nil(t, opening)
}

func TestLedgerCategoryFromMessage_RejectsEmptyName(t *testing.T) {
	_, err := ledgerCategoryFromMessage(&pb.LedgerCategory{Name: ""})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "分类名称不能为空")
}

func TestLedgerAccountToReply_MapsNullableFields(t *testing.T) {
	reply := ledgerAccountToReply(nil)
	assert.NotNil(t, reply)

	// 无信用额度时 creditLimit 返回空串，billingDay 不返回
	billing := int32(1)
	credit := mustLedgerDecimal(t, "5000")
	reply = ledgerAccountToReply(ledgerAccountFixture(credit, billing))
	assert.Equal(t, "5000", reply.CreditLimit)
	assert.Equal(t, int32(1), reply.GetBillingDay())

	reply = ledgerAccountToReply(ledgerAccountFixture(nil, 0))
	assert.Equal(t, "", reply.CreditLimit)
	assert.Nil(t, reply.BillingDay)
}

func TestLedgerService_SaveBudgetRejectsInvalidRequest(t *testing.T) {
	svc := NewLedgerService(nil)

	_, err := svc.SaveBudget(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "请求不能为空")

	_, err = svc.SaveBudget(context.Background(), &pb.LedgerBudgetSaveRequest{CategoryId: 0, Month: "2026-08", Amount: "100"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "预算分类不能为空")

	_, err = svc.SaveBudget(context.Background(), &pb.LedgerBudgetSaveRequest{CategoryId: 5, Month: "2026-08", Amount: "abc"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "预算金额不合法")
}

func TestLedgerService_DeleteBudgetRejectsInvalidId(t *testing.T) {
	svc := NewLedgerService(nil)

	reply, err := svc.DeleteBudget(context.Background(), &pb.DeleteLedgerBudgetRequest{Id: 0})

	assert.Error(t, err)
	assert.False(t, reply.Flag)
	assert.Contains(t, err.Error(), "预算 id 不合法")
}

func TestLedgerService_SaveRecurringRejectsInvalidRequest(t *testing.T) {
	svc := NewLedgerService(nil)

	_, err := svc.SaveRecurring(context.Background(), nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "请求不能为空")

	_, err = svc.SaveRecurring(context.Background(), &pb.LedgerRecurringSaveRequest{
		AccountId: 1, CategoryId: 5, Type: "expense", Amount: "abc", DayOfMonth: 1, StartMonth: "2026-08", Enabled: true,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "周期账单金额不合法")
}

func TestLedgerService_DeleteRecurringRejectsInvalidId(t *testing.T) {
	svc := NewLedgerService(nil)

	reply, err := svc.DeleteRecurring(context.Background(), &pb.DeleteLedgerRecurringRequest{Id: 0})

	assert.Error(t, err)
	assert.False(t, reply.Flag)
	assert.Contains(t, err.Error(), "周期账单 id 不合法")
}

func TestLedgerRecurringToReply_MapsAllFields(t *testing.T) {
	reply := ledgerRecurringToReply(nil)
	assert.NotNil(t, reply)

	reply = ledgerRecurringToReply(&biz.LedgerRecurring{
		Id: 3, AccountId: 1, AccountName: "现金", CategoryId: 5, CategoryName: "餐饮",
		Type: "expense", Amount: decimal.RequireFromString("100.50"), Remark: "房租",
		DayOfMonth: 31, StartMonth: "2026-01", LastGeneratedMonth: "2026-07", Enabled: true, NextDate: "2026-08-31",
	})
	assert.Equal(t, int64(3), reply.Id)
	assert.Equal(t, "现金", reply.AccountName)
	assert.Equal(t, "餐饮", reply.CategoryName)
	assert.Equal(t, "100.5", reply.Amount)
	assert.Equal(t, int32(31), reply.DayOfMonth)
	assert.Equal(t, "2026-07", reply.LastGeneratedMonth)
	assert.Equal(t, "2026-08-31", reply.NextDate)
	assert.True(t, reply.Enabled)
}
