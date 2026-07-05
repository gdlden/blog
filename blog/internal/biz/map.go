package biz

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

// Spot is the domain model for a fishing spot. Coordinates are kept as strings
// for proto-friendliness mirroring the Post domain struct convention.
type Spot struct {
	Id        string
	Name      string
	Latitude  string
	Longitude string
	Notes     string
	Tags      string
	Photos    []string
	Address   string
	CreatedAt string
	UpdatedAt string
}

// MapRepo abstracts the persistence operations for Spot.
type MapRepo interface {
	Save(context.Context, *Spot) (*Spot, error)
	Update(context.Context, *Spot) (*Spot, error)
	Delete(context.Context, int64) error
	FindByID(context.Context, int64) (*Spot, error)
	ListAll(context.Context) ([]*Spot, error)
}

type MapUsecase struct {
	repo MapRepo
	log  *log.Helper
}

func NewMapUsecase(repo MapRepo, logger log.Logger) *MapUsecase {
	return &MapUsecase{repo: repo, log: log.NewHelper(logger)}
}

func (uc *MapUsecase) CreateSpot(ctx context.Context, g *Spot) (*Spot, error) {
	uc.log.WithContext(ctx).Infof("CreateSpot: %v", g.Name)
	return uc.repo.Save(ctx, g)
}

func (uc *MapUsecase) UpdateSpot(ctx context.Context, id int64, g *Spot) (*Spot, error) {
	uc.log.WithContext(ctx).Infof("UpdateSpot: %v", id)
	return uc.repo.Update(ctx, g)
}

func (uc *MapUsecase) DeleteSpot(ctx context.Context, id int64) error {
	uc.log.WithContext(ctx).Infof("DeleteSpot: %v", id)
	return uc.repo.Delete(ctx, id)
}

func (uc *MapUsecase) GetSpot(ctx context.Context, id int64) (*Spot, error) {
	uc.log.WithContext(ctx).Infof("GetSpot: %v", id)
	return uc.repo.FindByID(ctx, id)
}

func (uc *MapUsecase) ListSpots(ctx context.Context) ([]*Spot, error) {
	uc.log.WithContext(ctx).Infof("ListSpots")
	return uc.repo.ListAll(ctx)
}

// ReverseGeocode calls the Gaode (AMap) Web Service reverse geocode API to
// translate a GPS coordinate into a human-readable address. The API key is
// read from the GAODE_WEB_API_KEY environment variable — it must be a
// Web Service key (server-side) and is never hardcoded in source (T-13-03).
func (uc *MapUsecase) ReverseGeocode(ctx context.Context, lng float64, lat float64) (string, error) {
	uc.log.WithContext(ctx).Infof("ReverseGeocode: lng=%v lat=%v", lng, lat)

	apiKey := os.Getenv("GAODE_WEB_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GAODE_WEB_API_KEY not configured")
	}

	url := fmt.Sprintf("https://restapi.amap.com/v3/geocode/regeo?location=%f,%f&key=%s&extensions=base", lng, lat, apiKey)

	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Get(url)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("ReverseGeocode HTTP error: %v", err)
		return "", fmt.Errorf("reverse geocode request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		uc.log.WithContext(ctx).Errorf("ReverseGeocode read body error: %v", err)
		return "", fmt.Errorf("reverse geocode read body failed: %w", err)
	}

	// Gaode regeo response shape — only the fields we need are decoded.
	type regeocodePayload struct {
		FormattedAddress string `json:"formatted_address"`
	}
	type gaodeRegeoResponse struct {
		Status    string           `json:"status"`
		Info      string           `json:"info"`
		Regeocode regeocodePayload `json:"regeocode"`
	}

	var result gaodeRegeoResponse
	if err := json.Unmarshal(body, &result); err != nil {
		uc.log.WithContext(ctx).Errorf("ReverseGeocode parse error: %v", err)
		return "", fmt.Errorf("reverse geocode parse failed: %w", err)
	}

	// Gaode status "1" = success, anything else is an error. We log status only,
	// not the API key, to avoid leaking secret material (T-13-03).
	if result.Status != "1" {
		uc.log.WithContext(ctx).Errorf("ReverseGeocode Gaode API error: status=%s info=%s", result.Status, result.Info)
		return "", fmt.Errorf("gaode reverse geocode failed: status=%s info=%s", result.Status, result.Info)
	}

	return result.Regeocode.FormattedAddress, nil
}
