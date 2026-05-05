package biz

import (
	"context"
	"errors"
	"testing"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

type mockFuelVehicleRepo struct {
	saveFunc       func(context.Context, *FuelVehicle) (uint, error)
	updateFunc     func(context.Context, *FuelVehicle) error
	deleteFunc     func(context.Context, string, uint) error
	findByUserFunc func(context.Context, string, uint) (*FuelVehicle, error)
	listFunc       func(context.Context, string, *FuelVehicleListQuery) ([]*FuelVehicle, int64, error)
	countFunc      func(context.Context, uint) (int64, error)
}

func (m *mockFuelVehicleRepo) Save(ctx context.Context, vehicle *FuelVehicle) (uint, error) {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, vehicle)
	}
	return 0, nil
}

func (m *mockFuelVehicleRepo) Update(ctx context.Context, vehicle *FuelVehicle) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, vehicle)
	}
	return nil
}

func (m *mockFuelVehicleRepo) DeleteByUserIdAndVehicleId(ctx context.Context, userId string, id uint) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, userId, id)
	}
	return nil
}

func (m *mockFuelVehicleRepo) FindByUserIdAndVehicleId(ctx context.Context, userId string, id uint) (*FuelVehicle, error) {
	if m.findByUserFunc != nil {
		return m.findByUserFunc(ctx, userId, id)
	}
	return nil, errors.New("not found")
}

func (m *mockFuelVehicleRepo) ListByUserId(ctx context.Context, userId string, query *FuelVehicleListQuery) ([]*FuelVehicle, int64, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, userId, query)
	}
	return nil, 0, nil
}

func (m *mockFuelVehicleRepo) CountRefuelRecordByVehicleId(ctx context.Context, vehicleId uint) (int64, error) {
	if m.countFunc != nil {
		return m.countFunc(ctx, vehicleId)
	}
	return 0, nil
}

type mockRefuelRecordRepo struct {
	listAllFunc func(context.Context, string, uint) ([]*RefuelRecord, error)
}

func (m *mockRefuelRecordRepo) Save(ctx context.Context, record *RefuelRecord) (uint, error) {
	return 0, nil
}

func (m *mockRefuelRecordRepo) Update(ctx context.Context, record *RefuelRecord) error {
	return nil
}

func (m *mockRefuelRecordRepo) DeleteByUserIdAndRecordId(ctx context.Context, userId string, id uint) error {
	return nil
}

func (m *mockRefuelRecordRepo) FindByUserIdAndRecordId(ctx context.Context, userId string, id uint) (*RefuelRecord, error) {
	return nil, nil
}

func (m *mockRefuelRecordRepo) ListByUserIdAndVehicleId(ctx context.Context, userId string, vehicleId uint, query *RefuelRecordListQuery) ([]*RefuelRecord, int64, error) {
	return nil, 0, nil
}

func (m *mockRefuelRecordRepo) ListAllByUserIdAndVehicleId(ctx context.Context, userId string, vehicleId uint) ([]*RefuelRecord, error) {
	if m.listAllFunc != nil {
		return m.listAllFunc(ctx, userId, vehicleId)
	}
	return nil, nil
}

func TestFuelUsecase_DeleteVehicleBlocksWhenRecordsExist(t *testing.T) {
	uc := NewFuelUsecase(&mockFuelVehicleRepo{
		findByUserFunc: func(ctx context.Context, userId string, id uint) (*FuelVehicle, error) {
			assert.Equal(t, "user-123", userId)
			return &FuelVehicle{Id: 1, UserId: userId}, nil
		},
		countFunc: func(ctx context.Context, vehicleId uint) (int64, error) {
			assert.Equal(t, uint(1), vehicleId)
			return 2, nil
		},
	}, &mockRefuelRecordRepo{}, log.DefaultLogger)

	err := uc.DeleteVehicle(withUser(context.Background(), "user-123"), 1)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "存在加油记录，禁止删除")
}

func TestFuelUsecase_GetStatsCalculatesFullTankIntervals(t *testing.T) {
	uc := NewFuelUsecase(&mockFuelVehicleRepo{
		findByUserFunc: func(ctx context.Context, userId string, id uint) (*FuelVehicle, error) {
			return &FuelVehicle{Id: int64(id), UserId: userId, Name: "Car"}, nil
		},
	}, &mockRefuelRecordRepo{
		listAllFunc: func(ctx context.Context, userId string, vehicleId uint) ([]*RefuelRecord, error) {
			return []*RefuelRecord{
				{
					Id:         1,
					VehicleId:  1,
					RefuelTime: "2026-01-01 08:00:00",
					Odometer:   decimal.NewFromInt(1000),
					Volume:     decimal.NewFromInt(30),
					Amount:     decimal.NewFromInt(210),
					IsFull:     true,
				},
				{
					Id:         2,
					VehicleId:  1,
					RefuelTime: "2026-01-10 08:00:00",
					Odometer:   decimal.NewFromInt(1300),
					Volume:     decimal.NewFromInt(20),
					Amount:     decimal.NewFromInt(140),
					IsFull:     false,
				},
				{
					Id:         3,
					VehicleId:  1,
					RefuelTime: "2026-01-20 08:00:00",
					Odometer:   decimal.NewFromInt(1600),
					Volume:     decimal.NewFromInt(25),
					Amount:     decimal.NewFromInt(175),
					IsFull:     true,
				},
				{
					Id:         4,
					VehicleId:  1,
					RefuelTime: "2026-01-25 08:00:00",
					Odometer:   decimal.NewFromInt(1500),
					Volume:     decimal.NewFromInt(10),
					Amount:     decimal.NewFromInt(70),
					IsFull:     true,
				},
			}, nil
		},
	}, log.DefaultLogger)

	stats, err := uc.GetFuelStats(withUser(context.Background(), "user-123"), 1)

	assert.NoError(t, err)
	assert.True(t, stats.TotalDistance.Equal(decimal.NewFromInt(600)))
	assert.True(t, stats.TotalVolume.Equal(decimal.NewFromInt(85)))
	assert.True(t, stats.TotalAmount.Equal(decimal.NewFromInt(595)))
	assert.True(t, stats.AverageConsumption.Equal(decimal.RequireFromString("7.50")))
	assert.True(t, stats.LatestConsumption.Equal(decimal.RequireFromString("7.50")))
	assert.True(t, stats.CostPerKm.Equal(decimal.RequireFromString("0.99")))
	assert.Len(t, stats.Trend, 1)
	assert.True(t, stats.Trend[0].Consumption.Equal(decimal.RequireFromString("7.50")))
	assert.Equal(t, "2026-01-20 08:00:00", stats.Trend[0].RefuelTime)
}
