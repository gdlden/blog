package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFuelOCRText_ParsesStationScreenFields(t *testing.T) {
	fields := ParseFuelOCRText("金额: 225.00\n油量: 30.50\n单价: 7.38")

	assert.Equal(t, "225.00", fields.Amount)
	assert.Equal(t, "30.50", fields.Volume)
	assert.Equal(t, "7.38", fields.UnitPrice)
	assert.Equal(t, "", fields.Odometer)
}

func TestParseFuelOCRText_ParsesOdometer(t *testing.T) {
	fields := ParseFuelOCRText("总里程: 52134")

	assert.Equal(t, "52134", fields.Odometer)
	assert.Equal(t, "", fields.Amount)
	assert.Equal(t, "", fields.Volume)
	assert.Equal(t, "", fields.UnitPrice)
}

func TestParseFuelOCRText_GarbageReturnsEmptyFields(t *testing.T) {
	fields := ParseFuelOCRText("乱码文本没有可解析内容")

	assert.Equal(t, FuelOCRFields{}, fields)
}

func TestParseFuelOCRText_EmptyReturnsEmptyFields(t *testing.T) {
	fields := ParseFuelOCRText("")

	assert.Equal(t, FuelOCRFields{}, fields)
}

func TestParseFuelOCRText_MissingFieldsStayEmpty(t *testing.T) {
	fields := ParseFuelOCRText("金额: 100.00")

	assert.Equal(t, "100.00", fields.Amount)
	assert.Equal(t, "", fields.Volume)
	assert.Equal(t, "", fields.UnitPrice)
	assert.Equal(t, "", fields.Odometer)
}

func TestParseFuelOCRText_ToleratesCurrencySymbolsAndCommas(t *testing.T) {
	fields := ParseFuelOCRText("金额：¥1,234.50\n单价：￥7.38\n总里程: 52,134")

	assert.Equal(t, "1234.50", fields.Amount)
	assert.Equal(t, "7.38", fields.UnitPrice)
	assert.Equal(t, "52134", fields.Odometer)
}

func TestParseFuelOCRText_ToleratesExtraWhitespace(t *testing.T) {
	fields := ParseFuelOCRText("  金额  :   225.00  \n油量 :30.5")

	assert.Equal(t, "225.00", fields.Amount)
	assert.Equal(t, "30.5", fields.Volume)
}
