package biz

import (
	"blog/internal/utils"
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

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
	StatsAnchor         bool              // 仅作时间范围统计的区间锚点，不参与统计
	Attachments         []*FuelAttachment // 请求：提交的附件引用；响应：含 url/fileName
}

// FuelAttachment 加油记录附件。FileUrl/FileName 仅在查询响应时填充。
type FuelAttachment struct {
	Id         int64
	FileId     uint
	AttachType string
	Sort       int32
	FileUrl    string
	FileName   string
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

const (
	maxFuelAttachments    = 6
	maxFuelAttachmentSize = 10 * 1024 * 1024 // 10MB
)

// fuelAttachmentTypes 附件类型枚举。
var fuelAttachmentTypes = map[string]bool{
	"receipt":     true,
	"environment": true,
	"other":       true,
}

// fuelImageExts 允许的图片扩展名。
var fuelImageExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
	".heic": true,
	".heif": true,
	".gif":  true,
	".bmp":  true,
}

// isFuelImageFile 通过 MIME 类型或扩展名判断文件是否为图片。
func isFuelImageFile(file *FileRecord) bool {
	if file == nil {
		return false
	}
	if strings.HasPrefix(strings.ToLower(file.FileType), "image/") {
		return true
	}
	return fuelImageExts[strings.ToLower(file.FileExt)]
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
	fileRepo    FileRepo
	log         *log.Helper
}

func NewFuelUsecase(vehicleRepo FuelVehicleRepo, recordRepo RefuelRecordRepo, fileRepo FileRepo, logger log.Logger) *FuelUsecase {
	return &FuelUsecase{
		vehicleRepo: vehicleRepo,
		recordRepo:  recordRepo,
		fileRepo:    fileRepo,
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
	if err := uc.validateFuelAttachments(ctx, userId, record.Attachments); err != nil {
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
	if record.Attachments != nil {
		// 显式提交了附件列表才校验并整组替换；未传则保持现有附件不变
		if err := uc.validateFuelAttachments(ctx, userId, record.Attachments); err != nil {
			return 0, err
		}
	}
	record.Id = dbRecord.Id
	record.UserId = userId
	if err := uc.recordRepo.Update(ctx, record); err != nil {
		return 0, err
	}
	return uint(dbRecord.Id), nil
}

// validateFuelAttachments 校验附件：数量≤6、类型枚举合法、fileId 归属当前用户、必须是图片且≤10MB。
// 校验通过后按提交顺序回填 sort。
func (uc *FuelUsecase) validateFuelAttachments(ctx context.Context, userId string, attachments []*FuelAttachment) error {
	if len(attachments) > maxFuelAttachments {
		return errors.New("附件数量不能超过 6 张")
	}
	for i, att := range attachments {
		if !fuelAttachmentTypes[att.AttachType] {
			return fmt.Errorf("附件类型不合法: %s", att.AttachType)
		}
		file, err := uc.fileRepo.GetByIdAndUserId(ctx, att.FileId, userId)
		if err != nil {
			return fmt.Errorf("附件文件不存在或无权访问: %d", att.FileId)
		}
		if !isFuelImageFile(file) {
			return fmt.Errorf("附件必须是图片: %s", file.FileName)
		}
		if file.FileSize > maxFuelAttachmentSize {
			return fmt.Errorf("附件大小不能超过 10MB: %s", file.FileName)
		}
		att.Sort = int32(i)
	}
	return nil
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

func (uc *FuelUsecase) GetFuelStats(ctx context.Context, vehicleId uint, startTime, endTime string) (*FuelStats, error) {
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
	filtered, anchor := filterFuelRecordsByTime(records, startTime, endTime)
	if anchor != nil {
		// 范围内无记录：锚点已标记 StatsAnchor，仅提供区间起点不参与统计
		filtered = append([]*RefuelRecord{anchor}, filtered...)
	}
	return CalculateFuelStats(int64(vehicleId), filtered), nil
}

// filterFuelRecordsByTime 按时间范围过滤记录，并返回范围起点前最近的一条加满记录作为锚点。
// 锚点（标记 StatsAnchor）及其之后（范围内第一条记录之前）的部分加油记录一并并入，以保证跨边界区间的油量数据完整。
func filterFuelRecordsByTime(records []*RefuelRecord, startTime, endTime string) ([]*RefuelRecord, *RefuelRecord) {
	if startTime == "" && endTime == "" {
		return records, nil
	}
	startTime, endTime = normalizeFuelTimeRange(startTime, endTime)
	if startTime != "" && endTime != "" && startTime > endTime {
		return nil, nil
	}

	var anchor *RefuelRecord
	var pending []*RefuelRecord
	filtered := make([]*RefuelRecord, 0, len(records))
	for _, record := range records {
		if endTime != "" && record.RefuelTime > endTime {
			continue
		}
		if startTime != "" && record.RefuelTime < startTime {
			if record.IsFull {
				// 持续更新为最近一条加满记录，之前暂存的部分加油记录作废
				anchor = record
				pending = pending[:0]
			} else if anchor != nil {
				// 锚点之后的部分加油记录：随锚点一并并入，保证区间油量完整
				pending = append(pending, record)
			}
			continue
		}
		// 遇到第一条 >= startTime 的记录：并入锚点与其后的部分加油记录
		if anchor != nil {
			anchor.StatsAnchor = true
			filtered = append(filtered, anchor)
			filtered = append(filtered, pending...)
			anchor = nil
			pending = nil
		}
		filtered = append(filtered, record)
	}
	if anchor != nil {
		// 范围内无记录：锚点仍返回，由调用方纳入（仅作区间起点，统计结果为空）
		anchor.StatsAnchor = true
	}
	return filtered, anchor
}

// normalizeFuelTimeRange 将日期归一化为 "2006-01-02 15:04:05" 格式以便字符串比较。
func normalizeFuelTimeRange(startTime, endTime string) (string, string) {
	if len(startTime) == 10 {
		startTime += " 00:00:00"
	}
	if len(endTime) == 10 {
		endTime += " 23:59:59"
	}
	return startTime, endTime
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

	var lastFullIndex = -1
	validDistance := decimal.Zero
	validVolume := decimal.Zero
	validAmount := decimal.Zero
	for i, record := range sortedRecords {
		if record.StatsAnchor {
			// 锚点仅提供区间起点，不参与任何统计
			lastFullIndex = i
			continue
		}
		if !record.IsFull {
			continue
		}
		if lastFullIndex >= 0 {
			distance := record.Odometer.Sub(sortedRecords[lastFullIndex].Odometer)
			if distance.GreaterThan(decimal.Zero) {
				intervalVolume := decimal.Zero
				intervalAmount := decimal.Zero
				for j := lastFullIndex + 1; j <= i; j++ {
					intervalVolume = intervalVolume.Add(sortedRecords[j].Volume)
					intervalAmount = intervalAmount.Add(sortedRecords[j].Amount)
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
				validAmount = validAmount.Add(intervalAmount)
				stats.LatestConsumption = consumption
			}
		}
		lastFullIndex = i
	}

	// 统一口径：总里程/总油量/总油费只累计加满锚点区间内实际消耗的部分
	stats.TotalDistance = validDistance
	stats.TotalVolume = validVolume
	stats.TotalAmount = validAmount
	if validDistance.GreaterThan(decimal.Zero) {
		stats.AverageConsumption = validVolume.Div(validDistance).Mul(decimal.NewFromInt(100)).Round(2)
		stats.CostPerKm = validAmount.Div(validDistance).Round(2)
	}
	return stats
}
