package service

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"
)

type DebtDetailOCRItem struct {
	DebtId      string `json:"debtId"`
	PostingDate string `json:"postingDate"`
	Principal   string `json:"principal"`
	Interest    string `json:"interest"`
	Period      string `json:"period"`
}

var debtDetailOCRLinePattern = regexp.MustCompile(`(?i)(?:第\s*)?(\d+)\s*期.*?本金\s*[¥￥]?\s*([0-9,]+(?:\.\d{1,2})?).*?利息\s*[¥￥]?\s*([0-9,]+(?:\.\d{1,2})?).*?入账日\s*((?:\d{4}[-/])?\d{1,2}[-/]\d{1,2})`)

func ParseDebtDetailOCRText(rawText string, debtId string, defaultYear int) ([]DebtDetailOCRItem, error) {
	lines := strings.Split(rawText, "\n")
	items := make([]DebtDetailOCRItem, 0, len(lines))
	for _, line := range lines {
		item, ok := parseDebtDetailOCRLine(line, debtId, defaultYear)
		if ok {
			items = append(items, item)
		}
	}
	if len(items) == 0 {
		item, ok := parseDebtDetailOCRLine(rawText, debtId, defaultYear)
		if ok {
			items = append(items, item)
		}
	}
	if len(items) == 0 {
		return nil, errors.New("no debt detail rows parsed")
	}
	return items, nil
}

func parseDebtDetailOCRLine(line string, debtId string, defaultYear int) (DebtDetailOCRItem, bool) {
	match := debtDetailOCRLinePattern.FindStringSubmatch(line)
	if len(match) != 5 {
		return DebtDetailOCRItem{}, false
	}
	postingDate, err := parseOCRPostingDate(match[4], defaultYear)
	if err != nil {
		return DebtDetailOCRItem{}, false
	}
	principal, err := normalizeOCRAmount(match[2])
	if err != nil {
		return DebtDetailOCRItem{}, false
	}
	interest, err := normalizeOCRAmount(match[3])
	if err != nil {
		return DebtDetailOCRItem{}, false
	}
	return DebtDetailOCRItem{
		DebtId:      debtId,
		Period:      match[1],
		Principal:   principal,
		Interest:    interest,
		PostingDate: postingDate,
	}, true
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
		return fmt.Sprintf("%04d-%02d-%02d 00:00:00", defaultYear, month, day), nil
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
		return fmt.Sprintf("%04d-%02d-%02d 00:00:00", year, month, day), nil
	}
	return "", fmt.Errorf("invalid posting date: %s", value)
}

func normalizeOCRAmount(value string) (string, error) {
	amount, err := decimal.NewFromString(strings.ReplaceAll(value, ",", ""))
	if err != nil {
		return "", err
	}
	return amount.StringFixed(2), nil
}
