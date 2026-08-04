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

	stats, err := uc.GetFuelStats(withUser(context.Background(), "user-123"), 1, "", "")

	assert.NoError(t, err)
	assert.True(t, stats.TotalDistance.Equal(decimal.NewFromInt(600)))
	// 新口径：总油量/总金额只累计加满区间内实际消耗的部分（记录2+记录3：20+25L、140+175元）
	assert.True(t, stats.TotalVolume.Equal(decimal.NewFromInt(45)))
	assert.True(t, stats.TotalAmount.Equal(decimal.NewFromInt(315)))
	assert.True(t, stats.AverageConsumption.Equal(decimal.RequireFromString("7.50")))
	assert.True(t, stats.LatestConsumption.Equal(decimal.RequireFromString("7.50")))
	assert.True(t, stats.CostPerKm.Equal(decimal.RequireFromString("0.53")))
	assert.Len(t, stats.Trend, 1)
	assert.True(t, stats.Trend[0].Consumption.Equal(decimal.RequireFromString("7.50")))
	assert.Equal(t, "2026-01-20 08:00:00", stats.Trend[0].RefuelTime)
}

func TestFuelUsecase_GetStatsFiltersByTimeRangeWithAnchor(t *testing.T) {
	records := []*RefuelRecord{
		{Id: 1, RefuelTime: "2026-01-01 08:00:00", Odometer: decimal.NewFromInt(1000), Volume: decimal.NewFromInt(30), Amount: decimal.NewFromInt(210), IsFull: true},
		{Id: 2, RefuelTime: "2026-01-10 08:00:00", Odometer: decimal.NewFromInt(1300), Volume: decimal.NewFromInt(20), Amount: decimal.NewFromInt(140), IsFull: false},
		{Id: 3, RefuelTime: "2026-01-20 08:00:00", Odometer: decimal.NewFromInt(1600), Volume: decimal.NewFromInt(25), Amount: decimal.NewFromInt(175), IsFull: true},
		{Id: 4, RefuelTime: "2026-02-01 08:00:00", Odometer: decimal.NewFromInt(2100), Volume: decimal.NewFromInt(40), Amount: decimal.NewFromInt(280), IsFull: true},
	}
	uc := NewFuelUsecase(&mockFuelVehicleRepo{
		findByUserFunc: func(ctx context.Context, userId string, id uint) (*FuelVehicle, error) {
			return &FuelVehicle{Id: int64(id), UserId: userId, Name: "Car"}, nil
		},
	}, &mockRefuelRecordRepo{
		listAllFunc: func(ctx context.Context, userId string, vehicleId uint) ([]*RefuelRecord, error) {
			return records, nil
		},
	}, log.DefaultLogger)

	// 范围从 2026-01-15 开始：锚点为记录1（01-01 加满），记录2（01-10）保留以保证区间油量完整
	stats, err := uc.GetFuelStats(withUser(context.Background(), "user-123"), 1, "2026-01-15", "2026-01-31")

	assert.NoError(t, err)
	assert.Len(t, stats.Trend, 1)
	assert.Equal(t, "2026-01-20 08:00:00", stats.Trend[0].RefuelTime)
	// 区间 [记录1, 记录3]：distance=600，油量=记录2+记录3=45L，金额=140+175=315
	assert.True(t, stats.Trend[0].Distance.Equal(decimal.NewFromInt(600)))
	assert.True(t, stats.Trend[0].Volume.Equal(decimal.NewFromInt(45)))
	assert.True(t, stats.Trend[0].Consumption.Equal(decimal.RequireFromString("7.50")))
	assert.True(t, stats.TotalDistance.Equal(decimal.NewFromInt(600)))
	assert.True(t, stats.TotalVolume.Equal(decimal.NewFromInt(45)))
	assert.True(t, stats.TotalAmount.Equal(decimal.NewFromInt(315)))
	// 范围外的记录4 不参与统计
	assert.Equal(t, 1, len(stats.Trend))

	// 范围起点之前没有加满锚点时，范围内首个加满记录无前锚点，不产生区间
	stats2, err := uc.GetFuelStats(withUser(context.Background(), "user-123"), 1, "2026-01-05", "2026-01-10")
	assert.NoError(t, err)
	assert.Len(t, stats2.Trend, 0)
	assert.True(t, stats2.TotalDistance.Equal(decimal.Zero))
}

func TestFuelUsecase_GetStatsUsesNearestAnchorBeforeRange(t *testing.T) {
	records := []*RefuelRecord{
		{Id: 1, RefuelTime: "2026-01-01 08:00:00", Odometer: decimal.NewFromInt(1000), Volume: decimal.NewFromInt(30), Amount: decimal.NewFromInt(210), IsFull: true},
		{Id: 2, RefuelTime: "2026-01-05 08:00:00", Odometer: decimal.NewFromInt(1300), Volume: decimal.NewFromInt(20), Amount: decimal.NewFromInt(140), IsFull: true},
		{Id: 3, RefuelTime: "2026-01-08 08:00:00", Odometer: decimal.NewFromInt(1450), Volume: decimal.NewFromInt(10), Amount: decimal.NewFromInt(70), IsFull: false},
		{Id: 4, RefuelTime: "2026-01-20 08:00:00", Odometer: decimal.NewFromInt(1750), Volume: decimal.NewFromInt(25), Amount: decimal.NewFromInt(175), IsFull: true},
	}
	uc := NewFuelUsecase(&mockFuelVehicleRepo{
		findByUserFunc: func(ctx context.Context, userId string, id uint) (*FuelVehicle, error) {
			return &FuelVehicle{Id: int64(id), UserId: userId, Name: "Car"}, nil
		},
	}, &mockRefuelRecordRepo{
		listAllFunc: func(ctx context.Context, userId string, vehicleId uint) ([]*RefuelRecord, error) {
			return records, nil
		},
	}, log.DefaultLogger)

	// 起点前有两条加满记录：锚点必须取最近的 01-05（记录2），更早的 01-01（记录1）完整区间不得计入
	stats, err := uc.GetFuelStats(withUser(context.Background(), "user-123"), 1, "2026-01-10", "")

	assert.NoError(t, err)
	assert.Len(t, stats.Trend, 1)
	assert.Equal(t, "2026-01-20 08:00:00", stats.Trend[0].RefuelTime)
	// 区间 [记录2(1300km), 记录4(1750km)]：distance=450，油量=记录3+记录4=35L，金额=70+175=245
	assert.True(t, stats.TotalDistance.Equal(decimal.NewFromInt(450)))
	assert.True(t, stats.TotalVolume.Equal(decimal.NewFromInt(35)))
	assert.True(t, stats.TotalAmount.Equal(decimal.NewFromInt(245)))
	assert.True(t, stats.Trend[0].Consumption.Equal(decimal.RequireFromString("7.78")))
	assert.True(t, stats.AverageConsumption.Equal(decimal.RequireFromString("7.78")))
}

func TestFuelUsecase_GetStatsRangeWithNoRecordsInside(t *testing.T) {
	records := []*RefuelRecord{
		{Id: 1, RefuelTime: "2026-01-01 08:00:00", Odometer: decimal.NewFromInt(1000), Volume: decimal.NewFromInt(30), Amount: decimal.NewFromInt(210), IsFull: true},
		{Id: 2, RefuelTime: "2026-01-20 08:00:00", Odometer: decimal.NewFromInt(1600), Volume: decimal.NewFromInt(25), Amount: decimal.NewFromInt(175), IsFull: true},
	}
	uc := NewFuelUsecase(&mockFuelVehicleRepo{
		findByUserFunc: func(ctx context.Context, userId string, id uint) (*FuelVehicle, error) {
			return &FuelVehicle{Id: int64(id), UserId: userId, Name: "Car"}, nil
		},
	}, &mockRefuelRecordRepo{
		listAllFunc: func(ctx context.Context, userId string, vehicleId uint) ([]*RefuelRecord, error) {
			return records, nil
		},
	}, log.DefaultLogger)

	// 范围内无记录：锚点纳入后无完整区间，统计全零
	stats, err := uc.GetFuelStats(withUser(context.Background(), "user-123"), 1, "2026-02-01", "2026-02-28")

	assert.NoError(t, err)
	assert.Len(t, stats.Trend, 0)
	assert.True(t, stats.TotalDistance.Equal(decimal.Zero))
	assert.True(t, stats.TotalVolume.Equal(decimal.Zero))
	assert.True(t, stats.TotalAmount.Equal(decimal.Zero))
}

func TestFilterFuelRecordsByTime_RejectsReversedRange(t *testing.T) {
	records := []*RefuelRecord{
		{Id: 1, RefuelTime: "2026-01-01 08:00:00", IsFull: true},
		{Id: 2, RefuelTime: "2026-01-20 08:00:00", IsFull: true},
	}

	filtered, anchor := filterFuelRecordsByTime(records, "2026-02-01", "2026-01-01")

	assert.Nil(t, filtered)
	assert.Nil(t, anchor)
}
