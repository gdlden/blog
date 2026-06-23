package service

import (
	"context"
	"encoding/json"

	pb "blog/api/app/v1"
	"blog/internal/biz"
)

type AppService struct {
	pb.UnimplementedAppServer
	uc *biz.AppVersionUsecase
}

func NewAppService(uc *biz.AppVersionUsecase) *AppService {
	return &AppService{
		uc: uc,
	}
}

// GetVersion returns the latest active app version for public check.
func (s *AppService) GetVersion(ctx context.Context, req *pb.GetVersionRequest) (*pb.GetVersionReply, error) {
	av, err := s.uc.GetActive(ctx)
	if err != nil {
		return nil, err
	}
	return &pb.GetVersionReply{
		Version:    av.Version,
		Info:       av.Info,
		IosUrl:     av.IosUrl,
		AndroidUrl: av.AndroidUrl,
	}, nil
}

// ---- Admin Management ----

func (s *AppService) CreateAppVersion(ctx context.Context, req *pb.CreateAppVersionRequest) (*pb.CreateAppVersionReply, error) {
	av := &biz.AppVersion{
		Version:    req.Version,
		Info:       req.Info,
		IosUrl:     req.IosUrl,
		AndroidUrl: req.AndroidUrl,
		IsActive:   req.IsActive,
	}
	id, err := s.uc.Create(ctx, av)
	if err != nil {
		return nil, err
	}
	return &pb.CreateAppVersionReply{Id: int64(id)}, nil
}

func (s *AppService) UpdateAppVersion(ctx context.Context, req *pb.UpdateAppVersionRequest) (*pb.UpdateAppVersionReply, error) {
	av := &biz.AppVersion{
		Id:         uint(req.Id),
		Version:    req.Version,
		Info:       req.Info,
		IosUrl:     req.IosUrl,
		AndroidUrl: req.AndroidUrl,
		IsActive:   req.IsActive,
	}
	if err := s.uc.Update(ctx, av); err != nil {
		return nil, err
	}
	return &pb.UpdateAppVersionReply{Id: req.Id}, nil
}

func (s *AppService) DeleteAppVersion(ctx context.Context, req *pb.DeleteAppVersionRequest) (*pb.DeleteAppVersionReply, error) {
	if err := s.uc.Delete(ctx, uint(req.Id)); err != nil {
		return nil, err
	}
	return &pb.DeleteAppVersionReply{Success: true}, nil
}

func (s *AppService) GetAppVersion(ctx context.Context, req *pb.GetAppVersionRequest) (*pb.AppVersionEntity, error) {
	av, err := s.uc.Get(ctx, uint(req.Id))
	if err != nil {
		return nil, err
	}
	return appVersionToReply(av), nil
}

func (s *AppService) ListAppVersion(ctx context.Context, req *pb.ListAppVersionRequest) (*pb.ListAppVersionReply, error) {
	page := int(req.Page)
	pageSize := int(req.PageSize)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	items, total, err := s.uc.ListPage(ctx, &biz.AppVersionPageRequest{
		Current:  page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, err
	}
	list := make([]*pb.AppVersionEntity, 0, len(items))
	for _, item := range items {
		list = append(list, appVersionToReply(item))
	}
	return &pb.ListAppVersionReply{
		Page:     int64(page),
		PageSize: int64(pageSize),
		Total:    total,
		List:     list,
	}, nil
}

func appVersionToReply(av *biz.AppVersion) *pb.AppVersionEntity {
	if av == nil {
		return nil
	}
	infoJSON, _ := json.Marshal(av.Info)
	_ = infoJSON // not used in response, proto handles []string directly
	return &pb.AppVersionEntity{
		Id:         int64(av.Id),
		Version:    av.Version,
		Info:       av.Info,
		IosUrl:     av.IosUrl,
		AndroidUrl: av.AndroidUrl,
		IsActive:   av.IsActive,
		CreatedAt:  av.CreatedAt,
		UpdatedAt:  av.UpdatedAt,
	}
}
