package service

import (
	"blog/internal/biz"
	"context"
	"errors"
	"strconv"

	pb "blog/api/fuel/v1"

	"github.com/shopspring/decimal"
)

type FuelService struct {
	pb.UnimplementedFuelServer
	uc *biz.FuelUsecase
}

func NewFuelService(uc *biz.FuelUsecase) *FuelService {
	return &FuelService{uc: uc}
}

func (s *FuelService) CreateVehicle(ctx context.Context, req *pb.FuelVehicle) (*pb.SaveFuelReply, error) {
	vehicle, err := fuelVehicleFromRequest(req)
	if err != nil {
		return nil, err
	}
	id, err := s.uc.CreateVehicle(ctx, vehicle)
	if err != nil {
		return nil, err
	}
	return &pb.SaveFuelReply{Id: strconv.FormatUint(uint64(id), 10), Message: "save success"}, nil
}

func (s *FuelService) UpdateVehicle(ctx context.Context, req *pb.FuelVehicle) (*pb.SaveFuelReply, error) {
	vehicle, err := fuelVehicleFromRequest(req)
	if err != nil {
		return nil, err
	}
	if vehicle.Id <= 0 {
		return nil, errors.New("invalid vehicle id")
	}
	id, err := s.uc.UpdateVehicle(ctx, vehicle)
	if err != nil {
		return nil, err
	}
	return &pb.SaveFuelReply{Id: strconv.FormatUint(uint64(id), 10), Message: "update success"}, nil
}

func (s *FuelService) DeleteVehicle(ctx context.Context, req *pb.DeleteFuelVehicleRequest) (*pb.DeleteFuelReply, error) {
	id, err := parseUintID(req.GetId(), "invalid vehicle id")
	if err != nil {
		return &pb.DeleteFuelReply{Flag: false}, err
	}
	if err := s.uc.DeleteVehicle(ctx, uint(id)); err != nil {
		return &pb.DeleteFuelReply{Flag: false}, err
	}
	return &pb.DeleteFuelReply{Flag: true}, nil
}

func (s *FuelService) GetVehicle(ctx context.Context, req *pb.GetFuelVehicleRequest) (*pb.FuelVehicle, error) {
	id, err := parseUintID(req.GetId(), "invalid vehicle id")
	if err != nil {
		return nil, err
	}
	vehicle, err := s.uc.GetVehicle(ctx, uint(id))
	if err != nil {
		return nil, err
	}
	return fuelVehicleToReply(vehicle), nil
}

func (s *FuelService) ListVehicles(ctx context.Context, req *pb.ListFuelVehiclesRequest) (*pb.ListFuelVehiclesReply, error) {
	if req == nil {
		req = &pb.ListFuelVehiclesRequest{}
	}
	page, pageSize := normalizeServicePage(req.Page, req.PageSize)
	items, total, err := s.uc.ListVehicles(ctx, &biz.FuelVehicleListQuery{
		Page:     page,
		PageSize: pageSize,
		Name:     req.Name,
		PlateNo:  req.PlateNo,
	})
	if err != nil {
		return nil, err
	}
	list := make([]*pb.FuelVehicle, 0, len(items))
	for _, vehicle := range items {
		list = append(list, fuelVehicleToReply(vehicle))
	}
	return &pb.ListFuelVehiclesReply{Page: page, Total: total, List: list}, nil
}

func (s *FuelService) CreateRefuelRecord(ctx context.Context, req *pb.RefuelRecord) (*pb.SaveFuelReply, error) {
	record, err := refuelRecordFromRequest(req)
	if err != nil {
		return nil, err
	}
	id, err := s.uc.CreateRefuelRecord(ctx, record)
	if err != nil {
		return nil, err
	}
	return &pb.SaveFuelReply{Id: strconv.FormatUint(uint64(id), 10), Message: "save success"}, nil
}

func (s *FuelService) UpdateRefuelRecord(ctx context.Context, req *pb.RefuelRecord) (*pb.SaveFuelReply, error) {
	record, err := refuelRecordFromRequest(req)
	if err != nil {
		return nil, err
	}
	if record.Id <= 0 {
		return nil, errors.New("invalid refuel record id")
	}
	id, err := s.uc.UpdateRefuelRecord(ctx, record)
	if err != nil {
		return nil, err
	}
	return &pb.SaveFuelReply{Id: strconv.FormatUint(uint64(id), 10), Message: "update success"}, nil
}

func (s *FuelService) DeleteRefuelRecord(ctx context.Context, req *pb.DeleteRefuelRecordRequest) (*pb.DeleteFuelReply, error) {
	id, err := parseUintID(req.GetId(), "invalid refuel record id")
	if err != nil {
		return &pb.DeleteFuelReply{Flag: false}, err
	}
	if err := s.uc.DeleteRefuelRecord(ctx, uint(id)); err != nil {
		return &pb.DeleteFuelReply{Flag: false}, err
	}
	return &pb.DeleteFuelReply{Flag: true}, nil
}

func (s *FuelService) GetRefuelRecord(ctx context.Context, req *pb.GetRefuelRecordRequest) (*pb.RefuelRecord, error) {
	id, err := parseUintID(req.GetId(), "invalid refuel record id")
	if err != nil {
		return nil, err
	}
	record, err := s.uc.GetRefuelRecord(ctx, uint(id))
	if err != nil {
		return nil, err
	}
	return refuelRecordToReply(record), nil
}

func (s *FuelService) ListRefuelRecords(ctx context.Context, req *pb.ListRefuelRecordsRequest) (*pb.ListRefuelRecordsReply, error) {
	if req == nil {
		req = &pb.ListRefuelRecordsRequest{}
	}
	vehicleId, err := parseUintID(req.VehicleId, "invalid vehicle id")
	if err != nil {
		return nil, err
	}
	page, pageSize := normalizeServicePage(req.Page, req.PageSize)
	items, total, err := s.uc.ListRefuelRecords(ctx, uint(vehicleId), &biz.RefuelRecordListQuery{Page: page, PageSize: pageSize})
	if err != nil {
		return nil, err
	}
	list := make([]*pb.RefuelRecord, 0, len(items))
	for _, record := range items {
		list = append(list, refuelRecordToReply(record))
	}
	return &pb.ListRefuelRecordsReply{Page: page, Total: total, List: list}, nil
}

func (s *FuelService) GetFuelStats(ctx context.Context, req *pb.GetFuelStatsRequest) (*pb.FuelStats, error) {
	if req == nil {
		return nil, errors.New("invalid vehicle id")
	}
	vehicleId, err := parseUintID(req.VehicleId, "invalid vehicle id")
	if err != nil {
		return nil, err
	}
	stats, err := s.uc.GetFuelStats(ctx, uint(vehicleId))
	if err != nil {
		return nil, err
	}
	return fuelStatsToReply(stats), nil
}

func parseUintID(value string, message string) (uint64, error) {
	if value == "" {
		return 0, errors.New(message)
	}
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 {
		return 0, errors.New(message)
	}
	return id, nil
}

func normalizeServicePage(page int64, pageSize int64) (int64, int64) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func parseDecimalOrZero(value string) (decimal.Decimal, error) {
	if value == "" {
		return decimal.Zero, nil
	}
	return decimal.NewFromString(value)
}

func fuelVehicleFromRequest(req *pb.FuelVehicle) (*biz.FuelVehicle, error) {
	if req == nil {
		return nil, errors.New("request is nil")
	}
	tankCapacity, err := parseDecimalOrZero(req.TankCapacity)
	if err != nil {
		return nil, err
	}
	return &biz.FuelVehicle{
		Id:           req.Id,
		Name:         req.Name,
		PlateNo:      req.PlateNo,
		Brand:        req.Brand,
		Model:        req.Model,
		TankCapacity: tankCapacity,
		Remark:       req.Remark,
	}, nil
}

func refuelRecordFromRequest(req *pb.RefuelRecord) (*biz.RefuelRecord, error) {
	if req == nil {
		return nil, errors.New("request is nil")
	}
	odometer, err := parseDecimalOrZero(req.Odometer)
	if err != nil {
		return nil, err
	}
	volume, err := parseDecimalOrZero(req.Volume)
	if err != nil {
		return nil, err
	}
	unitPrice, err := parseDecimalOrZero(req.UnitPrice)
	if err != nil {
		return nil, err
	}
	amount, err := parseDecimalOrZero(req.Amount)
	if err != nil {
		return nil, err
	}
	return &biz.RefuelRecord{
		Id:         req.Id,
		VehicleId:  req.VehicleId,
		RefuelTime: req.RefuelTime,
		Odometer:   odometer,
		Volume:     volume,
		UnitPrice:  unitPrice,
		Amount:     amount,
		Station:    req.Station,
		IsFull:     req.IsFull,
		Remark:     req.Remark,
	}, nil
}

func fuelVehicleToReply(vehicle *biz.FuelVehicle) *pb.FuelVehicle {
	if vehicle == nil {
		return &pb.FuelVehicle{}
	}
	return &pb.FuelVehicle{
		Id:           vehicle.Id,
		Name:         vehicle.Name,
		PlateNo:      vehicle.PlateNo,
		Brand:        vehicle.Brand,
		Model:        vehicle.Model,
		TankCapacity: vehicle.TankCapacity.String(),
		Remark:       vehicle.Remark,
	}
}

func refuelRecordToReply(record *biz.RefuelRecord) *pb.RefuelRecord {
	if record == nil {
		return &pb.RefuelRecord{}
	}
	return &pb.RefuelRecord{
		Id:                  record.Id,
		VehicleId:           record.VehicleId,
		RefuelTime:          record.RefuelTime,
		Odometer:            record.Odometer.String(),
		Volume:              record.Volume.String(),
		UnitPrice:           record.UnitPrice.String(),
		Amount:              record.Amount.String(),
		Station:             record.Station,
		IsFull:              record.IsFull,
		Remark:              record.Remark,
		IntervalConsumption: record.IntervalConsumption.String(),
	}
}

func fuelStatsToReply(stats *biz.FuelStats) *pb.FuelStats {
	if stats == nil {
		return &pb.FuelStats{}
	}
	trend := make([]*pb.FuelTrendPoint, 0, len(stats.Trend))
	for _, point := range stats.Trend {
		trend = append(trend, &pb.FuelTrendPoint{
			RefuelTime:     point.RefuelTime,
			Odometer:       point.Odometer.String(),
			Consumption:    point.Consumption.String(),
			Distance:       point.Distance.String(),
			Volume:         point.Volume.String(),
			RefuelRecordId: strconv.FormatInt(point.RefuelRecord, 10),
		})
	}
	return &pb.FuelStats{
		VehicleId:          stats.VehicleId,
		TotalDistance:      stats.TotalDistance.String(),
		TotalVolume:        stats.TotalVolume.String(),
		TotalAmount:        stats.TotalAmount.String(),
		AverageConsumption: stats.AverageConsumption.String(),
		LatestConsumption:  stats.LatestConsumption.String(),
		CostPerKm:          stats.CostPerKm.String(),
		Trend:              trend,
	}
}
