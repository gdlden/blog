package service

import (
	"testing"

	pb "blog/api/fuel/v1"
	"blog/internal/biz"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func validRefuelRequest() *pb.RefuelRecord {
	return &pb.RefuelRecord{
		Id:           1,
		VehicleId:    1,
		RefuelTime:   "2026-01-01 08:00:00",
		Odometer:     "1000",
		Volume:       "30",
		UnitPrice:    "7.5",
		Amount:       "225.00",
		ActualAmount: "225.00",
		IsFull:       true,
	}
}

func TestRefuelRecordFromRequest_RejectsInvalidVolume(t *testing.T) {
	for _, volume := range []string{"", "0", "-5"} {
		req := validRefuelRequest()
		req.Volume = volume
		req.Amount = "0"

		_, err := refuelRecordFromRequest(req, true)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "油量必须大于 0")
	}
}

func TestRefuelRecordFromRequest_RejectsNegativeUnitPrice(t *testing.T) {
	req := validRefuelRequest()
	req.UnitPrice = "-1"

	_, err := refuelRecordFromRequest(req, true)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "单价不能为负数")
}

func TestRefuelRecordFromRequest_RejectsNegativeOdometer(t *testing.T) {
	req := validRefuelRequest()
	req.Odometer = "-100"

	_, err := refuelRecordFromRequest(req, true)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "总里程不能为负数")
}

func TestRefuelRecordFromRequest_RejectsMissingRefuelTime(t *testing.T) {
	req := validRefuelRequest()
	req.RefuelTime = ""

	_, err := refuelRecordFromRequest(req, true)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "加油时间不能为空")
}

func TestRefuelRecordFromRequest_RejectsMismatchedAmount(t *testing.T) {
	req := validRefuelRequest()
	req.Amount = "200.00" // 30 × 7.5 = 225.00

	_, err := refuelRecordFromRequest(req, true)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "金额与油量×单价不一致")
}

func TestRefuelRecordFromRequest_ComputesAmountWhenEmpty(t *testing.T) {
	req := validRefuelRequest()
	req.Amount = ""

	record, err := refuelRecordFromRequest(req, true)

	assert.NoError(t, err)
	assert.True(t, record.Amount.String() == "225" || record.Amount.String() == "225.00")
	assert.True(t, record.Amount.Equal(record.Volume.Mul(record.UnitPrice).Round(2)))
}

func TestRefuelRecordFromRequest_AcceptsValidRequest(t *testing.T) {
	record, err := refuelRecordFromRequest(validRefuelRequest(), true)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), record.Id)
	assert.Equal(t, int64(1), record.VehicleId)
	assert.Equal(t, "2026-01-01 08:00:00", record.RefuelTime)
	assert.True(t, record.IsFull)
}

func TestRefuelRecordFromRequest_RecalculatesAmountOnEdit(t *testing.T) {
	// 编辑存量记录：手输金额与计算值不一致时自动以油量×单价重算，不拒绝
	req := validRefuelRequest()
	req.Amount = "200.00" // 30 × 7.5 = 225.00

	record, err := refuelRecordFromRequest(req, false)

	assert.NoError(t, err)
	assert.True(t, record.Amount.Equal(decimal.RequireFromString("225.00")))
}

func TestRefuelRecordFromRequest_RejectsMissingActualAmount(t *testing.T) {
	req := validRefuelRequest()
	req.ActualAmount = ""

	_, err := refuelRecordFromRequest(req, true)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "实付金额不能为空")
}

func TestRefuelRecordFromRequest_RejectsNegativeActualAmount(t *testing.T) {
	req := validRefuelRequest()
	req.ActualAmount = "-1"

	_, err := refuelRecordFromRequest(req, true)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "实付金额不能为负数")
}

func TestRefuelRecordFromRequest_AcceptsActualAmountGreaterThanAmount(t *testing.T) {
	// 实付金额允许大于应付金额（加燃油宝等场景）
	req := validRefuelRequest()
	req.ActualAmount = "300.00"

	record, err := refuelRecordFromRequest(req, true)

	assert.NoError(t, err)
	assert.True(t, record.ActualAmount.Equal(decimal.RequireFromString("300.00")))
}

func TestRefuelRecordToReply_IncludesActualAmount(t *testing.T) {
	reply := refuelRecordToReply(&biz.RefuelRecord{
		Amount:       decimal.RequireFromString("225.00"),
		ActualAmount: decimal.RequireFromString("215.50"),
	})

	assert.Equal(t, "225", reply.Amount)
	assert.Equal(t, "215.5", reply.ActualAmount)
}

func TestFuelVehicleFromRequest_RejectsEmptyName(t *testing.T) {
	_, err := fuelVehicleFromRequest(&pb.FuelVehicle{Name: ""})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "车辆名称不能为空")
}

func TestFuelVehicleFromRequest_RejectsNegativeTankCapacity(t *testing.T) {
	_, err := fuelVehicleFromRequest(&pb.FuelVehicle{Name: "Car", TankCapacity: "-10"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "油箱容量不能为负数")
}
