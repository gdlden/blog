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

func setupDebtTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	// Auto migrate the schema for both Debt and DebtDetail
	err = db.AutoMigrate(&Debt{}, &DebtDetail{})
	if err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}
	return db
}

func TestDebtRepo_Save(t *testing.T) {
	db := setupDebtTestDB(t)
	repo := &DebtRepo{
		data: &Data{db: db},
		log:  log.NewHelper(log.DefaultLogger),
	}

	ctx := context.Background()
	debt := &biz.Debt{
		Name:        "Test Debt",
		BankName:    "Test Bank",
		BankAccount: "1234567890",
		ApplyTime:   "2024-01-01 00:00:00",
		EndTime:     "2024-12-31 23:59:59",
		Amount:      decimal.NewFromInt(10000),
		Tenor:       decimal.NewFromInt(12),
		Status:      1,
		Remark:      "Test remark",
		Apr:         decimal.NewFromFloat(5.5),
		Fee:         decimal.NewFromInt(100),
		UserId:      "user-123",
	}

	id, err := repo.Save(ctx, debt)

	assert.NoError(t, err)
	assert.Greater(t, id, uint(0))

	// Verify in database
	var savedDebt Debt
	db.First(&savedDebt, id)
	assert.Equal(t, "Test Debt", savedDebt.Name)
	assert.Equal(t, "Test Bank", savedDebt.BankName)
	assert.Equal(t, "user-123", savedDebt.UserId)
	assert.True(t, savedDebt.Amount.Equal(decimal.NewFromInt(10000)))
}

func TestDebtRepo_FindByUserIdAndDebtId(t *testing.T) {
	db := setupDebtTestDB(t)
	repo := &DebtRepo{
		data: &Data{db: db},
		log:  log.NewHelper(log.DefaultLogger),
	}

	ctx := context.Background()

	// Create a debt
	debt := &biz.Debt{
		Name:        "Find Test Debt",
		BankName:    "Find Bank",
		BankAccount: "9876543210",
		ApplyTime:   "2024-01-01 00:00:00",
		EndTime:     "2024-12-31 23:59:59",
		Amount:      decimal.NewFromInt(5000),
		Tenor:       decimal.NewFromInt(6),
		Status:      1,
		UserId:      "user-123",
	}
	id, _ := repo.Save(ctx, debt)

	// Test finding with correct userId
	found, err := repo.FindByUserIdAndDebtId(ctx, "user-123", id)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, "Find Test Debt", found.Name)
	assert.Equal(t, int64(id), found.Id)

	// Test finding with wrong userId - should return error
	_, err = repo.FindByUserIdAndDebtId(ctx, "user-wrong", id)
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestDebtRepo_ListByUserId(t *testing.T) {
	db := setupDebtTestDB(t)
	repo := &DebtRepo{
		data: &Data{db: db},
		log:  log.NewHelper(log.DefaultLogger),
	}

	ctx := context.Background()

	// Create 5 debts for user-123
	for i := 0; i < 5; i++ {
		debt := &biz.Debt{
			Name:        "Debt for user-123",
			BankName:    "Bank A",
			BankAccount: "1111111111",
			ApplyTime:   "2024-01-01 00:00:00",
			EndTime:     "2024-12-31 23:59:59",
			Amount:      decimal.NewFromInt(1000),
			Tenor:       decimal.NewFromInt(12),
			Status:      1,
			UserId:      "user-123",
		}
		repo.Save(ctx, debt)
	}

	// Create 3 debts for user-456
	for i := 0; i < 3; i++ {
		debt := &biz.Debt{
			Name:        "Debt for user-456",
			BankName:    "Bank B",
			BankAccount: "2222222222",
			ApplyTime:   "2024-01-01 00:00:00",
			EndTime:     "2024-12-31 23:59:59",
			Amount:      decimal.NewFromInt(2000),
			Tenor:       decimal.NewFromInt(6),
			Status:      2,
			UserId:      "user-456",
		}
		repo.Save(ctx, debt)
	}

	// Test pagination: page 1, size 3 for user-123
	debts, total, err := repo.ListByUserId(ctx, "user-123", &biz.DebtListQuery{
		Page:     1,
		PageSize: 3,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, debts, 3)

	// Test with Name filter
	debts, total, err = repo.ListByUserId(ctx, "user-123", &biz.DebtListQuery{
		Page:     1,
		PageSize: 10,
		Name:     "user-123",
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, debts, 5)

	// Test with BankName filter
	debts, total, err = repo.ListByUserId(ctx, "user-456", &biz.DebtListQuery{
		Page:     1,
		PageSize: 10,
		BankName: "Bank B",
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, debts, 3)

	// Test with Status filter
	status := 2
	debts, total, err = repo.ListByUserId(ctx, "user-456", &biz.DebtListQuery{
		Page:     1,
		PageSize: 10,
		Status:   &status,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(3), total)
	assert.Len(t, debts, 3)

	// Test with Status filter for user-123 (status 1)
	status = 1
	debts, total, err = repo.ListByUserId(ctx, "user-123", &biz.DebtListQuery{
		Page:     1,
		PageSize: 10,
		Status:   &status,
	})
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, debts, 5)
}

func TestDebtRepo_Update(t *testing.T) {
	db := setupDebtTestDB(t)
	repo := &DebtRepo{
		data: &Data{db: db},
		log:  log.NewHelper(log.DefaultLogger),
	}

	ctx := context.Background()

	// Create a debt first
	debt := &biz.Debt{
		Name:        "Original Name",
		BankName:    "Original Bank",
		BankAccount: "1234567890",
		ApplyTime:   "2024-01-01 00:00:00",
		EndTime:     "2024-12-31 23:59:59",
		Amount:      decimal.NewFromInt(10000),
		Tenor:       decimal.NewFromInt(12),
		Status:      1,
		Remark:      "Original remark",
		Apr:         decimal.NewFromFloat(5.5),
		Fee:         decimal.NewFromInt(100),
		UserId:      "user-123",
	}
	id, _ := repo.Save(ctx, debt)

	// Update the debt
	updateDebt := &biz.Debt{
		Id:          int64(id),
		Name:        "Updated Name",
		BankName:    "Updated Bank",
		BankAccount: "1234567890",
		ApplyTime:   "2024-02-01 00:00:00",
		EndTime:     "2025-01-31 23:59:59",
		Amount:      decimal.NewFromInt(20000),
		Tenor:       decimal.NewFromInt(24),
		Status:      2,
		Remark:      "Updated remark",
		Apr:         decimal.NewFromFloat(6.5),
		Fee:         decimal.NewFromInt(200),
		UserId:      "user-123",
	}
	err := repo.Update(ctx, updateDebt)
	assert.NoError(t, err)

	// Verify in database
	var updatedDebt Debt
	db.First(&updatedDebt, id)
	assert.Equal(t, "Updated Name", updatedDebt.Name)
	assert.Equal(t, "Updated Bank", updatedDebt.BankName)
	assert.True(t, updatedDebt.Amount.Equal(decimal.NewFromInt(20000)))
	assert.Equal(t, 2, updatedDebt.Status)
}

func TestDebtRepo_DeleteByUserIdAndDebtId(t *testing.T) {
	db := setupDebtTestDB(t)
	repo := &DebtRepo{
		data: &Data{db: db},
		log:  log.NewHelper(log.DefaultLogger),
	}

	ctx := context.Background()

	// Create a debt
	debt := &biz.Debt{
		Name:        "To Delete",
		BankName:    "Delete Bank",
		BankAccount: "9999999999",
		ApplyTime:   "2024-01-01 00:00:00",
		EndTime:     "2024-12-31 23:59:59",
		Amount:      decimal.NewFromInt(5000),
		Tenor:       decimal.NewFromInt(6),
		Status:      1,
		UserId:      "user-123",
	}
	id, _ := repo.Save(ctx, debt)

	// Verify debt exists
	var count int64
	db.Model(&Debt{}).Count(&count)
	assert.Equal(t, int64(1), count)

	// Delete with wrong userId - should fail
	err := repo.DeleteByUserIdAndDebtId(ctx, "user-wrong", id)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no deletable debt found")

	// Verify debt still exists
	db.Model(&Debt{}).Count(&count)
	assert.Equal(t, int64(1), count)

	// Delete with correct userId
	err = repo.DeleteByUserIdAndDebtId(ctx, "user-123", id)
	assert.NoError(t, err)

	// Verify deletion
	db.Model(&Debt{}).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestDebtRepo_CountDebtDetailByDebtId(t *testing.T) {
	db := setupDebtTestDB(t)
	repo := &DebtRepo{
		data: &Data{db: db},
		log:  log.NewHelper(log.DefaultLogger),
	}

	ctx := context.Background()

	// Create a debt
	debt := &biz.Debt{
		Name:        "Debt with Details",
		BankName:    "Detail Bank",
		BankAccount: "1111111111",
		ApplyTime:   "2024-01-01 00:00:00",
		EndTime:     "2024-12-31 23:59:59",
		Amount:      decimal.NewFromInt(10000),
		Tenor:       decimal.NewFromInt(12),
		Status:      1,
		UserId:      "user-123",
	}
	debtId, _ := repo.Save(ctx, debt)

	// Create another debt (without details)
	debt2 := &biz.Debt{
		Name:        "Debt without Details",
		BankName:    "No Detail Bank",
		BankAccount: "2222222222",
		ApplyTime:   "2024-01-01 00:00:00",
		EndTime:     "2024-12-31 23:59:59",
		Amount:      decimal.NewFromInt(5000),
		Tenor:       decimal.NewFromInt(6),
		Status:      1,
		UserId:      "user-123",
	}
	debtId2, _ := repo.Save(ctx, debt2)

	// Create 2 DebtDetail records for the first debt
	detail1 := &DebtDetail{
		DebtId:      debtId,
		PostingDate: time.Now(),
		Principal:   decimal.NewFromInt(1000),
		Interest:    decimal.NewFromInt(50),
		Period:      1,
	}
	db.Create(&detail1)

	detail2 := &DebtDetail{
		DebtId:      debtId,
		PostingDate: time.Now(),
		Principal:   decimal.NewFromInt(1000),
		Interest:    decimal.NewFromInt(45),
		Period:      2,
	}
	db.Create(&detail2)

	// Test count for debt with 2 details
	count, err := repo.CountDebtDetailByDebtId(ctx, debtId)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), count)

	// Test count for debt with 0 details
	count, err = repo.CountDebtDetailByDebtId(ctx, debtId2)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// Test count for non-existent debt
	count, err = repo.CountDebtDetailByDebtId(ctx, 99999)
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestDebtRepo_FindByID(t *testing.T) {
	db := setupDebtTestDB(t)
	repo := &DebtRepo{
		data: &Data{db: db},
		log:  log.NewHelper(log.DefaultLogger),
	}

	ctx := context.Background()

	// Create a debt
	debt := &biz.Debt{
		Name:        "Find By ID Test",
		BankName:    "Find ID Bank",
		BankAccount: "3333333333",
		ApplyTime:   "2024-01-01 00:00:00",
		EndTime:     "2024-12-31 23:59:59",
		Amount:      decimal.NewFromInt(15000),
		Tenor:       decimal.NewFromInt(18),
		Status:      1,
		UserId:      "user-123",
	}
	id, _ := repo.Save(ctx, debt)

	// Find by ID
	found, err := repo.FindByID(ctx, int64(id))
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, "Find By ID Test", found.Name)
	assert.Equal(t, int64(id), found.Id)
	assert.True(t, found.Amount.Equal(decimal.NewFromInt(15000)))

	// Find non-existent ID
	_, err = repo.FindByID(ctx, 99999)
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}
