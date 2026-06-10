package service

import (
	"context"
	"strconv"

	pb "blog/api/price/v1"
	"blog/internal/biz"

	"github.com/shopspring/decimal"
)

type PriceService struct {
	pb.UnimplementedPriceServer
	pc *biz.PriceUscase
}

func NewPriceService(pcu *biz.PriceUscase) *PriceService {
	return &PriceService{
		pc: pcu,
	}
}

func (s *PriceService) CreatePrice(ctx context.Context, req *pb.CreatePriceRequest) (*pb.CreatePriceReply, error) {
	weight, _ := decimal.NewFromString(req.Weight)
	unitPrice, _ := decimal.NewFromString(req.UnitPrice)
	id, err := s.pc.CreatePrice(ctx, &biz.Price{
		ProductName: req.ProductName,
		Weight:      weight,
		UnitPrice:   unitPrice,
		PriceDate:   req.PriceDate,
	})
	if err != nil {
		return nil, err
	}
	return &pb.CreatePriceReply{
		Id: strconv.FormatUint(uint64(id), 10),
	}, nil
}

func (s *PriceService) UpdatePrice(ctx context.Context, req *pb.UpdatePriceRequest) (*pb.UpdatePriceReply, error) {
	weight, _ := decimal.NewFromString(req.Weight)
	unitPrice, _ := decimal.NewFromString(req.UnitPrice)
	err := s.pc.UpdatePrice(ctx, &biz.Price{
		ID:          uint(req.Id),
		ProductName: req.ProductName,
		Weight:      weight,
		UnitPrice:   unitPrice,
		PriceDate:   req.PriceDate,
	})
	if err != nil {
		return nil, err
	}
	return &pb.UpdatePriceReply{
		Info: priceToInfo(&biz.Price{
			ID:          uint(req.Id),
			ProductName: req.ProductName,
			Weight:      weight,
			UnitPrice:   unitPrice,
			PriceDate:   req.PriceDate,
			TotalPrice:  weight.Mul(unitPrice).Round(2),
		}),
	}, nil
}

func (s *PriceService) DeletePrice(ctx context.Context, req *pb.DeletePriceRequest) (*pb.DeletePriceReply, error) {
	if err := s.pc.DeletePrice(ctx, uint(req.Id)); err != nil {
		return nil, err
	}
	return &pb.DeletePriceReply{}, nil
}

func (s *PriceService) GetPrice(ctx context.Context, req *pb.GetPriceRequest) (*pb.GetPriceReply, error) {
	p, err := s.pc.GetPrice(ctx, uint(req.Id))
	if err != nil {
		return nil, err
	}
	return &pb.GetPriceReply{
		Info: priceToInfo(p),
	}, nil
}

func (s *PriceService) ListPrice(ctx context.Context, req *pb.ListPriceRequest) (*pb.ListPriceReply, error) {
	return &pb.ListPriceReply{}, nil
}

func (s *PriceService) PagePrice(ctx context.Context, req *pb.ListPriceRequest) (*pb.PagePriceReply, error) {
	current, _ := strconv.Atoi(req.Current)
	size, _ := strconv.Atoi(req.Size)
	if current < 1 {
		current = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	prices, total, err := s.pc.GetPricePage(ctx, &biz.PricePageRequest{
		Current: current,
		Size:    size,
	})
	if err != nil {
		return nil, err
	}
	data := make([]*pb.PriceInfo, 0, len(prices))
	for _, p := range prices {
		data = append(data, priceToInfo(p))
	}
	return &pb.PagePriceReply{
		Current: req.Current,
		Size:    req.Size,
		Total:   strconv.FormatInt(total, 10),
		Data:    data,
	}, nil
}

func priceToInfo(p *biz.Price) *pb.PriceInfo {
	if p == nil {
		return nil
	}
	return &pb.PriceInfo{
		Id:          uint32(p.ID),
		ProductName: p.ProductName,
		Weight:      p.Weight.String(),
		UnitPrice:   p.UnitPrice.String(),
		TotalPrice:  p.TotalPrice.String(),
		PriceDate:   p.PriceDate,
	}
}
