package biz

import (
	"context"
	"errors"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

// mockDebtRepo is a manual mock implementation of DebtRepo interface
type mockDebtRepo struct {
	saveFunc                    func(context.Context, *Debt) (uint, error)
	updateFunc                  func(context.Context, *Debt) error
	deleteByUserIdAndDebtIdFunc func(context.Context, string, uint) error
	countDetailFunc             func(context.Context, uint) (int64, error)
	findByIDFunc                func(context.Context, int64) (*Debt, error)
	findByUserIdAndDebtIdFunc   func(context.Context, string, uint) (*Debt, error)
	listByUserIdFunc            func(context.Context, string, *DebtListQuery) ([]*Debt, int64, error)
	listByHelloFunc             func(context.Context, string) ([]*Debt, error)
	listAllFunc                 func(context.Context) ([]*Debt, error)
}

func (m *mockDebtRepo) Save(ctx context.Context, d *Debt) (uint, error) {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, d)
	}
	return 0, nil
}

func (m *mockDebtRepo) Update(ctx context.Context, d *Debt) error {
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

func (m *mockDebtRepo) FindByID(ctx context.Context, id int64) (*Debt, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *mockDebtRepo) FindByUserIdAndDebtId(ctx context.Context, userId string, debtId uint) (*Debt, error) {
	if m.findByUserIdAndDebtIdFunc != nil {
		return m.findByUserIdAndDebtIdFunc(ctx, userId, debtId)
	}
	return nil, errors.New("not found")
}

func (m *mockDebtRepo) ListByUserId(ctx context.Context, userId string, query *DebtListQuery) ([]*Debt, int64, error) {
	if m.listByUserIdFunc != nil {
		return m.listByUserIdFunc(ctx, userId, query)
	}
	return nil, 0, nil
}

func (m *mockDebtRepo) ListByHello(ctx context.Context, hello string) ([]*Debt, error) {
	if m.listByHelloFunc != nil {
		return m.listByHelloFunc(ctx, hello)
	}
	return nil, nil
}

func (m *mockDebtRepo) ListAll(ctx context.Context) ([]*Debt, error) {
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

func TestDebtUsecase_CreateDebt(t *testing.T) {
	logger := log.With(log.DefaultLogger, "ts", log.DefaultTimestamp, "caller", log.DefaultCaller)

	tests := []struct {
		name           string
		ctx            context.Context
		debt           *Debt
		mockFn         func(ctx context.Context, d *Debt) (uint, error)
		wantID         uint
		wantUserId     string
		wantErr        bool
		wantErrContain string
	}{
		{
			name: "create debt successfully",
			ctx:  withUser(context.Background(), "user-123"),
			debt: &Debt{
				Name:   "Test Debt",
				Amount: decimal.NewFromInt(1000),
				Apr:    decimal.NewFromFloat(0.05),
			},
			mockFn: func(ctx context.Context, d *Debt) (uint, error) {
				// Verify UserId was set by the usecase
				if d.UserId != "user-123" {
					t.Errorf("expected UserId to be 'user-123', got %s", d.UserId)
				}
				return 1, nil
			},
			wantID:     1,
			wantUserId: "user-123",
			wantErr:    false,
		},
		{
			name: "create debt with save error",
			ctx:  withUser(context.Background(), "user-123"),
			debt: &Debt{
				Name: "Error Debt",
			},
			mockFn: func(ctx context.Context, d *Debt) (uint, error) {
				return 0, errors.New("database error")
			},
			wantID:         0,
			wantErr:        true,
			wantErrContain: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockDebtRepo{
				saveFunc: tt.mockFn,
			}
			uc := NewDebtUsecase(mock, logger)

			got, err := uc.CreateDebt(tt.ctx, tt.debt)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrContain != "" {
					assert.Contains(t, err.Error(), tt.wantErrContain)
				}
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantID, got)
		})
	}
}

func TestDebtUsecase_CreateDebt_Unauthorized(t *testing.T) {
	logger := log.With(log.DefaultLogger, "ts", log.DefaultTimestamp, "caller", log.DefaultCaller)
	mock := &mockDebtRepo{}
	uc := NewDebtUsecase(mock, logger)

	// Context without user
	ctx := context.Background()
	debt := &Debt{
		Name:   "Test Debt",
		Amount: decimal.NewFromInt(1000),
	}

	_, err := uc.CreateDebt(ctx, debt)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "未登录")
}

func TestDebtUsecase_Edit(t *testing.T) {
	logger := log.With(log.DefaultLogger, "ts", log.DefaultTimestamp, "caller", log.DefaultCaller)

	mock := &mockDebtRepo{
		findByUserIdAndDebtIdFunc: func(ctx context.Context, userId string, debtId uint) (*Debt, error) {
			return &Debt{
				Id:     1,
				Name:   "Original Debt",
				UserId: "user-123",
				Amount: decimal.NewFromInt(1000),
			}, nil
		},
		updateFunc: func(ctx context.Context, d *Debt) error {
			// Verify UserId is preserved
			if d.UserId != "user-123" {
				t.Errorf("expected UserId to be 'user-123', got %s", d.UserId)
			}
			return nil
		},
	}

	uc := NewDebtUsecase(mock, logger)
	ctx := withUser(context.Background(), "user-123")

	debt := &Debt{
		Id:   1,
		Name: "Updated Debt",
	}

	got, err := uc.Edit(ctx, debt)
	assert.NoError(t, err)
	assert.Equal(t, uint(1), got)
}

func TestDebtUsecase_Edit_NotOwner(t *testing.T) {
	logger := log.With(log.DefaultLogger, "ts", log.DefaultTimestamp, "caller", log.DefaultCaller)

	mock := &mockDebtRepo{
		findByUserIdAndDebtIdFunc: func(ctx context.Context, userId string, debtId uint) (*Debt, error) {
			// Return not found for wrong user
			return nil, errors.New("record not found")
		},
	}

	uc := NewDebtUsecase(mock, logger)
	// Wrong user context
	ctx := withUser(context.Background(), "user-456")

	debt := &Debt{
		Id:   1,
		Name: "Updated Debt",
	}

	_, err := uc.Edit(ctx, debt)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDebtUsecase_Delete(t *testing.T) {
	logger := log.With(log.DefaultLogger, "ts", log.DefaultTimestamp, "caller", log.DefaultCaller)

	mock := &mockDebtRepo{
		findByUserIdAndDebtIdFunc: func(ctx context.Context, userId string, debtId uint) (*Debt, error) {
			return &Debt{
				Id:     1,
				Name:   "Test Debt",
				UserId: "user-123",
			}, nil
		},
		countDetailFunc: func(ctx context.Context, debtId uint) (int64, error) {
			return 0, nil // No details
		},
		deleteByUserIdAndDebtIdFunc: func(ctx context.Context, userId string, debtId uint) error {
			return nil
		},
	}

	uc := NewDebtUsecase(mock, logger)
	ctx := withUser(context.Background(), "user-123")

	err := uc.Delete(ctx, 1)
	assert.NoError(t, err)
}

func TestDebtUsecase_Delete_WithDetails(t *testing.T) {
	logger := log.With(log.DefaultLogger, "ts", log.DefaultTimestamp, "caller", log.DefaultCaller)

	mock := &mockDebtRepo{
		findByUserIdAndDebtIdFunc: func(ctx context.Context, userId string, debtId uint) (*Debt, error) {
			return &Debt{
				Id:     1,
				Name:   "Test Debt",
				UserId: "user-123",
			}, nil
		},
		countDetailFunc: func(ctx context.Context, debtId uint) (int64, error) {
			return 2, nil // Has 2 details
		},
	}

	uc := NewDebtUsecase(mock, logger)
	ctx := withUser(context.Background(), "user-123")

	err := uc.Delete(ctx, 1)
	assert.Error(t, err)
	assert.Equal(t, "存在明细，禁止删除", err.Error())
}

func TestDebtUsecase_Delete_NotOwner(t *testing.T) {
	logger := log.With(log.DefaultLogger, "ts", log.DefaultTimestamp, "caller", log.DefaultCaller)

	mock := &mockDebtRepo{
		findByUserIdAndDebtIdFunc: func(ctx context.Context, userId string, debtId uint) (*Debt, error) {
			// Return not found for wrong user
			return nil, errors.New("record not found")
		},
	}

	uc := NewDebtUsecase(mock, logger)
	// Wrong user context
	ctx := withUser(context.Background(), "user-456")

	err := uc.Delete(ctx, 1)
	assert.Error(t, err)
}

func TestDebtUsecase_GetDebt(t *testing.T) {
	logger := log.With(log.DefaultLogger, "ts", log.DefaultTimestamp, "caller", log.DefaultCaller)

	tests := []struct {
		name    string
		debtID  uint
		mockFn  func(ctx context.Context, userId string, debtId uint) (*Debt, error)
		want    *Debt
		wantErr bool
	}{
		{
			name:   "get debt successfully",
			debtID: 1,
			mockFn: func(ctx context.Context, userId string, debtId uint) (*Debt, error) {
				return &Debt{
					Id:     1,
					Name:   "Test Debt",
					UserId: "user-123",
					Amount: decimal.NewFromInt(5000),
				}, nil
			},
			want: &Debt{
				Id:     1,
				Name:   "Test Debt",
				UserId: "user-123",
				Amount: decimal.NewFromInt(5000),
			},
			wantErr: false,
		},
		{
			name:   "get non-existent debt",
			debtID: 999,
			mockFn: func(ctx context.Context, userId string, debtId uint) (*Debt, error) {
				return nil, errors.New("record not found")
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockDebtRepo{
				findByUserIdAndDebtIdFunc: tt.mockFn,
			}
			uc := NewDebtUsecase(mock, logger)
			ctx := withUser(context.Background(), "user-123")

			got, err := uc.GetDebt(ctx, tt.debtID)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want.Id, got.Id)
			assert.Equal(t, tt.want.Name, got.Name)
		})
	}
}

func TestDebtUsecase_ListDebt(t *testing.T) {
	logger := log.With(log.DefaultLogger, "ts", log.DefaultTimestamp, "caller", log.DefaultCaller)

	mockDebts := []*Debt{
		{Id: 1, Name: "Debt 1", UserId: "user-123", Amount: decimal.NewFromInt(1000)},
		{Id: 2, Name: "Debt 2", UserId: "user-123", Amount: decimal.NewFromInt(2000)},
		{Id: 3, Name: "Debt 3", UserId: "user-123", Amount: decimal.NewFromInt(3000)},
	}

	mock := &mockDebtRepo{
		listByUserIdFunc: func(ctx context.Context, userId string, query *DebtListQuery) ([]*Debt, int64, error) {
			return mockDebts, 10, nil
		},
	}

	uc := NewDebtUsecase(mock, logger)
	ctx := withUser(context.Background(), "user-123")

	query := &DebtListQuery{
		Page:     1,
		PageSize: 3,
		Name:     "",
		BankName: "",
		Status:   nil,
	}

	debts, total, err := uc.ListDebt(ctx, query)
	assert.NoError(t, err)
	assert.Len(t, debts, 3)
	assert.Equal(t, int64(10), total)
	assert.Equal(t, int64(1), debts[0].Id)
	assert.Equal(t, "Debt 1", debts[0].Name)
}
