package gormigrator

import (
	"gorm.io/gorm"
)

// Migration model
type Migration struct {
	gorm.Model
	Code  string `gorm:"size:455"`
	Level string `gorm:"size:255"`
}

type MigrationStore struct {
	db *gorm.DB
}

func NewMigrationStore(db *gorm.DB) *MigrationStore {

	startMig := &Migration{}

	// Creating a migration table if not exists
	db.AutoMigrate(startMig)

	return &MigrationStore{
		db: db,
	}

}

// GetCurrentLevel returns the current level. In our case it is the filename, e.g. "mig0032.go"
func (m *MigrationStore) GetCurrentLevel() (string, string, error) {
	last := &Migration{}
	if err := m.db.Last(&last).Error; err != nil {
		return "", "", err
	}
	return last.Level, last.Code, nil
}

// func (m *MigrationStore) CountLevels() error {
// 	if err := m.db.Create(&currentState).Error; err != nil {
// 		return err
// 	}
// }

func (m *MigrationStore) SaveState(code, level string) error {
	currentState := &Migration{
		Code:  code,
		Level: level,
	}
	if err := m.db.Create(&currentState).Error; err != nil {
		return err
	}
	return nil
}
