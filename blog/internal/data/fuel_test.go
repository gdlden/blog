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
	if err := db.AutoMigrate(&FuelVehicle{}, &RefuelRecord{}, &FuelAttachment{}); err != nil {
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

func TestFuelVehicleRepo_ListByUserIdSearchesNameOrPlate(t *testing.T) {
	db := setupFuelTestDB(t)
	repo := &FuelVehicleRepo{data: &Data{db: db}, log: log.NewHelper(log.DefaultLogger)}
	ctx := context.Background()

	for _, vehicle := range []*biz.FuelVehicle{
		{Name: "家用车", PlateNo: "沪A12345", UserId: "user-123"},
		{Name: "商务车", PlateNo: "京B67890", UserId: "user-123"},
	} {
		_, err := repo.Save(ctx, vehicle)
		assert.NoError(t, err)
	}

	// 按车牌搜索（名称不含关键字），应命中 OR 匹配
	items, total, err := repo.ListByUserId(ctx, "user-123", &biz.FuelVehicleListQuery{Page: 1, PageSize: 10, PlateNo: "京B"})

	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, items, 1)
	assert.Equal(t, "商务车", items[0].Name)
}

func TestFuelVehicleRepo_CountRefuelRecordByVehicleId(t *testing.T) {
	db := setupFuelTestDB(t)
	vehicleRepo := &FuelVehicleRepo{data: &Data{db: db}, log: log.NewHelper(log.DefaultLogger)}
	recordRepo := &RefuelRecordRepo{data: &Data{db: db}, log: log.NewHelper(log.DefaultLogger)}
	ctx := context.Background()

	vehicleId, err := vehicleRepo.Save(ctx, &biz.FuelVehicle{Name: "Car", UserId: "user-123"})
	assert.NoError(t, err)
	_, err = recordRepo.Save(ctx, &biz.RefuelRecord{
		VehicleId:    int64(vehicleId),
		RefuelTime:   "2026-01-01 00:00:00",
		Odometer:     decimal.NewFromInt(1000),
		Volume:       decimal.NewFromInt(30),
		UnitPrice:    decimal.NewFromInt(7),
		Amount:       decimal.NewFromInt(210),
		ActualAmount: decimal.NewFromInt(200),
		IsFull:       true,
		UserId:       "user-123",
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
		{VehicleId: int64(vehicleId), RefuelTime: "2026-01-01 00:00:00", Odometer: decimal.NewFromInt(1000), Volume: decimal.NewFromInt(30), Amount: decimal.NewFromInt(210), ActualAmount: decimal.NewFromInt(200), UserId: "user-123"},
		{VehicleId: int64(vehicleId), RefuelTime: "2026-02-01 00:00:00", Odometer: decimal.NewFromInt(1500), Volume: decimal.NewFromInt(35), Amount: decimal.NewFromInt(245), ActualAmount: decimal.NewFromInt(235), UserId: "user-123"},
	} {
		_, err := recordRepo.Save(ctx, record)
		assert.NoError(t, err)
	}

	items, total, err := recordRepo.ListByUserIdAndVehicleId(ctx, "user-123", vehicleId, &biz.RefuelRecordListQuery{Page: 1, PageSize: 10})

	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, items, 2)
	assert.Equal(t, "2026-02-01 00:00:00", items[0].RefuelTime)
	assert.True(t, items[0].ActualAmount.Equal(decimal.NewFromInt(235)))
}

func TestRefuelRecordRepo_SaveListAndReplaceAttachments(t *testing.T) {
	db := setupFuelTestDB(t)
	if err := db.AutoMigrate(&fileRecord{}); err != nil {
		t.Fatalf("failed to migrate file_records: %v", err)
	}
	vehicleRepo := &FuelVehicleRepo{data: &Data{db: db}, log: log.NewHelper(log.DefaultLogger)}
	recordRepo := &RefuelRecordRepo{data: &Data{db: db}, log: log.NewHelper(log.DefaultLogger)}
	fileRepo := &fileRepo{data: &Data{db: db}, log: log.NewHelper(log.DefaultLogger)}
	ctx := context.Background()

	vehicleId, err := vehicleRepo.Save(ctx, &biz.FuelVehicle{Name: "Car", UserId: "user-123"})
	assert.NoError(t, err)
	fileId1, err := fileRepo.Save(ctx, &biz.FileRecord{UserId: "user-123", FileName: "receipt.jpg", FileType: "image/jpeg", FileExt: ".jpg", FileUrl: "/files/2026/01/01/receipt.jpg", FileSize: 1024})
	assert.NoError(t, err)
	fileId2, err := fileRepo.Save(ctx, &biz.FileRecord{UserId: "user-123", FileName: "scene.png", FileType: "image/png", FileExt: ".png", FileUrl: "/files/2026/01/01/scene.png", FileSize: 2048})
	assert.NoError(t, err)

	// 保存记录并带两张附件
	recordId, err := recordRepo.Save(ctx, &biz.RefuelRecord{
		VehicleId:    int64(vehicleId),
		RefuelTime:   "2026-03-01 00:00:00",
		Odometer:     decimal.NewFromInt(2000),
		Volume:       decimal.NewFromInt(40),
		Amount:       decimal.NewFromInt(280),
		ActualAmount: decimal.NewFromInt(270),
		UserId:       "user-123",
		Attachments: []*biz.FuelAttachment{
			{FileId: fileId1, AttachType: "station_screen", Sort: 0},
			{FileId: fileId2, AttachType: "dashboard", Sort: 1},
		},
	})
	assert.NoError(t, err)

	item, err := recordRepo.FindByUserIdAndRecordId(ctx, "user-123", recordId)
	assert.NoError(t, err)
	assert.Len(t, item.Attachments, 2)
	assert.Equal(t, "receipt.jpg", item.Attachments[0].FileName)
	assert.True(t, item.ActualAmount.Equal(decimal.NewFromInt(270)))
	assert.Equal(t, "/file/download/v1/1", item.Attachments[0].FileUrl)
	assert.Equal(t, "station_screen", item.Attachments[0].AttachType)
	assert.Equal(t, int32(0), item.Attachments[0].Sort)
	assert.Equal(t, "scene.png", item.Attachments[1].FileName)
	assert.Equal(t, int32(1), item.Attachments[1].Sort)

	// 列表同样带附件信息
	list, total, err := recordRepo.ListByUserIdAndVehicleId(ctx, "user-123", vehicleId, &biz.RefuelRecordListQuery{Page: 1, PageSize: 10})
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, list, 1)
	assert.Len(t, list[0].Attachments, 2)

	// 整组替换：只保留第二张
	err = recordRepo.Update(ctx, &biz.RefuelRecord{
		Id:           int64(recordId),
		VehicleId:    int64(vehicleId),
		RefuelTime:   "2026-03-01 00:00:00",
		Odometer:     decimal.NewFromInt(2000),
		Volume:       decimal.NewFromInt(40),
		Amount:       decimal.NewFromInt(280),
		ActualAmount: decimal.NewFromInt(270),
		UserId:       "user-123",
		Attachments:  []*biz.FuelAttachment{{FileId: fileId2, AttachType: "other", Sort: 0}},
	})
	assert.NoError(t, err)
	item, err = recordRepo.FindByUserIdAndRecordId(ctx, "user-123", recordId)
	assert.NoError(t, err)
	assert.Len(t, item.Attachments, 1)
	assert.Equal(t, "other", item.Attachments[0].AttachType)
	assert.Equal(t, "scene.png", item.Attachments[0].FileName)

	// Update 不传 attachments 时保持现有附件不变
	err = recordRepo.Update(ctx, &biz.RefuelRecord{
		Id:           int64(recordId),
		VehicleId:    int64(vehicleId),
		RefuelTime:   "2026-03-02 00:00:00",
		Odometer:     decimal.NewFromInt(2050),
		Volume:       decimal.NewFromInt(40),
		Amount:       decimal.NewFromInt(280),
		ActualAmount: decimal.NewFromInt(270),
		UserId:       "user-123",
	})
	assert.NoError(t, err)
	item, err = recordRepo.FindByUserIdAndRecordId(ctx, "user-123", recordId)
	assert.NoError(t, err)
	assert.Len(t, item.Attachments, 1)

	// 删除记录：附件关联一并清理，file_records 保留
	err = recordRepo.DeleteByUserIdAndRecordId(ctx, "user-123", recordId)
	assert.NoError(t, err)
	var attCount int64
	db.Model(&FuelAttachment{}).Where("record_id = ?", recordId).Count(&attCount)
	assert.Equal(t, int64(0), attCount)
	var fileCount int64
	db.Model(&fileRecord{}).Count(&fileCount)
	assert.Equal(t, int64(2), fileCount)
}
