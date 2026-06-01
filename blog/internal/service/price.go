package service

import (
	"context"
	"strconv"

	pb "blog/api/price/v1"
	"blog/internal/biz"
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
	id := s.pc.CreatePrice(ctx, &biz.Price{
		Name:      req.Name,
		Price:     req.Price,
		PriceDate: req.PriceDate,
	})
	rId := strconv.FormatUint(uint64(id), 10)
	return &pb.CreatePriceReply{
		Id: rId,
	}, nil
}
func (s *PriceService) UpdatePrice(ctx context.Context, req *pb.UpdatePriceRequest) (*pb.UpdatePriceReply, error) {
	return &pb.UpdatePriceReply{}, nil
}
func (s *PriceService) DeletePrice(ctx context.Context, req *pb.DeletePriceRequest) (*pb.DeletePriceReply, error) {
	return &pb.DeletePriceReply{}, nil
}
func (s *PriceService) GetPrice(ctx context.Context, req *pb.GetPriceRequest) (*pb.GetPriceReply, error) {
	return &pb.GetPriceReply{}, nil
}
func (s *PriceService) ListPrice(ctx context.Context, req *pb.ListPriceRequest) (*pb.ListPriceReply, error) {
	prices, err := s.pc.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]*pb.PriceInfo, 0, len(prices))
	for _, p := range prices {
		list = append(list, &pb.PriceInfo{
			Name:      p.Name,
			Price:     p.Price,
			PriceDate: p.PriceDate,
		})
	}
	return &pb.ListPriceReply{List: list}, nil
}
