package gormigrator

import (
	"gorm.io/gorm"
)

func init() {
	Code("user_table_start")
	Up(func(db *gorm.DB) error {

		type migtest struct {
			gorm.Model
			Name  string `gorm:"size:255"`
			Email string `gorm:"size:300"`
		}

		err := db.AutoMigrate(&migtest{})

		return err
	})

	Down(func(db *gorm.DB) error {

		err := db.Migrator().DropTable("migtests")

		return err
	})

}
