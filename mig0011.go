package gormigrator

import (
	"errors"

	"gorm.io/gorm"
)

func init() {
	Code("user_table_through_error")
	Up(func(db *gorm.DB) error {

		err := errors.New("fake error")

		return err
	})

	Down(func(db *gorm.DB) error {

		return nil
	})

}
