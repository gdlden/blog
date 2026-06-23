package data

import (
	"encoding/json"
	"errors"

	"blog/internal/biz"

	"context"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
)

type AppVersion struct {
	gorm.Model
	Version    string `gorm:"column:version;type:varchar(32);not null;index;comment:版本号"`
	Info       string `gorm:"column:info;type:text;comment:更新说明(JSON数组)"`
	IosUrl     string `gorm:"column:ios_url;type:text;comment:iOS下载地址"`
	AndroidUrl string `gorm:"column:android_url;type:text;comment:Android下载地址"`
	IsActive   bool   `gorm:"column:is_active;default:false;comment:是否当前激活版本"`
}

func (AppVersion) TableName() string {
	return "app_versions"
}

type appVersionRepo struct {
	data *Data
	log  *log.Helper
}

func NewAppVersionRepo(data *Data, logger log.Logger) biz.AppVersionRepo {
	return &appVersionRepo{
		data: data,
		log:  log.NewHelper(logger),
	}
}

func (r *appVersionRepo) Save(ctx context.Context, av *biz.AppVersion) (uint, error) {
	infoJSON, err := json.Marshal(av.Info)
	if err != nil {
		return 0, err
	}
	entity := AppVersion{
		Version:    av.Version,
		Info:       string(infoJSON),
		IosUrl:     av.IosUrl,
		AndroidUrl: av.AndroidUrl,
		IsActive:   av.IsActive,
	}
	if av.IsActive {
		if err := r.data.db.WithContext(ctx).Model(&AppVersion{}).Where("is_active = ?", true).Update("is_active", false).Error; err != nil {
			return 0, err
		}
	}
	tx := r.data.db.WithContext(ctx).Create(&entity)
	if tx.Error != nil {
		return 0, tx.Error
	}
	return entity.ID, nil
}

func (r *appVersionRepo) Update(ctx context.Context, av *biz.AppVersion) error {
	infoJSON, err := json.Marshal(av.Info)
	if err != nil {
		return err
	}
	updates := map[string]any{
		"version":     av.Version,
		"info":        string(infoJSON),
		"ios_url":     av.IosUrl,
		"android_url": av.AndroidUrl,
	}
	if av.IsActive {
		if err := r.data.db.WithContext(ctx).Model(&AppVersion{}).Where("is_active = ?", true).Update("is_active", false).Error; err != nil {
			return err
		}
		updates["is_active"] = true
	} else {
		updates["is_active"] = false
	}
	tx := r.data.db.WithContext(ctx).Model(&AppVersion{}).Where("id = ?", av.Id).Updates(updates)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("no updatable app version found")
	}
	return nil
}

func (r *appVersionRepo) Delete(ctx context.Context, id uint) error {
	tx := r.data.db.WithContext(ctx).Delete(&AppVersion{}, id)
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return errors.New("no deletable app version found")
	}
	return nil
}

func (r *appVersionRepo) GetById(ctx context.Context, id uint) (*biz.AppVersion, error) {
	var entity AppVersion
	tx := r.data.db.WithContext(ctx).First(&entity, id)
	if tx.Error != nil {
		return nil, tx.Error
	}
	return mapAppVersionToBiz(&entity), nil
}

func (r *appVersionRepo) ListPage(ctx context.Context, req *biz.AppVersionPageRequest) ([]*biz.AppVersion, int64, error) {
	db := r.data.db.WithContext(ctx).Model(&AppVersion{})
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var entities []AppVersion
	offset := (req.Current - 1) * req.PageSize
	if err := db.Order("created_at DESC").Offset(offset).Limit(req.PageSize).Find(&entities).Error; err != nil {
		return nil, 0, err
	}
	items := make([]*biz.AppVersion, 0, len(entities))
	for i := range entities {
		items = append(items, mapAppVersionToBiz(&entities[i]))
	}
	return items, total, nil
}

func (r *appVersionRepo) GetActive(ctx context.Context) (*biz.AppVersion, error) {
	var entity AppVersion
	tx := r.data.db.WithContext(ctx).Where("is_active = ?", true).Last(&entity)
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			// Fallback: return the latest version if none is marked active
			var latest AppVersion
			tx2 := r.data.db.WithContext(ctx).Order("created_at DESC").First(&latest)
			if tx2.Error != nil {
				return nil, tx2.Error
			}
			return mapAppVersionToBiz(&latest), nil
		}
		return nil, tx.Error
	}
	return mapAppVersionToBiz(&entity), nil
}

func (r *appVersionRepo) SetActive(ctx context.Context, id uint) error {
	if err := r.data.db.WithContext(ctx).Model(&AppVersion{}).Where("is_active = ?", true).Update("is_active", false).Error; err != nil {
		return err
	}
	return r.data.db.WithContext(ctx).Model(&AppVersion{}).Where("id = ?", id).Update("is_active", true).Error
}

func (r *appVersionRepo) ClearAllActive(ctx context.Context) error {
	return r.data.db.WithContext(ctx).Model(&AppVersion{}).Where("is_active = ?", true).Update("is_active", false).Error
}

func mapAppVersionToBiz(entity *AppVersion) *biz.AppVersion {
	if entity == nil {
		return nil
	}
	var info []string
	if entity.Info != "" {
		if err := json.Unmarshal([]byte(entity.Info), &info); err != nil {
			info = nil
		}
	}
	return &biz.AppVersion{
		Id:         entity.ID,
		Version:    entity.Version,
		Info:       info,
		IosUrl:     entity.IosUrl,
		AndroidUrl: entity.AndroidUrl,
		IsActive:   entity.IsActive,
		CreatedAt:  entity.CreatedAt.Format(time.RFC3339),
		UpdatedAt:  entity.UpdatedAt.Format(time.RFC3339),
	}
}
