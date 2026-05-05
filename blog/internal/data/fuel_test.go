package data

import (
	"context"
	"testing"

	"blog/internal/biz"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupFuelTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	if err := db.AutoMigrate(&FuelVehicle{}, &RefuelRecord{}); err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}
	return db
}

func TestFuelVehicleRepo_ListByUserIdScopesAndPaginates(t *testing.T) {
	db := setupFuelTestDB(t)
	repo := &FuelVehicleRepo{data: &Data{db: db}, log: log.NewHelper(log.DefaultLogger)}
	ctx := context.Background()

	for _, vehicle := range []*biz.FuelVehicle{
		{Name: "A Car", Brand: "A", TankCapacity: decimal.NewFromInt(50), UserId: "user-123"},
		{Name: "B Car", Brand: "B", TankCapacity: decimal.NewFromInt(60), UserId: "user-123"},
		{Name: "Other Car", Brand: "C", TankCapacity: decimal.NewFromInt(70), UserId: "user-456"},
	} {
		_, err := repo.Save(ctx, vehicle)
		assert.NoError(t, err)
	}

	items, total, err := repo.ListByUserId(ctx, "user-123", &biz.FuelVehicleListQuery{Page: 1, PageSize: 1})

	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, items, 1)
	assert.NotEqual(t, "Other Car", items[0].Name)
}

func TestFuelVehicleRepo_CountRefuelRecordByVehicleId(t *testing.T) {
	db := setupFuelTestDB(t)
	vehicleRepo := &FuelVehicleRepo{data: &Data{db: db}, log: log.NewHelper(log.DefaultLogger)}
	recordRepo := &RefuelRecordRepo{data: &Data{db: db}, log: log.NewHelper(log.DefaultLogger)}
	ctx := context.Background()

	vehicleId, err := vehicleRepo.Save(ctx, &biz.FuelVehicle{Name: "Car", UserId: "user-123"})
	assert.NoError(t, err)
	_, err = recordRepo.Save(ctx, &biz.RefuelRecord{
		VehicleId:  int64(vehicleId),
		RefuelTime: "2026-01-01 00:00:00",
		Odometer:   decimal.NewFromInt(1000),
		Volume:     decimal.NewFromInt(30),
		UnitPrice:  decimal.NewFromInt(7),
		Amount:     decimal.NewFromInt(210),
		IsFull:     true,
		UserId:     "user-123",
	})
	assert.NoError(t, err)

	count, err := vehicleRepo.CountRefuelRecordByVehicleId(ctx, vehicleId)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestRefuelRecordRepo_ListByUserIdAndVehicleIdOrdersByRefuelTimeDesc(t *testing.T) {
	db := setupFuelTestDB(t)
	vehicleRepo := &FuelVehicleRepo{data: &Data{db: db}, log: log.NewHelper(log.DefaultLogger)}
	recordRepo := &RefuelRecordRepo{data: &Data{db: db}, log: log.NewHelper(log.DefaultLogger)}
	ctx := context.Background()

	vehicleId, err := vehicleRepo.Save(ctx, &biz.FuelVehicle{Name: "Car", UserId: "user-123"})
	assert.NoError(t, err)
	for _, record := range []*biz.RefuelRecord{
		{VehicleId: int64(vehicleId), RefuelTime: "2026-01-01 00:00:00", Odometer: decimal.NewFromInt(1000), Volume: decimal.NewFromInt(30), Amount: decimal.NewFromInt(210), UserId: "user-123"},
		{VehicleId: int64(vehicleId), RefuelTime: "2026-02-01 00:00:00", Odometer: decimal.NewFromInt(1500), Volume: decimal.NewFromInt(35), Amount: decimal.NewFromInt(245), UserId: "user-123"},
	} {
		_, err := recordRepo.Save(ctx, record)
		assert.NoError(t, err)
	}

	items, total, err := recordRepo.ListByUserIdAndVehicleId(ctx, "user-123", vehicleId, &biz.RefuelRecordListQuery{Page: 1, PageSize: 10})

	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, items, 2)
	assert.Equal(t, "2026-02-01 00:00:00", items[0].RefuelTime)
}
