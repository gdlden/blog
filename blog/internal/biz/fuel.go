package biz

import (
	"blog/internal/utils"
	"context"
	"errors"
	"sort"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/shopspring/decimal"
)

type FuelVehicle struct {
	Id           int64
	Name         string
	PlateNo      string
	Brand        string
	Model        string
	TankCapacity decimal.Decimal
	Remark       string
	UserId       string
}

type RefuelRecord struct {
	Id                  int64
	VehicleId           int64
	RefuelTime          string
	Odometer            decimal.Decimal
	Volume              decimal.Decimal
	UnitPrice           decimal.Decimal
	Amount              decimal.Decimal
	Station             string
	IsFull              bool
	Remark              string
	IntervalConsumption decimal.Decimal
	UserId              string
}

type FuelTrendPoint struct {
	RefuelTime   string
	Odometer     decimal.Decimal
	Consumption  decimal.Decimal
	Distance     decimal.Decimal
	Volume       decimal.Decimal
	RefuelRecord int64
}

type FuelStats struct {
	VehicleId          int64
	TotalDistance      decimal.Decimal
	TotalVolume        decimal.Decimal
	TotalAmount        decimal.Decimal
	AverageConsumption decimal.Decimal
	LatestConsumption  decimal.Decimal
	CostPerKm          decimal.Decimal
	Trend              []*FuelTrendPoint
}

type FuelVehicleListQuery struct {
	Page     int64
	PageSize int64
	Name     string
	PlateNo  string
}

type RefuelRecordListQuery struct {
	Page     int64
	PageSize int64
}

type FuelVehicleRepo interface {
	Save(context.Context, *FuelVehicle) (uint, error)
	Update(context.Context, *FuelVehicle) error
	DeleteByUserIdAndVehicleId(context.Context, string, uint) error
	FindByUserIdAndVehicleId(context.Context, string, uint) (*FuelVehicle, error)
	ListByUserId(context.Context, string, *FuelVehicleListQuery) ([]*FuelVehicle, int64, error)
	CountRefuelRecordByVehicleId(context.Context, uint) (int64, error)
}

type RefuelRecordRepo interface {
	Save(context.Context, *RefuelRecord) (uint, error)
	Update(context.Context, *RefuelRecord) error
	DeleteByUserIdAndRecordId(context.Context, string, uint) error
	FindByUserIdAndRecordId(context.Context, string, uint) (*RefuelRecord, error)
	ListByUserIdAndVehicleId(context.Context, string, uint, *RefuelRecordListQuery) ([]*RefuelRecord, int64, error)
	ListAllByUserIdAndVehicleId(context.Context, string, uint) ([]*RefuelRecord, error)
}

type FuelUsecase struct {
	vehicleRepo FuelVehicleRepo
	recordRepo  RefuelRecordRepo
	log         *log.Helper
}

func NewFuelUsecase(vehicleRepo FuelVehicleRepo, recordRepo RefuelRecordRepo, logger log.Logger) *FuelUsecase {
	return &FuelUsecase{
		vehicleRepo: vehicleRepo,
		recordRepo:  recordRepo,
		log:         log.NewHelper(logger),
	}
}

func (uc *FuelUsecase) CreateVehicle(ctx context.Context, vehicle *FuelVehicle) (uint, error) {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return 0, err
	}
	vehicle.UserId = userId
	return uc.vehicleRepo.Save(ctx, vehicle)
}

func (uc *FuelUsecase) UpdateVehicle(ctx context.Context, vehicle *FuelVehicle) (uint, error) {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return 0, err
	}
	dbVehicle, err := uc.vehicleRepo.FindByUserIdAndVehicleId(ctx, userId, uint(vehicle.Id))
	if err != nil {
		return 0, err
	}
	vehicle.Id = dbVehicle.Id
	vehicle.UserId = userId
	if err := uc.vehicleRepo.Update(ctx, vehicle); err != nil {
		return 0, err
	}
	return uint(dbVehicle.Id), nil
}

func (uc *FuelUsecase) DeleteVehicle(ctx context.Context, vehicleId uint) error {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return err
	}
	if _, err := uc.vehicleRepo.FindByUserIdAndVehicleId(ctx, userId, vehicleId); err != nil {
		return err
	}
	count, err := uc.vehicleRepo.CountRefuelRecordByVehicleId(ctx, vehicleId)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("存在加油记录，禁止删除")
	}
	return uc.vehicleRepo.DeleteByUserIdAndVehicleId(ctx, userId, vehicleId)
}

func (uc *FuelUsecase) GetVehicle(ctx context.Context, vehicleId uint) (*FuelVehicle, error) {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return nil, err
	}
	return uc.vehicleRepo.FindByUserIdAndVehicleId(ctx, userId, vehicleId)
}

func (uc *FuelUsecase) ListVehicles(ctx context.Context, query *FuelVehicleListQuery) ([]*FuelVehicle, int64, error) {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return nil, 0, err
	}
	return uc.vehicleRepo.ListByUserId(ctx, userId, query)
}

func (uc *FuelUsecase) CreateRefuelRecord(ctx context.Context, record *RefuelRecord) (uint, error) {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return 0, err
	}
	if _, err := uc.vehicleRepo.FindByUserIdAndVehicleId(ctx, userId, uint(record.VehicleId)); err != nil {
		return 0, err
	}
	record.UserId = userId
	return uc.recordRepo.Save(ctx, record)
}

func (uc *FuelUsecase) UpdateRefuelRecord(ctx context.Context, record *RefuelRecord) (uint, error) {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return 0, err
	}
	dbRecord, err := uc.recordRepo.FindByUserIdAndRecordId(ctx, userId, uint(record.Id))
	if err != nil {
		return 0, err
	}
	if _, err := uc.vehicleRepo.FindByUserIdAndVehicleId(ctx, userId, uint(record.VehicleId)); err != nil {
		return 0, err
	}
	record.Id = dbRecord.Id
	record.UserId = userId
	if err := uc.recordRepo.Update(ctx, record); err != nil {
		return 0, err
	}
	return uint(dbRecord.Id), nil
}

func (uc *FuelUsecase) DeleteRefuelRecord(ctx context.Context, recordId uint) error {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return err
	}
	if _, err := uc.recordRepo.FindByUserIdAndRecordId(ctx, userId, recordId); err != nil {
		return err
	}
	return uc.recordRepo.DeleteByUserIdAndRecordId(ctx, userId, recordId)
}

func (uc *FuelUsecase) GetRefuelRecord(ctx context.Context, recordId uint) (*RefuelRecord, error) {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return nil, err
	}
	return uc.recordRepo.FindByUserIdAndRecordId(ctx, userId, recordId)
}

func (uc *FuelUsecase) ListRefuelRecords(ctx context.Context, vehicleId uint, query *RefuelRecordListQuery) ([]*RefuelRecord, int64, error) {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return nil, 0, err
	}
	if _, err := uc.vehicleRepo.FindByUserIdAndVehicleId(ctx, userId, vehicleId); err != nil {
		return nil, 0, err
	}
	return uc.recordRepo.ListByUserIdAndVehicleId(ctx, userId, vehicleId, query)
}

func (uc *FuelUsecase) GetFuelStats(ctx context.Context, vehicleId uint) (*FuelStats, error) {
	userId, err := utils.CurrentUserId(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := uc.vehicleRepo.FindByUserIdAndVehicleId(ctx, userId, vehicleId); err != nil {
		return nil, err
	}
	records, err := uc.recordRepo.ListAllByUserIdAndVehicleId(ctx, userId, vehicleId)
	if err != nil {
		return nil, err
	}
	return CalculateFuelStats(int64(vehicleId), records), nil
}

func CalculateFuelStats(vehicleId int64, records []*RefuelRecord) *FuelStats {
	stats := &FuelStats{
		VehicleId: vehicleId,
		Trend:     []*FuelTrendPoint{},
	}
	if len(records) == 0 {
		return stats
	}

	sortedRecords := append([]*RefuelRecord(nil), records...)
	sort.SliceStable(sortedRecords, func(i, j int) bool {
		if sortedRecords[i].RefuelTime == sortedRecords[j].RefuelTime {
			return sortedRecords[i].Id < sortedRecords[j].Id
		}
		return sortedRecords[i].RefuelTime < sortedRecords[j].RefuelTime
	})

	minOdometer := sortedRecords[0].Odometer
	maxOdometer := sortedRecords[0].Odometer
	for _, record := range sortedRecords {
		stats.TotalVolume = stats.TotalVolume.Add(record.Volume)
		stats.TotalAmount = stats.TotalAmount.Add(record.Amount)
		if record.Odometer.LessThan(minOdometer) {
			minOdometer = record.Odometer
		}
		if record.Odometer.GreaterThan(maxOdometer) {
			maxOdometer = record.Odometer
		}
	}
	if maxOdometer.GreaterThan(minOdometer) {
		stats.TotalDistance = maxOdometer.Sub(minOdometer)
		stats.CostPerKm = stats.TotalAmount.Div(stats.TotalDistance).Round(2)
	}

	var lastFullIndex = -1
	validDistance := decimal.Zero
	validVolume := decimal.Zero
	for i, record := range sortedRecords {
		if !record.IsFull {
			continue
		}
		if lastFullIndex >= 0 {
			distance := record.Odometer.Sub(sortedRecords[lastFullIndex].Odometer)
			if distance.GreaterThan(decimal.Zero) {
				intervalVolume := decimal.Zero
				for j := lastFullIndex + 1; j <= i; j++ {
					intervalVolume = intervalVolume.Add(sortedRecords[j].Volume)
				}
				consumption := intervalVolume.Div(distance).Mul(decimal.NewFromInt(100)).Round(2)
				stats.Trend = append(stats.Trend, &FuelTrendPoint{
					RefuelTime:   record.RefuelTime,
					Odometer:     record.Odometer,
					Consumption:  consumption,
					Distance:     distance,
					Volume:       intervalVolume,
					RefuelRecord: record.Id,
				})
				record.IntervalConsumption = consumption
				validDistance = validDistance.Add(distance)
				validVolume = validVolume.Add(intervalVolume)
				stats.LatestConsumption = consumption
			}
		}
		lastFullIndex = i
	}

	if validDistance.GreaterThan(decimal.Zero) {
		stats.AverageConsumption = validVolume.Div(validDistance).Mul(decimal.NewFromInt(100)).Round(2)
	}
	return stats
}
