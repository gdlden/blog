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
	p, err := s.pc.UpdatePrice(ctx, &biz.Price{
		ID:        uint(req.Id),
		Name:      req.Name,
		Price:     req.Price,
		PriceDate: req.PriceDate,
	})
	if err != nil {
		return nil, err
	}
	return &pb.UpdatePriceReply{
		Info: &pb.PriceInfo{
			Id:        uint32(p.ID),
			Name:      p.Name,
			Price:     p.Price,
			PriceDate: p.PriceDate,
		},
	}, nil
}
func (s *PriceService) DeletePrice(ctx context.Context, req *pb.DeletePriceRequest) (*pb.DeletePriceReply, error) {
	if err := s.pc.DeletePrice(ctx, req.Id); err != nil {
		return nil, err
	}
	return &pb.DeletePriceReply{}, nil
}
func (s *PriceService) GetPrice(ctx context.Context, req *pb.GetPriceRequest) (*pb.GetPriceReply, error) {
	p, err := s.pc.GetPrice(ctx, int64(req.Id))
	if err != nil {
		return nil, err
	}
	return &pb.GetPriceReply{
		Info: &pb.PriceInfo{
			Id:        uint32(p.ID),
			Name:      p.Name,
			Price:     p.Price,
			PriceDate: p.PriceDate,
		},
	}, nil
}
func (s *PriceService) ListPrice(ctx context.Context, req *pb.ListPriceRequest) (*pb.ListPriceReply, error) {
	prices, err := s.pc.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	list := make([]*pb.PriceInfo, 0, len(prices))
	for _, p := range prices {
		list = append(list, &pb.PriceInfo{
			Id:        uint32(p.ID),
			Name:      p.Name,
			Price:     p.Price,
			PriceDate: p.PriceDate,
		})
	}
	return &pb.ListPriceReply{List: list}, nil
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
		data = append(data, &pb.PriceInfo{
			Id:        uint32(p.ID),
			Name:      p.Name,
			Price:     p.Price,
			PriceDate: p.PriceDate,
		})
	}
	return &pb.PagePriceReply{
		Current: req.Current,
		Size:    req.Size,
		Total:   strconv.FormatInt(total, 10),
		Data:    data,
	}, nil
}
