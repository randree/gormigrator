package gormigrator

import (
	"gorm.io/gorm"
)

// Migration model
type Migration struct {
	gorm.Model
	Code  string `gorm:"size:455"`
	Level string `gorm:"size:255"`
	User  string `gorm:"size:255"`
}
