package data

import (
	"blog/internal/conf"
	"blog/internal/model"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo, NewPostRepo, NewUserRepo, NewDb, NewPriceRepo, NewDebtRepo, NewDebtDetailRepo, NewFuelVehicleRepo, NewRefuelRecordRepo)

// Data .
type Data struct {
	db *gorm.DB
}

func NewDb(c *conf.Data) *gorm.DB {
	db, err := gorm.Open(postgres.Open(c.Database.Source), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Info("打开数据库连接失败")
	}
	models := []interface{}{
		&Debt{},
		&DebtDetail{},
		&FuelVehicle{},
		&RefuelRecord{},
		&Post{},
		&Price{},
		&model.User{},
	}
	db.AutoMigrate(models...)
	return db
}

// NewData .
func NewData(c *conf.Data, db *gorm.DB, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{
		db: db,
	}, cleanup, nil
}
