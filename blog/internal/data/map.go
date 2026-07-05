package data

import (
	"context"
	"strconv"
	"strings"

	"blog/internal/biz"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

// Spot is the GORM model for a fishing spot. gps coordinates use double
// precision; tags is stored comma-separated (D-05); photos is a PostgreSQL
// text array of image URLs (D-06); address is the cached reverse geocode
// result (D-08).
type Spot struct {
	gorm.Model
	Name      string   `gorm:"type:varchar(255);not null"`
	Latitude  float64  `gorm:"type:double precision;not null"`
	Longitude float64  `gorm:"type:double precision;not null"`
	Notes     string   `gorm:"type:text"`
	Tags      string   `gorm:"type:text"`
	Photos    []string `gorm:"type:text[]"`
	Address   string   `gorm:"type:text"`
}

type mapRepo struct {
	data *Data
	log  *log.Helper
}

// NewMapRepo returns a biz.MapRepo backed by GORM.
func NewMapRepo(data *Data, logger log.Logger) biz.MapRepo {
	return &mapRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *mapRepo) Save(ctx context.Context, g *biz.Spot) (*biz.Spot, error) {
	var spot = Spot{
		Name:      g.Name,
		Latitude:  parseFloat(g.Latitude),
		Longitude: parseFloat(g.Longitude),
		Notes:     g.Notes,
		Tags:      g.Tags,
		Photos:    g.Photos,
		Address:   g.Address,
	}
	res := r.data.db.WithContext(ctx).Create(&spot)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected <= 0 {
		r.log.WithContext(ctx).Warn("mapRepo.Save: no rows affected")
	}
	g.Id = strconv.FormatUint(uint64(spot.ID), 10)
	g.CreatedAt = spot.CreatedAt.Format("2006-01-02 15:04:05")
	g.UpdatedAt = spot.UpdatedAt.Format("2006-01-02 15:04:05")
	return g, nil
}

func (r *mapRepo) Update(ctx context.Context, g *biz.Spot) (*biz.Spot, error) {
	id, _ := strconv.ParseUint(g.Id, 10, 64)
	err := r.data.db.WithContext(ctx).Model(&Spot{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":      g.Name,
		"latitude":  parseFloat(g.Latitude),
		"longitude": parseFloat(g.Longitude),
		"notes":     g.Notes,
		"tags":      g.Tags,
		"photos":    g.Photos,
		"address":   g.Address,
	}).Error
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (r *mapRepo) Delete(ctx context.Context, id int64) error {
	return r.data.db.WithContext(ctx).Delete(&Spot{}, id).Error
}

func (r *mapRepo) FindByID(ctx context.Context, id int64) (*biz.Spot, error) {
	var spot Spot
	err := r.data.db.WithContext(ctx).Where("id = ?", id).First(&spot).Error
	if err != nil {
		return nil, err
	}
	return spotModelToBiz(&spot), nil
}

func (r *mapRepo) ListAll(ctx context.Context) ([]*biz.Spot, error) {
	var spots []Spot
	err := r.data.db.WithContext(ctx).Find(&spots).Error
	if err != nil {
		return nil, err
	}
	result := make([]*biz.Spot, 0, len(spots))
	for i := range spots {
		result = append(result, spotModelToBiz(&spots[i]))
	}
	return result, nil
}

// spotModelToBiz maps a GORM Spot row to the biz domain struct. Latitude and
// Longitude are reformatted as strings to match the proto-friendly domain
// representation (mirrors the Post.go conventions where everything is string).
func spotModelToBiz(s *Spot) *biz.Spot {
	return &biz.Spot{
		Id:        strconv.FormatUint(uint64(s.ID), 10),
		Name:      s.Name,
		Latitude:  strconv.FormatFloat(s.Latitude, 'f', -1, 64),
		Longitude: strconv.FormatFloat(s.Longitude, 'f', -1, 64),
		Notes:     s.Notes,
		Tags:      s.Tags,
		Photos:    s.Photos,
		Address:   s.Address,
		CreatedAt: s.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: s.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// parseFloat tolerates empty strings by returning 0 — guards against malformed
// proto inputs while still letting GORM persist a non-null coordinate.
func parseFloat(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}
