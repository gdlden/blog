package service

import (
	"blog/internal/biz"
	"context"
	"errors"
	"github.com/shopspring/decimal"
	"strconv"

	pb "blog/api/debt/v1"
)

type DebtDetailService struct {
	ddu           *biz.DebtDetailUsecase
	ocrRecognizer VisionTextRecognizer
	pb.UnimplementedDebtDetailServer
}

func NewDebtDetailService(usecase *biz.DebtDetailUsecase) *DebtDetailService {
	return NewDebtDetailServiceWithRecognizer(usecase, NewArkVisionTextRecognizer(""))
}

func NewDebtDetailServiceWithRecognizer(usecase *biz.DebtDetailUsecase, recognizer VisionTextRecognizer) *DebtDetailService {
	if recognizer == nil {
		recognizer = NewArkVisionTextRecognizer("")
	}
	return &DebtDetailService{
		ddu:           usecase,
		ocrRecognizer: recognizer,
	}
}

func parseOptionalInt(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	return strconv.Atoi(s)
}

func parseOptionalDecimal(s string) (decimal.Decimal, error) {
	if s == "" {
		return decimal.Zero, nil
	}
	return decimal.NewFromString(s)
}

func (s *DebtDetailService) CreateDebtDetail(ctx context.Context, req *pb.DebtDetailData) (*pb.CreateDebtDetailReply, error) {
	debtId, err := strconv.Atoi(req.DebtId)
	if err != nil {
		return nil, err
	}
	period, err := parseOptionalInt(req.Period)
	if err != nil {
		return nil, err
	}
	principal, err := parseOptionalDecimal(req.Principal)
	if err != nil {
		return nil, err
	}
	interest, err := parseOptionalDecimal(req.Interest)
	if err != nil {
		return nil, err
	}
	id, err := s.ddu.Save(ctx, &biz.DebtDetail{
		DebtId:      uint(debtId),
		PostingDate: req.PostingDate,
		Principal:   principal,
		Interest:    interest,
		Period:      uint(period),
	})
	if err != nil {
		return nil, err
	}
	return &pb.CreateDebtDetailReply{
		Id: id,
	}, nil
}
func (s *DebtDetailService) UpdateDebtDetail(ctx context.Context, req *pb.DebtDetailData) (*pb.UpdateDebtDetailReply, error) {
	id, err := strconv.ParseUint(req.Id, 10, 64)
	if err != nil {
		return nil, errors.New("invalid debt detail id")
	}
	debtId, err := strconv.Atoi(req.DebtId)
	if err != nil {
		return nil, err
	}
	period, err := parseOptionalInt(req.Period)
	if err != nil {
		return nil, err
	}
	principal, err := parseOptionalDecimal(req.Principal)
	if err != nil {
		return nil, err
	}
	interest, err := parseOptionalDecimal(req.Interest)
	if err != nil {
		return nil, err
	}
	err = s.ddu.Edit(ctx, &biz.DebtDetail{
		Id:          uint(id),
		DebtId:      uint(debtId),
		PostingDate: req.PostingDate,
		Principal:   principal,
		Interest:    interest,
		Period:      uint(period),
	})
	if err != nil {
		return nil, err
	}
	return &pb.UpdateDebtDetailReply{Id: req.Id}, nil
}
func (s *DebtDetailService) DeleteDebtDetail(ctx context.Context, req *pb.DeleteDebtDetailRequest) (*pb.DeleteDebtDetailReply, error) {
	id, err := strconv.ParseUint(req.Id, 10, 64)
	if err != nil {
		return nil, errors.New("invalid debt detail id")
	}
	err = s.ddu.Delete(ctx, uint(id))
	if err != nil {
		return nil, err
	}
	return &pb.DeleteDebtDetailReply{Success: true}, nil
}
func (s *DebtDetailService) GetDebtDetail(ctx context.Context, req *pb.GetDebtDetailRequest) (*pb.DebtDetailData, error) {
	if req == nil || req.Id == "" {
		return nil, errors.New("invalid debt detail id")
	}
	id, err := strconv.ParseUint(req.Id, 10, 64)
	if err != nil || id == 0 {
		return nil, errors.New("invalid debt detail id")
	}
	detail, err := s.ddu.Get(ctx, uint(id))
	if err != nil {
		return nil, err
	}
	return debtDetailToReply(detail), nil
}
func (s *DebtDetailService) ListDebtDetail(ctx context.Context, req *pb.DebtDetailData) (*pb.ListDebtDetailReply, error) {
	debtId, err := strconv.ParseUint(req.DebtId, 10, 64)
	if err != nil {
		return nil, err
	}
	items, err := s.ddu.ListByDebtId(ctx, uint(debtId))
	if err != nil {
		return nil, err
	}
	list := make([]*pb.DebtDetailData, 0, len(items))
	for _, item := range items {
		list = append(list, debtDetailToReply(item))
	}
	return &pb.ListDebtDetailReply{List: list}, nil
}

func debtDetailToReply(detail *biz.DebtDetail) *pb.DebtDetailData {
	if detail == nil {
		return &pb.DebtDetailData{}
	}
	return &pb.DebtDetailData{
		Id:          strconv.FormatUint(uint64(detail.Id), 10),
		DebtId:      strconv.FormatUint(uint64(detail.DebtId), 10),
		PostingDate: detail.PostingDate,
		Principal:   detail.Principal.String(),
		Interest:    detail.Interest.String(),
		Period:      strconv.FormatUint(uint64(detail.Period), 10),
	}
}
