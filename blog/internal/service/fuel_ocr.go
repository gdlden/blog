package service

import (
	"context"
	stderrors "errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"strings"

	kerrors "github.com/go-kratos/kratos/v2/errors"
	kratoshttp "github.com/go-kratos/kratos/v2/transport/http"
)

const (
	fuelOCRAttachTypeStationScreen = "station_screen"
	fuelOCRAttachTypeDashboard     = "dashboard"
)

type FuelOCRRequest struct {
	AttachType string
	Image      multipart.File
	Header     *multipart.FileHeader
}

type FuelOCRReply struct {
	RawText   string `json:"rawText"`
	Amount    string `json:"amount"`
	Volume    string `json:"volume"`
	UnitPrice string `json:"unitPrice"`
	Odometer  string `json:"odometer"`
}

func fuelOCRPrompt(attachType string) string {
	if attachType == fuelOCRAttachTypeDashboard {
		return `请识别车辆仪表盘图片中的总里程（ODO，单位 km），并只输出一行可解析的文本。
输出格式示例：总里程: 52134
不要输出解释、表格 Markdown 或无关内容。`
	}
	return `请识别加油机屏幕图片中的金额、油量、单价，并只输出可解析的文本行。
输出格式示例：
金额: 225.00
油量: 30.50
单价: 7.38
不要输出解释、表格 Markdown 或无关内容。`
}

func (s *FuelService) RecognizeFuelOCR(ctx context.Context, req *FuelOCRRequest) (*FuelOCRReply, error) {
	if req == nil {
		return nil, badFuelOCRRequest("invalid fuel ocr request")
	}
	if req.Image != nil {
		defer func() {
			_ = req.Image.Close()
			req.Image = nil
		}()
	}
	if s == nil || s.ocrRecognizer == nil {
		return nil, stderrors.New("fuel ocr recognizer unavailable")
	}
	attachType := strings.TrimSpace(req.AttachType)
	if attachType != fuelOCRAttachTypeStationScreen && attachType != fuelOCRAttachTypeDashboard {
		return nil, badFuelOCRRequest("invalid attach type")
	}
	if req.Image == nil || req.Header == nil {
		return nil, badFuelOCRRequest("invalid ocr image")
	}
	imageDataURI, err := buildOCRImageDataURI(req.Image, req.Header)
	if err != nil {
		return nil, err
	}
	rawText, err := s.ocrRecognizer.RecognizeText(ctx, imageDataURI, fuelOCRPrompt(attachType))
	if err != nil {
		return nil, fmt.Errorf("fuel ocr failed: %w", err)
	}
	fields := ParseFuelOCRText(rawText)
	reply := &FuelOCRReply{RawText: rawText}
	if attachType == fuelOCRAttachTypeStationScreen {
		reply.Amount = fields.Amount
		reply.Volume = fields.Volume
		reply.UnitPrice = fields.UnitPrice
	} else {
		reply.Odometer = fields.Odometer
	}
	return reply, nil
}

func (s *FuelService) RecognizeFuelOCRHTTP(ctx kratoshttp.Context) error {
	kratoshttp.SetOperation(ctx, "/api.fuel.v1.Fuel/RecognizeFuelOCR")
	ctx.Request().Body = http.MaxBytesReader(ctx.Response(), ctx.Request().Body, debtDetailOCRMaxRequestBytes)
	in, err := parseFuelOCRHTTPRequest(ctx.Request())
	if err != nil {
		cleanupFuelOCRRequest(ctx.Request(), nil)
		return err
	}
	defer cleanupFuelOCRRequest(ctx.Request(), in)
	h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
		return s.RecognizeFuelOCR(ctx, req.(*FuelOCRRequest))
	})
	out, err := h(ctx, in)
	if err != nil {
		return err
	}
	return ctx.Result(http.StatusOK, out.(*FuelOCRReply))
}

func parseFuelOCRHTTPRequest(r *http.Request) (*FuelOCRRequest, error) {
	if err := r.ParseMultipartForm(debtDetailOCRMaxImageBytes); err != nil {
		return nil, badFuelOCRRequest(err.Error())
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		return nil, badFuelOCRRequest("invalid ocr image")
	}
	return &FuelOCRRequest{
		AttachType: strings.TrimSpace(r.FormValue("attachType")),
		Image:      file,
		Header:     header,
	}, nil
}

func cleanupFuelOCRRequest(r *http.Request, req *FuelOCRRequest) {
	if req != nil && req.Image != nil {
		_ = req.Image.Close()
		req.Image = nil
	}
	if r != nil && r.MultipartForm != nil {
		_ = r.MultipartForm.RemoveAll()
	}
}

func badFuelOCRRequest(message string) error {
	return kerrors.BadRequest("FUEL_OCR_BAD_REQUEST", message)
}
