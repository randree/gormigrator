package gormigrator

import (
	"gorm.io/gorm"
)

func init() {
	type migtest struct {
		// only new column in struct
		Testcol string
	}
	Code("user_table_add_column")
	Up(func(db *gorm.DB) error {

		err := db.Migrator().AddColumn(&migtest{}, "testcol")

		return err
	})

	Down(func(db *gorm.DB) error {
		err := db.Migrator().DropColumn(&migtest{}, "testcol")
		return err
	})

}
