package service

import (
	"blog/internal/biz"
	"context"
	"errors"
	"github.com/shopspring/decimal"
	"strconv"

	pb "blog/api/debt/v1"
)

type DebtService struct {
	pb.UnimplementedDebtServer
	du *biz.DebtUsecase
}

func NewDebtService(du *biz.DebtUsecase) *DebtService {
	return &DebtService{
		du: du,
	}
}

func parseDebtStatus(status string) (int, error) {
	switch status {
	case "进行中":
		return 0, nil
	case "已结清":
		return 1, nil
	}
	return strconv.Atoi(status)
}

func (s *DebtService) CreateDebt(ctx context.Context, req *pb.DebtEntity) (*pb.CreateDebtReply, error) {
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, err
	}
	tenor, err := decimal.NewFromString(req.Tenor)
	if err != nil {
		return nil, err
	}
	fee, err := decimal.NewFromString(req.Fee)
	if err != nil {
		return nil, err
	}
	apr, err := decimal.NewFromString(req.Apr)
	if err != nil {
		return nil, err
	}
	status, err := parseDebtStatus(req.Status)
	if err != nil {
		return nil, err
	}
	id, err := s.du.CreateDebt(ctx, &biz.Debt{
		Name:        req.Name,
		BankName:    req.BankName,
		BankAccount: req.BankAccount,
		ApplyTime:   req.ApplyTime,
		EndTime:     req.EndTime,
		Amount:      amount,
		Tenor:       tenor,
		Status:      status,
		Remark:      req.Remark,
		Apr:         apr,
		Fee:         fee,
	})
	if err != nil {
		return nil, err
	}
	return &pb.CreateDebtReply{
		Id:      strconv.FormatUint(uint64(id), 10),
		Message: "save success",
	}, nil
}

func (s *DebtService) UpdateDebt(ctx context.Context, req *pb.DebtEntity) (*pb.UpdateDebtReply, error) {
	if req == nil {
		return nil, errors.New("request is nil")
	}
	if req.Id <= 0 {
		return nil, errors.New("invalid debt id")
	}

	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, err
	}
	tenor, err := decimal.NewFromString(req.Tenor)
	if err != nil {
		return nil, err
	}
	fee, err := decimal.NewFromString(req.Fee)
	if err != nil {
		return nil, err
	}
	apr, err := decimal.NewFromString(req.Apr)
	if err != nil {
		return nil, err
	}
	status, err := parseDebtStatus(req.Status)
	if err != nil {
		return nil, err
	}
	id, err := s.du.Edit(ctx, &biz.Debt{
		Id:          req.Id,
		Name:        req.Name,
		BankName:    req.BankName,
		BankAccount: req.BankAccount,
		ApplyTime:   req.ApplyTime,
		EndTime:     req.EndTime,
		Amount:      amount,
		Tenor:       tenor,
		Status:      status,
		Remark:      req.Remark,
		Apr:         apr,
		Fee:         fee,
	})
	if err != nil {
		return nil, err
	}
	return &pb.UpdateDebtReply{
		Id:      strconv.FormatUint(uint64(id), 10),
		Message: "update success",
	}, nil
}

func (s *DebtService) DeleteDebt(ctx context.Context, req *pb.DeleteDebtRequest) (*pb.DeleteDebtReply, error) {
	if req == nil {
		return &pb.DeleteDebtReply{Flag: false}, errors.New("request is nil")
	}
	if req.Id == "" {
		return &pb.DeleteDebtReply{Flag: false}, errors.New("invalid debt id")
	}

	id, err := strconv.ParseUint(req.Id, 10, 64)
	if err != nil || id == 0 {
		return &pb.DeleteDebtReply{Flag: false}, errors.New("invalid debt id")
	}

	if err := s.du.Delete(ctx, uint(id)); err != nil {
		return &pb.DeleteDebtReply{Flag: false}, err
	}
	return &pb.DeleteDebtReply{Flag: true}, nil
}
func (s *DebtService) GetDebt(ctx context.Context, req *pb.GetDebtRequest) (*pb.DebtEntity, error) {
	if req == nil || req.Id == "" {
		return nil, errors.New("invalid debt id")
	}
	id, err := strconv.ParseUint(req.Id, 10, 64)
	if err != nil || id == 0 {
		return nil, errors.New("invalid debt id")
	}
	debt, err := s.du.GetDebt(ctx, uint(id))
	if err != nil {
		return nil, err
	}
	return debtToReply(debt), nil
}
func (s *DebtService) ListDebt(ctx context.Context, req *pb.ListDebtRequest) (*pb.ListDebtReply, error) {
	if req == nil {
		req = &pb.ListDebtRequest{}
	}
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	var status *int
	if req.Status != "" {
		parsedStatus, err := strconv.Atoi(req.Status)
		if err != nil {
			return nil, errors.New("invalid status")
		}
		status = &parsedStatus
	}

	items, total, err := s.du.ListDebt(ctx, &biz.DebtListQuery{
		Page:     page,
		PageSize: pageSize,
		Name:     req.Name,
		BankName: req.BankName,
		Status:   status,
	})
	if err != nil {
		return nil, err
	}

	list := make([]*pb.DebtEntity, 0, len(items))
	for _, debt := range items {
		list = append(list, debtToReply(debt))
	}
	return &pb.ListDebtReply{
		Page:  page,
		Total: total,
		List:  list,
	}, nil
}

func formatDebtStatus(status int) string {
	if status == 1 {
		return "已结清"
	}
	return "进行中"
}

func debtToReply(debt *biz.Debt) *pb.DebtEntity {
	if debt == nil {
		return &pb.DebtEntity{}
	}
	return &pb.DebtEntity{
		Id:          debt.Id,
		Name:        debt.Name,
		BankName:    debt.BankName,
		BankAccount: debt.BankAccount,
		ApplyTime:   debt.ApplyTime,
		EndTime:     debt.EndTime,
		Amount:      debt.Amount.String(),
		Status:      formatDebtStatus(debt.Status),
		Remark:      debt.Remark,
		Apr:         debt.Apr.String(),
		Fee:         debt.Fee.String(),
		Tenor:       debt.Tenor.String(),
	}
}
