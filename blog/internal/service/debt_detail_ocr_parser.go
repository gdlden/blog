package service

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type DebtDetailOCRItem struct {
	DebtId      string `json:"debtId"`
	PostingDate string `json:"postingDate"`
	Principal   string `json:"principal"`
	Interest    string `json:"interest"`
	Period      string `json:"period"`
}

var debtDetailOCRLinePattern = regexp.MustCompile(`(?i)(?:第\s*)?(\d+)\s*期.*?本金\s*[:：]?\s*[¥￥]?\s*([0-9,]+(?:\.\d{1,2})?).*?利息\s*[:：]?\s*[¥￥]?\s*([0-9,]+(?:\.\d{1,2})?).*?入账日(?:期)?\s*[:：]?\s*((?:\d{4}[-/])?\d{1,2}[-/]\d{1,2})`)

func ParseDebtDetailOCRText(rawText string, debtId string, defaultYear int) ([]DebtDetailOCRItem, error) {
	lines := strings.Split(rawText, "\n")
	items := make([]DebtDetailOCRItem, 0, len(lines))
	for _, line := range lines {
		item, ok, err := parseDebtDetailOCRLine(line, debtId, defaultYear)
		if err != nil {
			return nil, err
		}
		if ok {
			items = append(items, item)
		}
	}
	if len(items) == 0 {
		item, ok, err := parseDebtDetailOCRLine(rawText, debtId, defaultYear)
		if err != nil {
			return nil, err
		}
		if ok {
			items = append(items, item)
		}
	}
	if len(items) == 0 {
		return nil, errors.New("no debt detail rows parsed")
	}
	return items, nil
}

func parseDebtDetailOCRLine(line string, debtId string, defaultYear int) (DebtDetailOCRItem, bool, error) {
	match := debtDetailOCRLinePattern.FindStringSubmatch(line)
	if len(match) != 5 {
		return DebtDetailOCRItem{}, false, nil
	}
	postingDate, err := parseOCRPostingDate(match[4], defaultYear)
	if err != nil {
		return DebtDetailOCRItem{}, false, fmt.Errorf("invalid debt detail row: %w", err)
	}
	principal, err := normalizeOCRAmount(match[2])
	if err != nil {
		return DebtDetailOCRItem{}, false, fmt.Errorf("invalid debt detail row: %w", err)
	}
	interest, err := normalizeOCRAmount(match[3])
	if err != nil {
		return DebtDetailOCRItem{}, false, fmt.Errorf("invalid debt detail row: %w", err)
	}
	return DebtDetailOCRItem{
		DebtId:      debtId,
		Period:      match[1],
		Principal:   principal,
		Interest:    interest,
		PostingDate: postingDate,
	}, true, nil
}

func parseOCRPostingDate(value string, defaultYear int) (string, error) {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == '-' || r == '/'
	})
	if len(parts) == 2 {
		month, err := strconv.Atoi(parts[0])
		if err != nil {
			return "", err
		}
		day, err := strconv.Atoi(parts[1])
		if err != nil {
			return "", err
		}
		return formatValidOCRPostingDate(defaultYear, month, day)
	}
	if len(parts) == 3 {
		year, err := strconv.Atoi(parts[0])
		if err != nil {
			return "", err
		}
		month, err := strconv.Atoi(parts[1])
		if err != nil {
			return "", err
		}
		day, err := strconv.Atoi(parts[2])
		if err != nil {
			return "", err
		}
		return formatValidOCRPostingDate(year, month, day)
	}
	return "", fmt.Errorf("invalid posting date: %s", value)
}

func formatValidOCRPostingDate(year int, month int, day int) (string, error) {
	normalizedDate := fmt.Sprintf("%04d-%02d-%02d", year, month, day)
	parsedDate, err := time.Parse("2006-01-02", normalizedDate)
	if err != nil {
		return "", err
	}
	if parsedDate.Format("2006-01-02") != normalizedDate {
		return "", fmt.Errorf("invalid posting date: %s", normalizedDate)
	}
	return normalizedDate + " 00:00:00", nil
}

func normalizeOCRAmount(value string) (string, error) {
	amount, err := decimal.NewFromString(strings.ReplaceAll(value, ",", ""))
	if err != nil {
		return "", err
	}
	return amount.StringFixed(2), nil
}
