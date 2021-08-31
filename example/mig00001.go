package main

import (
	gm "github.com/randree/gormigrator/v1"
	"gorm.io/gorm"
)

func init() {

	gm.Mig(gm.State{

		Tag: "roles",

		Up: func(db *gorm.DB) error {

			type Role struct {
				ID   int `gorm:"primarykey"`
				Role string
			}

			db.AutoMigrate(&Role{})

			return db.Create(&Role{
				ID:   2,
				Role: "Guest",
			}).Error
		},

		Down: func(db *gorm.DB) error {
			err := db.Migrator().DropTable("roles")

			return err
		},
	})

}
