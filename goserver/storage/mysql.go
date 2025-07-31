package storage

import (
	"server/model"

	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var DB *gorm.DB

type InsertDataDao struct {
	DB *gorm.DB
}

// func init() {
// 	dsn := "root:12345678@tcp(127.0.0.1:3306)/your_database?charset=utf8mb4&parseTime=True&loc=Local"
// 	var err error
// 	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
// 	if err != nil {
// 		log.Fatal("数据库连接失败:", err)
// 	}
// 	err = DB.AutoMigrate(&model.ForwardIndex{})
// 	if err != nil {
// 		log.Fatal("迁移表失败:", err)
// 	}
// }

func InsertIndexData(in *model.ForwardIndex) (err error) {
	err = DB.Model(&model.ForwardIndex{}).Create(&in).Error
	if err != nil {
		return errors.Wrap(err, "failed to insert IndexData")
	}
	return
}
