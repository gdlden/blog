package service

import (
	"regexp"
	"strconv"
	"strings"
)

// FuelOCRFields 从模型输出文本中解析出的结构化字段；识别不到的字段保持空字符串。
type FuelOCRFields struct {
	Amount    string
	Volume    string
	UnitPrice string
	Odometer  string
}

var (
	fuelOCRAmountPattern    = regexp.MustCompile(`金额\s*[:：]?\s*[¥￥]?\s*([0-9,]+(?:\.\d+)?)`)
	fuelOCRVolumePattern    = regexp.MustCompile(`油量\s*[:：]?\s*([0-9,]+(?:\.\d+)?)`)
	fuelOCRUnitPricePattern = regexp.MustCompile(`单价\s*[:：]?\s*[¥￥]?\s*([0-9,]+(?:\.\d+)?)`)
	fuelOCROdometerPattern  = regexp.MustCompile(`总里程\s*[:：]?\s*([0-9,]+(?:\.\d+)?)`)
)

// ParseFuelOCRText 确定性解析模型输出；任何字段识别不到都留空，永不报错。
func ParseFuelOCRText(rawText string) FuelOCRFields {
	fields := FuelOCRFields{}
	if m := fuelOCRAmountPattern.FindStringSubmatch(rawText); len(m) == 2 {
		fields.Amount = normalizeFuelOCRNumber(m[1])
	}
	if m := fuelOCRVolumePattern.FindStringSubmatch(rawText); len(m) == 2 {
		fields.Volume = normalizeFuelOCRNumber(m[1])
	}
	if m := fuelOCRUnitPricePattern.FindStringSubmatch(rawText); len(m) == 2 {
		fields.UnitPrice = normalizeFuelOCRNumber(m[1])
	}
	if m := fuelOCROdometerPattern.FindStringSubmatch(rawText); len(m) == 2 {
		fields.Odometer = normalizeFuelOCRNumber(m[1])
	}
	return fields
}

func normalizeFuelOCRNumber(value string) string {
	value = strings.ReplaceAll(value, ",", "")
	if _, err := strconv.ParseFloat(value, 64); err != nil {
		return ""
	}
	return value
}
