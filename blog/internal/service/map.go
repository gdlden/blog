package service

import (
	"context"
	"strconv"

	mapV1 "blog/api/map/v1"
	"blog/internal/biz"
)

// MapService implements the Map proto service by translating proto requests to
// biz.MapUsecase calls and mapping the results back to proto replies. It
// follows the same pattern as PostService: the struct embeds the generated
// UnimplementedMapServer, holds a *biz.MapUsecase, and maps proto ↔ biz in
// each handler.
type MapService struct {
	mapV1.UnimplementedMapServer
	mu *biz.MapUsecase
}

func NewMapService(mu *biz.MapUsecase) *MapService {
	return &MapService{mu: mu}
}

// CreateSpot maps the proto request to a biz.Spot, persists it, and returns the
// saved spot as a SpotEntity. Coordinate strings are formatted from the proto
// doubles to mirror the biz.Spot string-typed coordinate convention.
func (s *MapService) CreateSpot(ctx context.Context, req *mapV1.CreateSpotRequest) (*mapV1.CreateSpotReply, error) {
	spot, _ := s.mu.CreateSpot(ctx, &biz.Spot{
		Name:      req.Name,
		Latitude:  strconv.FormatFloat(req.Latitude, 'f', -1, 64),
		Longitude: strconv.FormatFloat(req.Longitude, 'f', -1, 64),
		Notes:     req.Notes,
		Tags:      req.Tags,
		Photos:    req.Photos,
		Address:   req.Address,
	})
	if spot != nil {
		return &mapV1.CreateSpotReply{
			Spot: spotToEntity(spot),
		}, nil
	}
	return &mapV1.CreateSpotReply{}, nil
}

// ListSpots returns all stored spots as a list of SpotEntity.
func (s *MapService) ListSpots(ctx context.Context, req *mapV1.ListSpotsRequest) (*mapV1.ListSpotsReply, error) {
	spots, _ := s.mu.ListSpots(ctx)
	var entities []*mapV1.SpotEntity
	for _, spot := range spots {
		entities = append(entities, spotToEntity(spot))
	}
	return &mapV1.ListSpotsReply{Spots: entities}, nil
}

// UpdateSpot parses the id string to int64, builds a biz.Spot from the request,
// calls the usecase, and maps the result back to a proto reply.
func (s *MapService) UpdateSpot(ctx context.Context, req *mapV1.UpdateSpotRequest) (*mapV1.UpdateSpotReply, error) {
	id, _ := strconv.ParseInt(req.Id, 10, 64)
	spot, _ := s.mu.UpdateSpot(ctx, id, &biz.Spot{
		Id:        req.Id,
		Name:      req.Name,
		Latitude:  strconv.FormatFloat(req.Latitude, 'f', -1, 64),
		Longitude: strconv.FormatFloat(req.Longitude, 'f', -1, 64),
		Notes:     req.Notes,
		Tags:      req.Tags,
		Photos:    req.Photos,
		Address:   req.Address,
	})
	if spot != nil {
		return &mapV1.UpdateSpotReply{
			Spot: spotToEntity(spot),
		}, nil
	}
	return &mapV1.UpdateSpotReply{}, nil
}

// DeleteSpot parses the id string to int64, deletes via the usecase, and
// reports success mirroring PostService's err == nil convention.
func (s *MapService) DeleteSpot(ctx context.Context, req *mapV1.DeleteSpotRequest) (*mapV1.DeleteSpotReply, error) {
	id, _ := strconv.ParseInt(req.Id, 10, 64)
	err := s.mu.DeleteSpot(ctx, id)
	return &mapV1.DeleteSpotReply{Success: err == nil}, nil
}

// GetSpot parses the id string to int64, fetches via the usecase, and maps the
// result back to a proto reply.
func (s *MapService) GetSpot(ctx context.Context, req *mapV1.GetSpotRequest) (*mapV1.GetSpotReply, error) {
	id, _ := strconv.ParseInt(req.Id, 10, 64)
	spot, _ := s.mu.GetSpot(ctx, id)
	if spot != nil {
		return &mapV1.GetSpotReply{
			Spot: spotToEntity(spot),
		}, nil
	}
	return &mapV1.GetSpotReply{}, nil
}

// ReverseGeocode delegates to the usecase, passing longitude as the first (lng)
// argument and latitude as the second (lat) argument to match the Gaode API
// location=lng,lat ordering.
func (s *MapService) ReverseGeocode(ctx context.Context, req *mapV1.ReverseGeocodeRequest) (*mapV1.ReverseGeocodeReply, error) {
	address, _ := s.mu.ReverseGeocode(ctx, req.Longitude, req.Latitude)
	return &mapV1.ReverseGeocodeReply{Address: address}, nil
}

// spotToEntity converts a biz.Spot to a proto SpotEntity, parsing the string
// coordinates back to the doubles expected by the proto contract.
func spotToEntity(spot *biz.Spot) *mapV1.SpotEntity {
	lat, _ := strconv.ParseFloat(spot.Latitude, 64)
	lng, _ := strconv.ParseFloat(spot.Longitude, 64)
	return &mapV1.SpotEntity{
		Id:        spot.Id,
		Name:      spot.Name,
		Latitude:  lat,
		Longitude: lng,
		Notes:     spot.Notes,
		Tags:      spot.Tags,
		Photos:    spot.Photos,
		Address:   spot.Address,
		CreatedAt: spot.CreatedAt,
		UpdatedAt: spot.UpdatedAt,
	}
}
