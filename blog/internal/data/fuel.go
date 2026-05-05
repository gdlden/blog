package data

import (
	"blog/internal/biz"
	"context"
	"errors"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type FuelVehicle struct {
	gorm.Model
	Name         string          `gorm:"comment:vehicle name"`
	PlateNo      string          `gorm:"comment:plate number"`
	Brand        string          `gorm:"comment:brand"`
	VehicleModel string          `gorm:"column:model;comment:model"`
	TankCapacity decimal.Decimal `gorm:"type:decimal(10,2);comment:tank capacity"`
	Remark       string          `gorm:"comment:remark"`
	UserId       string          `gorm:"comment:user id"`
}

type RefuelRecord struct {
	gorm.Model
	VehicleId  uint            `gorm:"comment:vehicle id;index"`
	RefuelTime time.Time       `gorm:"comment:refuel time;index"`
	Odometer   decimal.Decimal `gorm:"type:decimal(12,2);comment:odometer"`
	Volume     decimal.Decimal `gorm:"type:decimal(10,2);comment:volume"`
	UnitPrice  decimal.Decimal `gorm:"type:decimal(10,2);comment:unit price"`
	Amount     decimal.Decimal `gorm:"type:decimal(10,2);comment:amount"`
	Station    string          `gorm:"comment:station"`
	IsFull     bool            `gorm:"comment:is full tank"`
	Remark     string          `gorm:"comment:remark"`
	UserId     string          `gorm:"comment:user id;index"`
}

type FuelVehicleRepo struct {
	data *Data
	log  *log.Helper
}

type RefuelRecordRepo struct {
	data *Data
	log  *log.Helper
}

func NewFuelVehicleRepo(data *Data, logger log.Logger) biz.FuelVehicleRepo {
	return &FuelVehicleRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func NewRefuelRecordRepo(data *Data, logger log.Logger) biz.RefuelRecordRepo {
	return &RefuelRecordRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *FuelVehicleRepo) Save(ctx context.Context, vehicle *biz.FuelVehicle) (uint, error) {
	dbVehicle := FuelVehicle{
		Name:         vehicle.Name,
		PlateNo:      vehicle.PlateNo,
		Brand:        vehicle.Brand,
		VehicleModel: vehicle.Model,
		TankCapacity: vehicle.TankCapacity,
		Remark:       vehicle.Remark,
		UserId:       vehicle.UserId,
	}
	tx := r.data.db.WithContext(ctx).Create(&dbVehicle)
	if tx.Error != nil {
		return 0, tx.Error
	}
	return dbVehicle.ID, nil
}

func (r *FuelVehicleRepo) Update(ctx context.Context, vehicle *biz.FuelVehicle) error {
	updates := map[string]any{
		"name":          vehicle.Name,
		"plate_no":      vehicle.PlateNo,
		"brand":         vehicle.Brand,
		"model":         vehicle.Model,
		"tank_capacity": vehicle.TankCapacity,
		"remark":        vehicle.Remark,
	}
	tx := r.data.db.WithContext(ctx).
		Model(&FuelVehicle{}).
		Where("id = ? AND user_id = ?", vehicle.Id, vehicle.UserId).
		Updates(updates)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("no updatable fuel vehicle found for current user")
	}
	return nil
}

func (r *FuelVehicleRepo) DeleteByUserIdAndVehicleId(ctx context.Context, userId string, id uint) error {
	tx := r.data.db.WithContext(ctx).Where("user_id = ? AND id = ?", userId, id).Delete(&FuelVehicle{})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("no deletable fuel vehicle found or permission denied")
	}
	return nil
}

func (r *FuelVehicleRepo) FindByUserIdAndVehicleId(ctx context.Context, userId string, id uint) (*biz.FuelVehicle, error) {
	var vehicle FuelVehicle
	tx := r.data.db.WithContext(ctx).Where("user_id = ? AND id = ?", userId, id).First(&vehicle)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return mapFuelVehicleToBiz(&vehicle), nil
}

func (r *FuelVehicleRepo) ListByUserId(ctx context.Context, userId string, query *biz.FuelVehicleListQuery) ([]*biz.FuelVehicle, int64, error) {
	db := r.data.db.WithContext(ctx).Model(&FuelVehicle{}).Where("user_id = ?", userId)
	if query != nil {
		if query.Name != "" {
			db = db.Where("name LIKE ?", "%"+query.Name+"%")
		}
		if query.PlateNo != "" {
			db = db.Where("plate_no LIKE ?", "%"+query.PlateNo+"%")
		}
	}

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page, pageSize := normalizeFuelPage(query)
	var vehicles []FuelVehicle
	err := db.Order("created_at DESC").
		Offset(int((page - 1) * pageSize)).
		Limit(int(pageSize)).
		Find(&vehicles).Error
	if err != nil {
		return nil, 0, err
	}

	items := make([]*biz.FuelVehicle, 0, len(vehicles))
	for i := range vehicles {
		items = append(items, mapFuelVehicleToBiz(&vehicles[i]))
	}
	return items, total, nil
}

func (r *FuelVehicleRepo) CountRefuelRecordByVehicleId(ctx context.Context, vehicleId uint) (int64, error) {
	var count int64
	tx := r.data.db.WithContext(ctx).Model(&RefuelRecord{}).Where("vehicle_id = ?", vehicleId).Count(&count)
	if tx.Error != nil {
		return 0, tx.Error
	}
	return count, nil
}

func (r *RefuelRecordRepo) Save(ctx context.Context, record *biz.RefuelRecord) (uint, error) {
	refuelTime, err := parseFuelTime(record.RefuelTime)
	if err != nil {
		return 0, err
	}
	dbRecord := RefuelRecord{
		VehicleId:  uint(record.VehicleId),
		RefuelTime: refuelTime,
		Odometer:   record.Odometer,
		Volume:     record.Volume,
		UnitPrice:  record.UnitPrice,
		Amount:     record.Amount,
		Station:    record.Station,
		IsFull:     record.IsFull,
		Remark:     record.Remark,
		UserId:     record.UserId,
	}
	tx := r.data.db.WithContext(ctx).Create(&dbRecord)
	if tx.Error != nil {
		return 0, tx.Error
	}
	return dbRecord.ID, nil
}

func (r *RefuelRecordRepo) Update(ctx context.Context, record *biz.RefuelRecord) error {
	refuelTime, err := parseFuelTime(record.RefuelTime)
	if err != nil {
		return err
	}
	updates := map[string]any{
		"vehicle_id":  record.VehicleId,
		"refuel_time": refuelTime,
		"odometer":    record.Odometer,
		"volume":      record.Volume,
		"unit_price":  record.UnitPrice,
		"amount":      record.Amount,
		"station":     record.Station,
		"is_full":     record.IsFull,
		"remark":      record.Remark,
	}
	tx := r.data.db.WithContext(ctx).
		Model(&RefuelRecord{}).
		Where("id = ? AND user_id = ?", record.Id, record.UserId).
		Updates(updates)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("no updatable refuel record found for current user")
	}
	return nil
}

func (r *RefuelRecordRepo) DeleteByUserIdAndRecordId(ctx context.Context, userId string, id uint) error {
	tx := r.data.db.WithContext(ctx).Where("user_id = ? AND id = ?", userId, id).Delete(&RefuelRecord{})
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("no deletable refuel record found or permission denied")
	}
	return nil
}

func (r *RefuelRecordRepo) FindByUserIdAndRecordId(ctx context.Context, userId string, id uint) (*biz.RefuelRecord, error) {
	var record RefuelRecord
	tx := r.data.db.WithContext(ctx).Where("user_id = ? AND id = ?", userId, id).First(&record)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return mapRefuelRecordToBiz(&record), nil
}

func (r *RefuelRecordRepo) ListByUserIdAndVehicleId(ctx context.Context, userId string, vehicleId uint, query *biz.RefuelRecordListQuery) ([]*biz.RefuelRecord, int64, error) {
	db := r.data.db.WithContext(ctx).Model(&RefuelRecord{}).Where("user_id = ? AND vehicle_id = ?", userId, vehicleId)
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page, pageSize := normalizeRefuelPage(query)
	var records []RefuelRecord
	err := db.Order("refuel_time DESC, id DESC").
		Offset(int((page - 1) * pageSize)).
		Limit(int(pageSize)).
		Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	items := make([]*biz.RefuelRecord, 0, len(records))
	for i := range records {
		items = append(items, mapRefuelRecordToBiz(&records[i]))
	}
	return items, total, nil
}

func (r *RefuelRecordRepo) ListAllByUserIdAndVehicleId(ctx context.Context, userId string, vehicleId uint) ([]*biz.RefuelRecord, error) {
	var records []RefuelRecord
	err := r.data.db.WithContext(ctx).
		Where("user_id = ? AND vehicle_id = ?", userId, vehicleId).
		Order("refuel_time ASC, id ASC").
		Find(&records).Error
	if err != nil {
		return nil, err
	}

	items := make([]*biz.RefuelRecord, 0, len(records))
	for i := range records {
		items = append(items, mapRefuelRecordToBiz(&records[i]))
	}
	return items, nil
}

func parseFuelTime(value string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02 15:04:05", value, time.Local)
}

func normalizeFuelPage(query *biz.FuelVehicleListQuery) (int64, int64) {
	page := int64(1)
	pageSize := int64(10)
	if query != nil {
		if query.Page > 0 {
			page = query.Page
		}
		if query.PageSize > 0 {
			pageSize = query.PageSize
		}
	}
	return page, pageSize
}

func normalizeRefuelPage(query *biz.RefuelRecordListQuery) (int64, int64) {
	page := int64(1)
	pageSize := int64(10)
	if query != nil {
		if query.Page > 0 {
			page = query.Page
		}
		if query.PageSize > 0 {
			pageSize = query.PageSize
		}
	}
	return page, pageSize
}

func mapFuelVehicleToBiz(vehicle *FuelVehicle) *biz.FuelVehicle {
	if vehicle == nil {
		return nil
	}
	return &biz.FuelVehicle{
		Id:           int64(vehicle.ID),
		Name:         vehicle.Name,
		PlateNo:      vehicle.PlateNo,
		Brand:        vehicle.Brand,
		Model:        vehicle.VehicleModel,
		TankCapacity: vehicle.TankCapacity,
		Remark:       vehicle.Remark,
		UserId:       vehicle.UserId,
	}
}

func mapRefuelRecordToBiz(record *RefuelRecord) *biz.RefuelRecord {
	if record == nil {
		return nil
	}
	return &biz.RefuelRecord{
		Id:         int64(record.ID),
		VehicleId:  int64(record.VehicleId),
		RefuelTime: record.RefuelTime.Format("2006-01-02 15:04:05"),
		Odometer:   record.Odometer,
		Volume:     record.Volume,
		UnitPrice:  record.UnitPrice,
		Amount:     record.Amount,
		Station:    record.Station,
		IsFull:     record.IsFull,
		Remark:     record.Remark,
		UserId:     record.UserId,
	}
}
