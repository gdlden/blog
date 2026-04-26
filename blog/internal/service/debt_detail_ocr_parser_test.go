package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDebtDetailOCRText_MultipleRows(t *testing.T) {
	raw := `第1期 本金¥1,000.00 利息¥12.34 入账日03-15
第2期 本金￥900 利息￥10 入账日04-15 待入账`

	items, err := ParseDebtDetailOCRText(raw, "42", 2026)

	require.NoError(t, err)
	require.Len(t, items, 2)
	assert.Equal(t, "42", items[0].DebtId)
	assert.Equal(t, "1", items[0].Period)
	assert.Equal(t, "1000.00", items[0].Principal)
	assert.Equal(t, "12.34", items[0].Interest)
	assert.Equal(t, "2026-03-15 00:00:00", items[0].PostingDate)
	assert.Equal(t, "2", items[1].Period)
	assert.Equal(t, "900.00", items[1].Principal)
	assert.Equal(t, "10.00", items[1].Interest)
	assert.Equal(t, "2026-04-15 00:00:00", items[1].PostingDate)
}

func TestParseDebtDetailOCRText_FullDate(t *testing.T) {
	raw := "第12期 本金 88.8 利息 0.2 入账日2027-01-05"

	items, err := ParseDebtDetailOCRText(raw, "7", 2026)

	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "12", items[0].Period)
	assert.Equal(t, "88.80", items[0].Principal)
	assert.Equal(t, "0.20", items[0].Interest)
	assert.Equal(t, "2027-01-05 00:00:00", items[0].PostingDate)
}

func TestParseDebtDetailOCRText_LabelSeparatorsAndPostingDateLabel(t *testing.T) {
	raw := `第3期 本金: 1000 利息：10 入账日期 03-15
第4期 本金：2,000.5 利息: 20.25 入账日 2026/04/16`

	items, err := ParseDebtDetailOCRText(raw, "99", 2026)

	require.NoError(t, err)
	require.Len(t, items, 2)
	assert.Equal(t, "3", items[0].Period)
	assert.Equal(t, "1000.00", items[0].Principal)
	assert.Equal(t, "10.00", items[0].Interest)
	assert.Equal(t, "2026-03-15 00:00:00", items[0].PostingDate)
	assert.Equal(t, "4", items[1].Period)
	assert.Equal(t, "2000.50", items[1].Principal)
	assert.Equal(t, "20.25", items[1].Interest)
	assert.Equal(t, "2026-04-16 00:00:00", items[1].PostingDate)
}

func TestParseDebtDetailOCRText_InvalidDates(t *testing.T) {
	tests := []struct {
		name string
		raw  string
	}{
		{
			name: "invalid month day full date",
			raw:  "第1期 本金 1000 利息 10 入账日2026-99-99",
		},
		{
			name: "invalid leap day",
			raw:  "第1期 本金 1000 利息 10 入账日2026-02-31",
		},
		{
			name: "invalid default year date",
			raw:  "第1期 本金 1000 利息 10 入账日13-40",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, err := ParseDebtDetailOCRText(tt.raw, "1", 2026)

			require.Error(t, err)
			assert.Empty(t, items)
			assert.Contains(t, err.Error(), "invalid debt detail row")
		})
	}
}

func TestParseDebtDetailOCRText_MixedValidAndInvalidDateRows(t *testing.T) {
	raw := `第1期 本金 1000 利息 10 入账日2026-03-15
第2期 本金 900 利息 9 入账日2026-02-31`

	items, err := ParseDebtDetailOCRText(raw, "1", 2026)

	require.Error(t, err)
	assert.Empty(t, items)
	assert.Contains(t, err.Error(), "invalid debt detail row")
}

func TestParseDebtDetailOCRText_NoRows(t *testing.T) {
	items, err := ParseDebtDetailOCRText("无法识别的文本", "1", 2026)

	require.Error(t, err)
	assert.Empty(t, items)
	assert.Contains(t, err.Error(), "no debt detail rows parsed")
}
