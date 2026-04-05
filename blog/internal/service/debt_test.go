package service

import (
	"context"
	"errors"
	"testing"

	pb "blog/api/debt/v1"
	"blog/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// mockDebtRepo is a manual mock implementation of biz.DebtRepo interface
type mockDebtRepo struct {
	saveFunc                    func(context.Context, *biz.Debt) (uint, error)
	updateFunc                  func(context.Context, *biz.Debt) error
	deleteByUserIdAndDebtIdFunc func(context.Context, string, uint) error
	countDetailFunc             func(context.Context, uint) (int64, error)
	findByIDFunc                func(context.Context, int64) (*biz.Debt, error)
	findByUserIdAndDebtIdFunc   func(context.Context, string, uint) (*biz.Debt, error)
	listByUserIdFunc            func(context.Context, string, *biz.DebtListQuery) ([]*biz.Debt, int64, error)
	listByHelloFunc             func(context.Context, string) ([]*biz.Debt, error)
	listAllFunc                 func(context.Context) ([]*biz.Debt, error)
}

func (m *mockDebtRepo) Save(ctx context.Context, d *biz.Debt) (uint, error) {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, d)
	}
	return 0, nil
}

func (m *mockDebtRepo) Update(ctx context.Context, d *biz.Debt) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, d)
	}
	return nil
}

func (m *mockDebtRepo) DeleteByUserIdAndDebtId(ctx context.Context, userId string, debtId uint) error {
	if m.deleteByUserIdAndDebtIdFunc != nil {
		return m.deleteByUserIdAndDebtIdFunc(ctx, userId, debtId)
	}
	return nil
}

func (m *mockDebtRepo) CountDebtDetailByDebtId(ctx context.Context, debtId uint) (int64, error) {
	if m.countDetailFunc != nil {
		return m.countDetailFunc(ctx, debtId)
	}
	return 0, nil
}

func (m *mockDebtRepo) FindByID(ctx context.Context, id int64) (*biz.Debt, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *mockDebtRepo) FindByUserIdAndDebtId(ctx context.Context, userId string, debtId uint) (*biz.Debt, error) {
	if m.findByUserIdAndDebtIdFunc != nil {
		return m.findByUserIdAndDebtIdFunc(ctx, userId, debtId)
	}
	return nil, errors.New("not found")
}

func (m *mockDebtRepo) ListByUserId(ctx context.Context, userId string, query *biz.DebtListQuery) ([]*biz.Debt, int64, error) {
	if m.listByUserIdFunc != nil {
		return m.listByUserIdFunc(ctx, userId, query)
	}
	return nil, 0, nil
}

func (m *mockDebtRepo) ListByHello(ctx context.Context, hello string) ([]*biz.Debt, error) {
	if m.listByHelloFunc != nil {
		return m.listByHelloFunc(ctx, hello)
	}
	return nil, nil
}

func (m *mockDebtRepo) ListAll(ctx context.Context) ([]*biz.Debt, error) {
	if m.listAllFunc != nil {
		return m.listAllFunc(ctx)
	}
	return nil, nil
}

// withUser creates a context with JWT claims for testing
func withUser(ctx context.Context, userId string) context.Context {
	return jwt.NewContext(ctx, jwtv5.MapClaims{
		"userId": userId,
	})
}

// createDebtServiceWithMock creates a DebtService with a mocked repository
func createDebtServiceWithMock(mockRepo *mockDebtRepo) *DebtService {
	logger := log.With(log.DefaultLogger, "ts", log.DefaultTimestamp, "caller", log.DefaultCaller)
	uc := biz.NewDebtUsecase(mockRepo, logger)
	return NewDebtService(uc)
}

func TestDebtService_CreateDebt(t *testing.T) {
	mock := &mockDebtRepo{
		saveFunc: func(ctx context.Context, d *biz.Debt) (uint, error) {
			// Verify UserId was set by the usecase
			if d.UserId != "user-123" {
				t.Errorf("expected UserId to be 'user-123', got %s", d.UserId)
			}
			return 1, nil
		},
	}

	svc := createDebtServiceWithMock(mock)
	ctx := withUser(context.Background(), "user-123")

	req := &pb.DebtEntity{
		Name:      "Test Debt",
		BankName:  "Test Bank",
		Amount:    "10000.00",
		Tenor:     "12",
		Apr:       "5.5",
		Fee:       "100.00",
		Status:    "1",
		ApplyTime: "2024-01-01 00:00:00",
		EndTime:   "2024-12-31 00:00:00",
	}

	reply, err := svc.CreateDebt(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, reply)
	assert.Equal(t, "1", reply.Id)
	assert.Equal(t, "save success", reply.Message)
}

func TestDebtService_CreateDebt_InvalidAmount(t *testing.T) {
	mock := &mockDebtRepo{}
	svc := createDebtServiceWithMock(mock)
	ctx := withUser(context.Background(), "user-123")

	req := &pb.DebtEntity{
		Name:   "Test Debt",
		Amount: "invalid",
		Tenor:  "12",
		Apr:    "5.5",
		Fee:    "100.00",
		Status: "1",
	}

	reply, err := svc.CreateDebt(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, reply)
}

func TestDebtService_UpdateDebt(t *testing.T) {
	mock := &mockDebtRepo{
		findByUserIdAndDebtIdFunc: func(ctx context.Context, userId string, debtId uint) (*biz.Debt, error) {
			return &biz.Debt{
				Id:     1,
				Name:   "Original Debt",
				UserId: "user-123",
				Amount: decimal.NewFromInt(10000),
			}, nil
		},
		updateFunc: func(ctx context.Context, d *biz.Debt) error {
			return nil
		},
	}

	svc := createDebtServiceWithMock(mock)
	ctx := withUser(context.Background(), "user-123")

	req := &pb.DebtEntity{
		Id:        1,
		Name:      "Updated Debt",
		BankName:  "Test Bank",
		Amount:    "15000.00",
		Tenor:     "12",
		Apr:       "5.5",
		Fee:       "100.00",
		Status:    "1",
		ApplyTime: "2024-01-01 00:00:00",
		EndTime:   "2024-12-31 00:00:00",
	}

	reply, err := svc.UpdateDebt(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, reply)
	assert.Equal(t, "1", reply.Id)
	assert.Equal(t, "update success", reply.Message)
}

func TestDebtService_UpdateDebt_InvalidId(t *testing.T) {
	mock := &mockDebtRepo{}
	svc := createDebtServiceWithMock(mock)
	ctx := withUser(context.Background(), "user-123")

	req := &pb.DebtEntity{
		Id:        0,
		Name:      "Updated Debt",
		Amount:    "15000.00",
		Tenor:     "12",
		Apr:       "5.5",
		Fee:       "100.00",
		Status:    "1",
		ApplyTime: "2024-01-01 00:00:00",
		EndTime:   "2024-12-31 00:00:00",
	}

	reply, err := svc.UpdateDebt(ctx, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid debt id")
	assert.Nil(t, reply)
}

func TestDebtService_DeleteDebt(t *testing.T) {
	mock := &mockDebtRepo{
		findByUserIdAndDebtIdFunc: func(ctx context.Context, userId string, debtId uint) (*biz.Debt, error) {
			return &biz.Debt{
				Id:     1,
				Name:   "Test Debt",
				UserId: "user-123",
			}, nil
		},
		countDetailFunc: func(ctx context.Context, debtId uint) (int64, error) {
			return 0, nil
		},
		deleteByUserIdAndDebtIdFunc: func(ctx context.Context, userId string, debtId uint) error {
			return nil
		},
	}

	svc := createDebtServiceWithMock(mock)
	ctx := withUser(context.Background(), "user-123")

	req := &pb.DeleteDebtRequest{
		Id: "1",
	}

	reply, err := svc.DeleteDebt(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, reply)
	assert.True(t, reply.Flag)
}

func TestDebtService_DeleteDebt_InvalidId(t *testing.T) {
	mock := &mockDebtRepo{}
	svc := createDebtServiceWithMock(mock)
	ctx := withUser(context.Background(), "user-123")

	req := &pb.DeleteDebtRequest{
		Id: "",
	}

	reply, err := svc.DeleteDebt(ctx, req)

	assert.Error(t, err)
	assert.NotNil(t, reply)
	assert.False(t, reply.Flag)
}

func TestDebtService_GetDebt(t *testing.T) {
	mock := &mockDebtRepo{
		findByUserIdAndDebtIdFunc: func(ctx context.Context, userId string, debtId uint) (*biz.Debt, error) {
			return &biz.Debt{
				Id:       1,
				Name:     "Test Debt",
				UserId:   "user-123",
				Amount:   decimal.NewFromInt(10000),
				BankName: "Test Bank",
				Status:   1,
			}, nil
		},
	}

	svc := createDebtServiceWithMock(mock)
	ctx := withUser(context.Background(), "user-123")

	req := &pb.GetDebtRequest{
		Id: "1",
	}

	reply, err := svc.GetDebt(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, reply)
	assert.Equal(t, int64(1), reply.Id)
	assert.Equal(t, "Test Debt", reply.Name)
	assert.Equal(t, "10000", reply.Amount)
}

func TestDebtService_ListDebt(t *testing.T) {
	mockDebts := []*biz.Debt{
		{Id: 1, Name: "Debt 1", UserId: "user-123", Amount: decimal.NewFromInt(1000)},
		{Id: 2, Name: "Debt 2", UserId: "user-123", Amount: decimal.NewFromInt(2000)},
	}

	mock := &mockDebtRepo{
		listByUserIdFunc: func(ctx context.Context, userId string, query *biz.DebtListQuery) ([]*biz.Debt, int64, error) {
			return mockDebts, 10, nil
		},
	}

	svc := createDebtServiceWithMock(mock)
	ctx := withUser(context.Background(), "user-123")

	req := &pb.ListDebtRequest{
		Page:     1,
		PageSize: 2,
	}

	reply, err := svc.ListDebt(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, reply)
	assert.Len(t, reply.List, 2)
	assert.Equal(t, int64(10), reply.Total)
	assert.Equal(t, int64(1), reply.Page)
}

func TestDebtService_ListDebt_DefaultPagination(t *testing.T) {
	var capturedQuery *biz.DebtListQuery

	mock := &mockDebtRepo{
		listByUserIdFunc: func(ctx context.Context, userId string, query *biz.DebtListQuery) ([]*biz.Debt, int64, error) {
			capturedQuery = query
			return []*biz.Debt{}, 0, nil
		},
	}

	svc := createDebtServiceWithMock(mock)
	ctx := withUser(context.Background(), "user-123")

	// Empty request should apply defaults
	req := &pb.ListDebtRequest{}

	reply, err := svc.ListDebt(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, reply)
	assert.NotNil(t, capturedQuery)
	assert.Equal(t, int64(1), capturedQuery.Page)
	assert.Equal(t, int64(10), capturedQuery.PageSize)
}
