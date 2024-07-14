package congress

import "gorm.io/gorm"

type CongressCommitte struct {
	gorm.Model
}

func RegisterModels(db *gorm.DB) error {
	if err := db.AutoMigrate(&CongressCommitte{}); err != nil {
		return err
	}
	return nil
}
