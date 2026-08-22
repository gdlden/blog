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
	uc            *biz.FuelUsecase
	ocrRecognizer VisionTextRecognizer
}

func NewFuelService(uc *biz.FuelUsecase) *FuelService {
	return NewFuelServiceWithRecognizer(uc, NewVisionTextRecognizerFromEnv())
}

// NewFuelServiceWithRecognizer 供测试注入 mock 识别器；recognizer 为 nil 时回退到环境变量配置。
func NewFuelServiceWithRecognizer(uc *biz.FuelUsecase, recognizer VisionTextRecognizer) *FuelService {
	if recognizer == nil {
		recognizer = NewVisionTextRecognizerFromEnv()
	}
	return &FuelService{uc: uc, ocrRecognizer: recognizer}
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
	record, err := refuelRecordFromRequest(req, true)
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
	record, err := refuelRecordFromRequest(req, false)
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
	stats, err := s.uc.GetFuelStats(ctx, uint(vehicleId), req.StartTime, req.EndTime)
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
	if req.Name == "" {
		return nil, errors.New("车辆名称不能为空")
	}
	tankCapacity, err := parseDecimalOrZero(req.TankCapacity)
	if err != nil {
		return nil, err
	}
	if tankCapacity.IsNegative() {
		return nil, errors.New("油箱容量不能为负数")
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

// refuelRecordFromRequest 校验并转换加油记录请求。
// strictAmount=true（新建）：amount 非空时必须与油量×单价一致，否则拒绝；
// strictAmount=false（编辑）：amount 自动以油量×单价重算，避免存量手输金额记录编辑报错。
func refuelRecordFromRequest(req *pb.RefuelRecord, strictAmount bool) (*biz.RefuelRecord, error) {
	if req == nil {
		return nil, errors.New("request is nil")
	}
	if req.RefuelTime == "" {
		return nil, errors.New("加油时间不能为空")
	}
	odometer, err := parseDecimalOrZero(req.Odometer)
	if err != nil {
		return nil, err
	}
	if odometer.IsNegative() {
		return nil, errors.New("总里程不能为负数")
	}
	volume, err := parseDecimalOrZero(req.Volume)
	if err != nil {
		return nil, err
	}
	if volume.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("油量必须大于 0")
	}
	unitPrice, err := parseDecimalOrZero(req.UnitPrice)
	if err != nil {
		return nil, err
	}
	if unitPrice.IsNegative() {
		return nil, errors.New("单价不能为负数")
	}
	amount, err := parseDecimalOrZero(req.Amount)
	if err != nil {
		return nil, err
	}
	expectedAmount := volume.Mul(unitPrice).Round(2)
	if strictAmount {
		if req.Amount == "" {
			// 未传金额时按油量×单价自动计算
			amount = expectedAmount
		} else if !amount.Equal(expectedAmount) {
			return nil, errors.New("金额与油量×单价不一致")
		}
	} else {
		amount = expectedAmount
	}
	actualAmount, err := parseDecimalOrZero(req.ActualAmount)
	if err != nil {
		return nil, err
	}
	if req.ActualAmount == "" {
		return nil, errors.New("实付金额不能为空")
	}
	if actualAmount.IsNegative() {
		return nil, errors.New("实付金额不能为负数")
	}
	attachments, err := fuelAttachmentsFromRequest(req.Attachments)
	if err != nil {
		return nil, err
	}
	return &biz.RefuelRecord{
		Id:           req.Id,
		VehicleId:    req.VehicleId,
		RefuelTime:   req.RefuelTime,
		Odometer:     odometer,
		Volume:       volume,
		UnitPrice:    unitPrice,
		Amount:       amount,
		ActualAmount: actualAmount,
		Station:      req.Station,
		IsFull:       req.IsFull,
		Remark:       req.Remark,
		Attachments:  attachments,
	}, nil
}

// fuelAttachmentsFromRequest 转换附件引用列表。
// 保持 nil（未传）与空 slice（传了空数组）的语义差异，供更新时区分"保持不变"与"清空附件"。
func fuelAttachmentsFromRequest(atts []*pb.FuelAttachment) ([]*biz.FuelAttachment, error) {
	if atts == nil {
		return nil, nil
	}
	result := make([]*biz.FuelAttachment, 0, len(atts))
	for _, att := range atts {
		if att == nil || att.FileId == "" {
			return nil, errors.New("附件文件 ID 不能为空")
		}
		fileId, err := strconv.ParseUint(att.FileId, 10, 64)
		if err != nil || fileId == 0 {
			return nil, errors.New("附件文件 ID 不合法")
		}
		result = append(result, &biz.FuelAttachment{
			FileId:     uint(fileId),
			AttachType: att.AttachType,
			Sort:       att.Sort,
		})
	}
	return result, nil
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
	reply := &pb.RefuelRecord{
		Id:                  record.Id,
		VehicleId:           record.VehicleId,
		RefuelTime:          record.RefuelTime,
		Odometer:            record.Odometer.String(),
		Volume:              record.Volume.String(),
		UnitPrice:           record.UnitPrice.String(),
		Amount:              record.Amount.String(),
		ActualAmount:        record.ActualAmount.String(),
		Station:             record.Station,
		IsFull:              record.IsFull,
		Remark:              record.Remark,
		IntervalConsumption: record.IntervalConsumption.String(),
		AttachmentInfos:     make([]*pb.FuelAttachmentInfo, 0, len(record.Attachments)),
	}
	for _, att := range record.Attachments {
		reply.AttachmentInfos = append(reply.AttachmentInfos, &pb.FuelAttachmentInfo{
			Id:         strconv.FormatInt(att.Id, 10),
			FileId:     strconv.FormatUint(uint64(att.FileId), 10),
			Url:        att.FileUrl,
			FileName:   att.FileName,
			AttachType: att.AttachType,
			Sort:       att.Sort,
		})
	}
	return reply
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
