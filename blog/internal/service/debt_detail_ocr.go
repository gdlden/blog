package service

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
)

const debtDetailOCRMaxImageBytes = 8 << 20

type DebtDetailOCRRequest struct {
	DebtId string
	Image  multipart.File
	Header *multipart.FileHeader
	Year   int
}

type DebtDetailOCRReply struct {
	RawText string              `json:"rawText"`
	Items   []DebtDetailOCRItem `json:"items"`
}

func debtDetailOCRPrompt() string {
	return `请识别图片中的分期还款明细，并只输出可解析的文本行。
每行包含：期数、本金、利息、入账日期。
请优先识别“入账日”“入账日期”“待入账”等字段附近的日期。
输出格式示例：第1期 本金: 1000.00 利息: 12.34 入账日: 03-25
不要输出解释、表格 Markdown 或无关内容。`
}

func (s *DebtDetailService) RecognizeDebtDetailOCR(ctx context.Context, req *DebtDetailOCRRequest) (*DebtDetailOCRReply, error) {
	if req == nil {
		return nil, errors.New("invalid debt detail ocr request")
	}
	req.DebtId = strings.TrimSpace(req.DebtId)
	if req.DebtId == "" {
		return nil, errors.New("invalid debt id")
	}
	if req.Image == nil || req.Header == nil {
		return nil, errors.New("invalid ocr image")
	}
	year := req.Year
	if year == 0 {
		year = time.Now().Year()
	}
	imageDataURI, err := buildOCRImageDataURI(req.Image, req.Header)
	if err != nil {
		return nil, err
	}
	rawText, err := s.ocrRecognizer.RecognizeText(ctx, imageDataURI, debtDetailOCRPrompt())
	if err != nil {
		return nil, fmt.Errorf("debt detail ocr failed: %w", err)
	}
	items, err := ParseDebtDetailOCRText(rawText, req.DebtId, year)
	if err != nil {
		return &DebtDetailOCRReply{RawText: rawText, Items: []DebtDetailOCRItem{}}, nil
	}
	return &DebtDetailOCRReply{RawText: rawText, Items: items}, nil
}

func (s *DebtDetailService) RecognizeDebtDetailOCRHTTP(ctx kratoshttp.Context) error {
	kratoshttp.SetOperation(ctx, "/api.debt.v1.DebtDetail/RecognizeDebtDetailOCR")
	in, err := parseDebtDetailOCRHTTPRequest(ctx.Request())
	if err != nil {
		return err
	}
	h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
		return s.RecognizeDebtDetailOCR(ctx, req.(*DebtDetailOCRRequest))
	})
	out, err := h(ctx, in)
	if err != nil {
		return err
	}
	return ctx.Result(http.StatusOK, out.(*DebtDetailOCRReply))
}

func parseDebtDetailOCRHTTPRequest(r *http.Request) (*DebtDetailOCRRequest, error) {
	if err := r.ParseMultipartForm(debtDetailOCRMaxImageBytes); err != nil {
		return nil, err
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		return nil, err
	}
	year, err := parseOCRYear(r.FormValue("year"))
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	return &DebtDetailOCRRequest{
		DebtId: strings.TrimSpace(r.FormValue("debtId")),
		Image:  file,
		Header: header,
		Year:   year,
	}, nil
}

func parseOCRYear(value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	year, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	if year < 2000 || year > 2100 {
		return 0, fmt.Errorf("invalid year: %d", year)
	}
	return year, nil
}

func buildOCRImageDataURI(file multipart.File, header *multipart.FileHeader) (string, error) {
	defer file.Close()
	if header == nil {
		return "", errors.New("invalid ocr image")
	}
	if header.Size > debtDetailOCRMaxImageBytes {
		return "", fmt.Errorf("image exceeds %d bytes", debtDetailOCRMaxImageBytes)
	}
	data, err := io.ReadAll(io.LimitReader(file, debtDetailOCRMaxImageBytes+1))
	if err != nil {
		return "", err
	}
	if len(data) == 0 {
		return "", errors.New("empty ocr image")
	}
	if len(data) > debtDetailOCRMaxImageBytes {
		return "", fmt.Errorf("image exceeds %d bytes", debtDetailOCRMaxImageBytes)
	}
	contentType := http.DetectContentType(data)
	if !strings.HasPrefix(contentType, "image/") {
		return "", fmt.Errorf("invalid image content type: %s", contentType)
	}
	return "data:" + contentType + ";base64," + base64.StdEncoding.EncodeToString(data), nil
}
